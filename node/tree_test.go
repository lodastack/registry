package node

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/store"
)

// Test create
func TestCreateNode(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	var leafID, nonLeafID, childID string
	// Test reate Leaf node and create bucket.
	if leafID, err = tree.NewNode("l1", rootNode, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if err := tree.setByteToStore(leafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to leafID fail: %s", err.Error())
	}
	// Test reate NonLeaf node and create bucket.
	if nonLeafID, err = tree.NewNode("n1", rootNode, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}
	if err := tree.setByteToStore(nonLeafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to nonLeafID fail: %s", err.Error())
	}
	// Test reate node under leaf node and create bucket.
	if _, err := tree.NewNode("n1", "1", NonLeaf); err == nil {
		t.Fatalf("create node under unexist root success, not match with expect")
	}
	if _, err := tree.NewNode("n1", "l1."+rootNode, NonLeaf); err == nil {
		t.Fatalf("create node under leaf success, not match with expect")
	}
	// Test reate node under nonleaf node and create bucket.
	if childID, err = tree.NewNode("n1", "n1."+rootNode, NonLeaf); err != nil {
		t.Fatalf("create node behind nonLeaf node fail: %s\n", err.Error())
	}
	if err := tree.setByteToStore(childID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to childID fail: %s", err.Error())
	}
}

func TestCopyTemplateDuringCreateNode(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatalf("NewTree fail: %s\n", err.Error())
	}

	if _, err = tree.NewNode("testl", rootNode, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if _, err = tree.NewNode("testnl", rootNode, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}

	if res, err := tree.GetResourceList(rootNode, template+"collect"); err != nil || len(*res) != 31 {
		t.Fatalf("get root collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if res, err := tree.GetResourceList("testnl.loda", template+"collect"); err != nil || len(*res) != 31 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	} else {
	}
	if res, err := tree.GetResourceList("testl.loda", "collect"); err != nil || len(*res) != 31 {
		t.Fatalf("get LeafNode collect not match with expect, len: %d, err: %v\n", len(*res), err)
	}
}

func TestUpdateTemplate(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatalf("NewTree fail: %s\n", err.Error())
	}

	resource1, _ := model.NewResourceList(resMap1)
	err = tree.SetResource(rootNode, template+"collect", *resource1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	if _, err = tree.NewNode("testl", rootNode, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if _, err = tree.NewNode("testnl", rootNode, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}
	if res, err := tree.GetResourceList("testnl.loda", template+"collect"); err != nil || len(*res) != 2 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if res, err := tree.GetResourceList("testl.loda", "collect"); err != nil || len(*res) != 2 {
		t.Fatalf("get LeafNode collect not match with expect, len: %d, err: %v\n", len(*res), err)
	}
}

func TestInitPoolNode(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatal("newtree fail")
	}

	// Test root pool node.
	if node, err := tree.GetNode(poolNode + nodeDeli + rootNode); err != nil || node.MachineReg != "^$" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}
}

func TestTreeGetLeaf(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatal("NewTree error")
	}

	tree.Nodes = &nodes
	if err := tree.saveTree(); err != nil {
		t.Fatal("saveTree error")
	}
	allNodes, err := tree.AllNodes()
	if err != nil {
		t.Fatal("AllNodes fail")
	}
	if tree.Cache, err = allNodes.initNsCache(); err != nil {
		t.Fatal("initNsCache fail")
	}
	childIDs, err := tree.LeafIDs(rootNode)
	t.Log("result of ID LeafIDs:", childIDs)
	if err != nil || len(childIDs) != 4 {
		t.Fatalf("LeafIDs not match with expect, leaf: %+v, error: %v", childIDs, err)
	}
	if !checkStringInList(childIDs, "0-2-1") ||
		!checkStringInList(childIDs, "0-2-2-1") ||
		!checkStringInList(childIDs, "0-3-2-1") ||
		!checkStringInList(childIDs, "0-4") {
		t.Fatal("GetLeafChild not match with expect")
	}
}

func TestTreeUpdateNode(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatal("NewTree error")
	}
	tree.Nodes = &nodes
	if err := tree.saveTree(); err != nil {
		t.Fatal("saveTree error")
	}

	if err := tree.UpdateNode("0-4.loda", "0-5", "test update"); err != nil {
		t.Fatalf("tree UpdateNode error: %s", err.Error())
	}
	if node, err := tree.GetNode("0-5.loda"); err != nil || node.MachineReg != "test update" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}

	if err := tree.UpdateNode("0-3.loda", "0-6", "test update"); err != nil {
		t.Fatalf("tree UpdateNode error: %s", err.Error())
	}
	if node, err := tree.GetNode("0-6.loda"); err != nil || node.MachineReg != "test update" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}
	if node, err := tree.GetNode("0-3-1.0-6.loda"); err != nil {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}

	if err := tree.UpdateNode("0-2-1.0-2.loda", "0-2-2", "test update"); err == nil {
		t.Fatal("tree UpdateNode 0-2-1.0-2.loda success, not match with expect")
	}
}

func TestDelNode(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatal("NewTree error")
	}

	_, err = tree.NewNode("test1", rootNode, Leaf, "test1")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test2", rootNode, Leaf, "test2")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	// 127.0.0.1 and 127.0.0.2
	resource1, _ := model.NewResourceList(resMap1)

	// test1.loda have 127.0.0.1 and 127.0.0.2
	err = tree.SetResource("test1."+rootNode, "machine", *resource1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	if err := tree.DelNode("test1.loda"); err == nil {
		t.Fatal("delete ns still have machine success, not match wich expect")
	}
	if err := tree.DelNode("test2.loda"); err != nil {
		t.Fatalf("delete ns have no machine fail, not match wich expect, error: %s", err.Error())
	}
}

func mustNewStore(t *testing.T) *store.Store {
	path := mustTempDir()
	var err error
	// Ugly
	model.LogBackend, err = log.NewFileBackend(path)
	if err != nil {
		t.Fatalf("new store error: create logger fail at %s, error: %s \n", path, err.Error())
	}
	s := store.New(path, mustMockTransport())
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func mustNewStoreB(b *testing.B) *store.Store {
	path := mustTempDir()
	var err error
	// Ugly
	model.LogBackend, err = log.NewFileBackend(path)
	if err != nil {
		return nil
	}
	s := store.New(path, mustMockTransport())
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

type mockTransport struct {
	ln net.Listener
}

func mustMockTransport() store.Transport {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic("failed to create new transport" + err.Error())
	}
	return &mockTransport{ln}
}

func (m *mockTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}

func (m *mockTransport) Accept() (net.Conn, error) { return m.ln.Accept() }

func (m *mockTransport) Close() error { return m.ln.Close() }

func (m *mockTransport) Addr() net.Addr { return m.ln.Addr() }
