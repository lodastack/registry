package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/lodastack/registry/model"

	"github.com/hashicorp/raft"
)

var ErrNotLeader = raft.ErrNotLeader

var (
	TypPeer    = []byte("peer")
	TypCBucket = []byte("createrBucket")
	TypRBucket = []byte("removeBucket")
	TypUpdate  = []byte("update")
	TypBatch   = []byte("batch")
	TypJoin    = []byte("join")
	TypRemove  = []byte("remove")
)

type response struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SetPeer will set the mapping between raftAddr and apiAddr for the entire cluster.
func (s *Service) SetPeer(raftAddr, apiAddr string) error {
	// Try the local store. It might be the leader.
	err := s.store.UpdateAPIPeers(map[string]string{raftAddr: apiAddr})
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"api":  []byte(apiAddr),
		"raft": []byte(raftAddr),
	}

	msg["type"] = TypPeer
	return s.WriteLeader(msg)

}

// Join joins the node, reachable at addr, to the cluster.
func (s *Service) Join(addr string) error {
	// Try the local store. It might be the leader.
	err := s.store.Join(addr)
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"addr": []byte(addr),
	}

	msg["type"] = TypJoin
	return s.WriteLeader(msg)

}

// Remove removes a node from the store, specified by addr.
func (s *Service) Remove(addr string) error {
	// Try the local store. It might be the leader.
	err := s.store.Remove(addr)
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"addr": []byte(addr),
	}

	msg["type"] = TypRemove
	return s.WriteLeader(msg)

}

// CreateBucket will create bucket via the cluster.
func (s *Service) CreateBucket(name []byte) error {
	// Try the local store. It might be the leader.
	err := s.store.CreateBucket(name)
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"name": name,
	}

	msg["type"] = TypCBucket
	return s.WriteLeader(msg)
}

// RemoveBucket will remove bucket via the cluster.
func (s *Service) RemoveBucket(name []byte) error {
	// Try the local store. It might be the leader.
	err := s.store.RemoveBucket(name)
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"name": name,
	}

	msg["type"] = TypRBucket
	return s.WriteLeader(msg)
}

// Get returns the value for the given key.
func (s *Service) View(bucket, key []byte) ([]byte, error) {
	return s.store.View(bucket, key)
}

// TODO: Get buckets list by search ns.
func searchBucket(node []byte) []string {
	return []string{string(node)}
}

// Update will update the value of the given key in bucket via the cluster.
func (s *Service) Update(bucket []byte, key []byte, value []byte) error {
	// Try the local store. It might be the leader.
	err := s.store.Update(bucket, key, value)
	if err == nil || err != ErrNotLeader {
		return err
	}

	msg := map[string][]byte{
		"key":    key,
		"value":  value,
		"bucket": bucket,
	}

	msg["type"] = TypUpdate
	return s.WriteLeader(msg)
}

// Batch update values for given keys in given buckets, via distributed consensus.
func (s *Service) Batch(rows []model.Row) error {
	// Try the local store. It might be the leader.
	err := s.store.Batch(rows)
	if err == nil || err != ErrNotLeader {
		return err
	}

	// Don't use binary to encode?
	// https://github.com/golang/go/issues/478
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(rows); err != nil {
		return err
	}

	msg := map[string][]byte{
		"rows": buf.Bytes(),
	}

	msg["type"] = TypBatch
	return s.WriteLeader(msg)
}

// Backup database.
func (s *Service) Backup() ([]byte, error) {
	return s.store.Backup()
}

func (s *Service) WriteLeader(msg interface{}) error {
	// Try talking to the leader over the network.
	if leader := s.store.Leader(); leader == "" {
		return fmt.Errorf("no leader available")
	}
	conn, err := s.tn.Dial(s.store.Leader(), connectionTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := conn.Write(b); err != nil {
		return err
	}

	// Wait for the response and verify the operation went through.
	resp := response{}
	d := json.NewDecoder(conn)
	err = d.Decode(&resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func (s *Service) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	s.logger.Printf("received connection from %s", conn.RemoteAddr().String())

	// Only handles peers updates for now.
	msg := make(map[string][]byte)
	d := json.NewDecoder(conn)
	err := d.Decode(&msg)
	if err != nil {
		return
	}

	t, ok := msg["type"]
	if !ok {
		return
	}

	switch string(t) {
	case string(TypPeer):
		s.handleSetPeer(msg, conn)
	case string(TypCBucket):
		s.handleCreateBucket(msg, conn)
	case string(TypRBucket):
		s.handleRemoveBucket(msg, conn)
	case string(TypUpdate):
		s.handleUpdate(msg, conn)
	case string(TypBatch):
		s.handleBatch(msg, conn)
	case string(TypJoin):
		s.handleJoin(msg, conn)
	case string(TypRemove):
		s.handleRemove(msg, conn)
	default:
		s.logger.Errorf("unknown message type: %s", string(t))
		return
	}
}

func (s *Service) writeResponse(resp interface{}, conn net.Conn) {
	defer conn.Close()
	b, err := json.Marshal(resp)
	if err != nil {
		s.logger.Errorf("marshal resp error: %s", err.Error())
		return
	} else {
		if _, err := conn.Write(b); err != nil {
			s.logger.Errorf("write resp error: %s", err.Error())
			return
		}
	}
}

func (s *Service) handleSetPeer(msg map[string][]byte, conn net.Conn) {
	raftAddr, rok := msg["raft"]
	apiAddr, aok := msg["api"]
	if !rok || !aok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	// Update the peers.
	if err := s.store.UpdateAPIPeers(map[string]string{string(raftAddr): string(apiAddr)}); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}

func (s *Service) handleJoin(msg map[string][]byte, conn net.Conn) {
	addr, ok := msg["addr"]
	if !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	// Join the cluster.
	if err := s.store.Join(string(addr)); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}

func (s *Service) handleRemove(msg map[string][]byte, conn net.Conn) {
	addr, ok := msg["addr"]
	if !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	// Remove from the cluster.
	if err := s.store.Remove(string(addr)); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
}

func (s *Service) handleCreateBucket(msg map[string][]byte, conn net.Conn) {
	name, ok := msg["name"]
	if !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	if err := s.store.CreateBucket(name); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}

func (s *Service) handleRemoveBucket(msg map[string][]byte, conn net.Conn) {
	name, ok := msg["name"]
	if !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	if err := s.store.RemoveBucket(name); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}

func (s *Service) handleUpdate(msg map[string][]byte, conn net.Conn) {
	var bucket, key, value []byte
	var ok bool

	if bucket, ok = msg["bucket"]; !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	if key, ok = msg["key"]; !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	if value, ok = msg["value"]; !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	if err := s.store.Update(bucket, key, value); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}

func (s *Service) handleBatch(msg map[string][]byte, conn net.Conn) {
	var b []byte
	var rows []model.Row
	var ok bool

	if b, ok = msg["rows"]; !ok {
		resp := response{1, "need para"}
		s.writeResponse(resp, conn)
		return
	}

	reader := bytes.NewReader(b)
	d := json.NewDecoder(reader)
	err := d.Decode(&rows)
	if err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}

	if err := s.store.Batch(rows); err != nil {
		resp := response{1, err.Error()}
		s.writeResponse(resp, conn)
		return
	}
	s.writeResponse(response{}, conn)
	return
}
