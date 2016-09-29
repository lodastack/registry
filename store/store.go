// Package store provides a bolt distributed key-value store. The keys and
// associated values are changed via distributed consensus, meaning that the
// values are changed only when a majority of nodes in the cluster agree on
// the new value.
//
// Distributed consensus is provided via the Raft algorithm.
package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

var bucketNotFound = errors.New("bucket not found")
var ErrNotLeader = errors.New("not leader")

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
	leaderWaitDelay     = 100 * time.Millisecond

	boltFile = "registry.db"

	// cacheMaxMemorySize is the maximum size
	cacheMaxMemorySize = 1024 * 1024 * 50
)

type commandType int

const (
	update       commandType = iota // Commands which query the database.
	batch                           // Commands which modify the database.
	createBucket                    // Commands which create the bucket.
	removeBucket                    // Commands which remove the bucket.
)

// ClusterState defines the possible Raft states the current node can be in
type ClusterState int

// Represents the Raft cluster states
const (
	Leader ClusterState = iota
	Follower
	Candidate
	Shutdown
	Unknown
)

type command struct {
	Typ   commandType `json:"op,omitempty"`
	Name  []byte      `json:"name,omitempty"`  // bucket name for bucket management
	Batch []Row       `json:"batch,omitempty"` // for batch update
}

type Row struct {
	Key    []byte `json:"key,omitempty"`
	Value  []byte `json:"value,omitempty"`
	Bucket []byte `json:"bucket,omitempty"`
}

// Store is a bolt key-value store, where all changes are made via Raft consensus.
type Store struct {
	raftDir  string
	raftBind string
	dbPath   string

	mu sync.Mutex
	db *bolt.DB // The backend bolt store for the system.

	cache *Cache

	raft          *raft.Raft // The consensus mechanism
	peerStore     raft.PeerStore
	raftTransport *raft.NetworkTransport

	// TODO: need config from user?
	SnapshotThreshold uint64
	HeartbeatTimeout  time.Duration

	logger *log.Logger
}

