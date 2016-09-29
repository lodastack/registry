package store

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"
)

const (
	node0  = "localhost:8300"
	node1  = "localhost:8301"
	bucket = "test-bucket"
	key    = "test-key"
	value  = "mdzz123"
)

func Test_IsLeader(t *testing.T) {
	s := mustNewStore(node0)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	if !s.IsLeader() {
		t.Fatalf("single node is not leader!")
	}
}

func Test_OpenCloseStore(t *testing.T) {
	s := mustNewStore(node0)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}

	if err := s.Close(true); err != nil {
		t.Fatalf("failed to close single-node store: %s", err.Error())
	}
}

func Test_SingleNode_CreateRemoveBucket(t *testing.T) {
	s := mustNewStore(node0)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	if err := s.CreateBucket([]byte(bucket)); err != nil {
		t.Fatalf("failed to create bucket: %s", err.Error())
	}

	if err := s.RemoveBucket([]byte(bucket)); err != nil {
		t.Fatalf("failed to remove bucket: %s", err.Error())
	}
}

func Test_SingleNode_SetGetKey(t *testing.T) {
	s := mustNewStore(node0)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	if err := s.CreateBucket([]byte(bucket)); err != nil {
		t.Fatalf("failed to create bucket: %s", err.Error())
	}

	if err := s.Update([]byte(bucket), []byte(key), []byte(value)); err != nil {
		t.Fatalf("failed to update key: %s", err.Error())
	}

	var v []byte
	var err error
	if v, err = s.View([]byte(bucket), []byte(key)); err != nil {
		t.Fatalf("failed to get key: %s", err.Error())
	}

	if string(v) != value {
		t.Fatalf("funexpected results for get: %s - %s ", string(v), value)
	}
}

func Test_MultiNode_SetGetKey(t *testing.T) {
	s0 := mustNewStore(node0)
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore(node1)
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	// Join the second node to the first.
	if err := s0.Join(s1.Addr()); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	if err := s0.CreateBucket([]byte(bucket)); err != nil {
		t.Fatalf("failed to create bucket: %s", err.Error())
	}

	if err := s0.Update([]byte(bucket), []byte(key), []byte(value)); err != nil {
		t.Fatalf("failed to update key: %s", err.Error())
	}

	time.Sleep(1 * time.Second)

	var v []byte
	var err error
	if v, err = s1.View([]byte(bucket), []byte(key)); err != nil {
		t.Fatalf("failed to get key: %s", err.Error())
	}

	if string(v) != value {
		t.Fatalf("funexpected results for get: %s - %s ", string(v), value)
	}
}

func Test_MultiNode_JoinRemove(t *testing.T) {
	s0 := mustNewStore(node0)
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore(node1)
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	// Get sorted list of cluster nodes.
	storeNodes := []string{s0.Addr(), s1.Addr()}
	sort.StringSlice(storeNodes).Sort()

	// Join the second node to the first.
	if err := s0.Join(s1.Addr()); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	nodes, err := s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes: %s", err.Error())
	}
	sort.StringSlice(nodes).Sort()

	if len(nodes) != len(storeNodes) {
		t.Fatalf("size of cluster is not correct")
	}
	if storeNodes[0] != nodes[0] && storeNodes[1] != nodes[1] {
		t.Fatalf("cluster does not have correct nodes")
	}

	// Remove a node.
	if err := s0.Remove(s1.Addr()); err != nil {
		t.Fatalf("failed to remove %s from cluster: %s", s1.Addr(), err.Error())
	}

	nodes, err = s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes post remove: %s", err.Error())
	}
	if len(nodes) != 1 {
		t.Fatalf("size of cluster is not correct post remove")
	}
	if s0.Addr() != nodes[0] {
		t.Fatalf("cluster does not have correct nodes post remove")
	}
}

func mustNewStore(addr string) *Store {
	path := mustTempDir()

	s := New(path, addr)
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func mustTempDir() string {
	var err error
	path, err := ioutil.TempDir("", "registry-test-")
	if err != nil {
		panic("failed to create temp dir")
	}
	return path
}