// New returns a new Store.
func New(path string, listen string) *Store {
	return &Store{
		raftDir:  path,
		raftBind: listen,
		dbPath:   filepath.Join(path, boltFile),
		cache:    NewCache(cacheMaxMemorySize, nil),
		logger:   log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

// raftConfig returns a new Raft config for the store.
func (s *Store) raftConfig() *raft.Config {
	config := raft.DefaultConfig()
	if s.SnapshotThreshold != 0 {
		config.SnapshotThreshold = s.SnapshotThreshold
	}
	if s.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = s.HeartbeatTimeout
	}
	return config
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
func (s *Store) Open(enableSingle bool) error {

	if err := os.MkdirAll(s.raftDir, 0700); err != nil {
		return err
	}

	// Open backend storage
	db, err := bolt.Open(s.dbPath, 0600, nil)
	if err != nil {
		return err
	}
	s.db = db

	// Setup Raft configuration.
	config := raft.DefaultConfig()

	// Check for any existing peers.
	peers, err := readPeersJSON(filepath.Join(s.raftDir, "peers.json"))
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if enableSingle && len(peers) <= 1 {
		s.logger.Println("enabling single-node mode")
		config.EnableSingleNode = true
		config.DisableBootstrapAfterElect = false
	}

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", s.raftBind)
	if err != nil {
		return err
	}
	s.raftTransport, err = raft.NewTCPTransport(s.raftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create peer storage.
	s.peerStore = raft.NewJSONPeers(s.raftDir, s.raftTransport)

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.raftDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, logStore, snapshots, s.peerStore, s.raftTransport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	return nil
}

// Close closes the store. If wait is true, waits for a graceful shutdown.
func (s *Store) Close(wait bool) error {
	if err := s.db.Close(); err != nil {
		return err
	}
	f := s.raft.Shutdown()
	if wait {
		if e := f.(raft.Future); e.Error() != nil {
			return e.Error()
		}
	}
	return nil
}

// IsLeader is used to determine if the current node is cluster leader
func (s *Store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// Path returns the path to the store's storage directory.
func (s *Store) Path() string {
	return s.raftDir
}

// Leader returns the current leader. Returns a blank string if there is
// no leader.
func (s *Store) Leader() string {
	return s.raft.Leader()
}

// Nodes returns the list of current peers.
func (s *Store) Nodes() ([]string, error) {
	return s.peerStore.Peers()
}

// Addr returns the address of the store.
func (s *Store) Addr() string {
	return s.raftTransport.LocalAddr()
}

// State returns the current node's Raft state
func (s *Store) State() ClusterState {
	state := s.raft.State()
	switch state {
	case raft.Leader:
		return Leader
	case raft.Candidate:
		return Candidate
	case raft.Follower:
		return Follower
	case raft.Shutdown:
		return Shutdown
	default:
		return Unknown
	}
}

// WaitForLeader blocks until a leader is detected, or the timeout expires.
func (s *Store) WaitForLeader(timeout time.Duration) (string, error) {
	tck := time.NewTicker(leaderWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			l := s.Leader()
			if l != "" {
				return l, nil
			}
		case <-tmr.C:
			return "", fmt.Errorf("timeout expired")
		}
	}
}

// View returns the value for the given key.
func (s *Store) View(bucket, key []byte) ([]byte, error) {
	var value []byte
	if v, exist := s.cache.Get(bucket, key); exist {
		return v, nil
	}

	err := s.db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(bucket)
			if b == nil {
				return bucketNotFound
			}
			value = b.Get(key)
			return nil
		})
	// if the key not exist, bolt will return nil.
	if value != nil {
		s.cache.Add(bucket, key, value)
	}
	return value, err
}

// Update the value for the given key.
func (s *Store) Update(bucket []byte, key []byte, value []byte) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	rows := []Row{
		{
			Bucket: bucket,
			Key:    key,
			Value:  value,
		}}

	c := &command{
		Typ:   update,
		Batch: rows,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	r := f.Response().(*fsmGenericResponse)
	return r.error
}

// Batch update the values for the given keys.
func (s *Store) Batch(rows []Row) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	if len(rows) == 0 {
		return fmt.Errorf("no data in batch")
	}

	c := &command{
		Typ:   batch,
		Batch: rows,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	r := f.Response().(*fsmGenericResponse)
	return r.error
}

// CreateBucket create a bucket.
func (s *Store) CreateBucket(name []byte) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := &command{
		Typ:  createBucket,
		Name: name,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	r := f.Response().(*fsmGenericResponse)
	return r.error
}

// RemoveBucket remove a bucket.
func (s *Store) RemoveBucket(name []byte) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := &command{
		Typ:  removeBucket,
		Name: name,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	r := f.Response().(*fsmGenericResponse)
	return r.error
}

// Backup returns a snapshot of the store.
func (s *Store) Backup() ([]byte, error) {
	// TODO: not only leader can backup
	if s.raft.State() != raft.Leader {
		return nil, ErrNotLeader
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tmpFile, err := ioutil.TempFile("", "registry-backup-")
	if err != nil {
		return nil, err
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tx, err := s.db.Begin(true)
	if err != nil {
		return nil, err
	}

	if err := tx.CopyFile(tmpFile.Name(), 0600); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	var data []byte
	data, err = ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Join joins a node, located at addr, to this store. The node must be ready to
// respond to Raft communications at that address.
func (s *Store) Join(addr string) error {
	s.logger.Printf("received join request for remote node as %s", addr)

	f := s.raft.AddPeer(addr)
	if f.Error() != nil {
		return f.Error()
	}
	s.logger.Printf("node at %s joined successfully", addr)
	return nil
}

// Remove removes a node from the store, specified by addr.
func (s *Store) Remove(addr string) error {
	s.logger.Printf("received request to remove node %s", addr)

	f := s.raft.RemovePeer(addr)
	if f.Error() != nil {
		return f.Error()
	}
	s.logger.Printf("node %s removed successfully", addr)
	return nil
}

type fsm Store

type fsmGenericResponse struct {
	error error
}

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Typ {
	case update:
		err := f.applyUpdate(c.Batch)
		return &fsmGenericResponse{error: err}
	case batch:
		err := f.applyBatch(c.Batch)
		return &fsmGenericResponse{error: err}
	case createBucket:
		err := f.applyCreateBucket(c.Name)
		return &fsmGenericResponse{error: err}
	case removeBucket:
		err := f.applyRemoveBucket(c.Name)
		return &fsmGenericResponse{error: err}
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Typ))
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	snapFile, err := ioutil.TempFile("", "registry-snap-")
	if err != nil {
		return nil, err
	}
	snapFile.Close()
	defer os.Remove(snapFile.Name())

	tx, err := f.db.Begin(true)
	if err != nil {
		return nil, err
	}

	if err := tx.CopyFile(snapFile.Name(), 0600); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	fsm := &fsmSnapshot{}
	fsm.database, err = ioutil.ReadFile(snapFile.Name())
	if err != nil {
		log.Printf("Failed to read database for snapshot: %s", err.Error())
		return nil, err
	}

	return fsm, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	if err := f.db.Close(); err != nil {
		return err
	}

	// Get size of database.
	var sz int
	sz = binary.Size(rc)

	// Now read in the database file data and restore.
	database := make([]byte, sz)
	if _, err := io.ReadFull(rc, database); err != nil {
		return err
	}

	var db *bolt.DB
	var err error

	// Write snapshot over any existing database file.
	if err := ioutil.WriteFile(f.dbPath, database, 0660); err != nil {
		return err
	}

	// Re-open it.
	// Open backend storage
	db, err = bolt.Open(f.dbPath, 0600, nil)
	if err != nil {
		return err
	}

	f.db = db
	return nil
}

func (f *fsm) applyUpdate(rows []Row) error {
	if len(rows) != 1 {
		return fmt.Errorf("update just accept 1 row data: %d", len(rows))
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return f.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(rows[0].Bucket)
		if b == nil {
			return bucketNotFound
		}
		err := b.Put(rows[0].Key, rows[0].Value)

		// remove cache at last
		f.cache.Remove(rows[0].Bucket, rows[0].Key)
		return err
	})
}

func (f *fsm) applyBatch(rows []Row) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.db.Batch(func(tx *bolt.Tx) error {
		for _, row := range rows {
			b := tx.Bucket(row.Bucket)
			if b == nil {
				return bucketNotFound
			}
			if err := b.Put(row.Key, row.Value); err != nil {
				return err
			}
			// remove cache
			f.cache.Remove(row.Bucket, row.Key)
		}
		return nil
	})
}

func (f *fsm) applyCreateBucket(name []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// remove cache at first
	f.cache.RemoveBucket(name)

	return f.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(name)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (f *fsm) applyRemoveBucket(name []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(name)
		if err != nil {
			return fmt.Errorf("remove bucket: %s - %s", err, string(name))
		}
		// remove cache at last
		f.cache.RemoveBucket(name)
		return nil
	})
}

type fsmSnapshot struct {
	database []byte
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.database)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		if err := sink.Close(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f *fsmSnapshot) Release() {}

func readPeersJSON(path string) ([]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(b) == 0 {
		return nil, nil
	}

	var peers []string
	dec := json.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&peers); err != nil {
		return nil, err
	}

	return peers, nil
}
