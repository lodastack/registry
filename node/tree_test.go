package node

import (
	"encoding/json"
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
	if leafID, err = tree.NewNode("l1", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if err := tree.setResourceByNodeID(leafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to leafID fail: %s", err.Error())
	}
	// Test reate NonLeaf node and create bucket.
	if nonLeafID, err = tree.NewNode("n1", rootID, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}
	if err := tree.setResourceByNodeID(nonLeafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to nonLeafID fail: %s", err.Error())
	}
	// Test reate node under leaf node and create bucket.
	if _, err := tree.NewNode("n1", "1", NonLeaf); err == nil {
		t.Fatalf("create node under unexist root success, not match with expect")
	}
	if _, err := tree.NewNode("n1", leafID, NonLeaf); err == nil {
		t.Fatalf("create node under leaf success, not match with expect")
	}
	// Test reate node under nonleaf node and create bucket.
	if childID, err = tree.NewNode("n1", nonLeafID, NonLeaf); err != nil {
		t.Fatalf("create node behind nonLeaf node fail:%s\n", err.Error())
	}
	if err := tree.setResourceByNodeID(childID, "test", []byte("test")); err != nil {
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
	var leafID, nonLeafID string
	if leafID, err = tree.NewNode("testl", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if nonLeafID, err = tree.NewNode("testnl", rootID, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}

	if res, err := tree.GetResourceByNs(rootNode, template+"collect"); err != nil || len(*res) != 31 {
		t.Fatalf("get root collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if res, err := tree.GetResourceByNodeID(nonLeafID, template+"collect"); err != nil || len(*res) != 31 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	} else {
	}
	if res, err := tree.GetResourceByNodeID(leafID, "collect"); err != nil || len(*res) != 31 {
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

	resourceByte, _ := json.Marshal(resMap1)
	err = tree.SetResourceByNs(rootNode, template+"collect", resourceByte)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	var leafID, nonLeafID string
	if leafID, err = tree.NewNode("testl", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if nonLeafID, err = tree.NewNode("testnl", rootID, NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}
	if res, err := tree.GetResourceByNodeID(nonLeafID, template+"collect"); err != nil || len(*res) != 2 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if res, err := tree.GetResourceByNodeID(leafID, "collect"); err != nil || len(*res) != 2 {
		t.Fatalf("get LeafNode collect not match with expect, len: %d, err: %v\n", len(*res), err)
	}
}

func TestCreatePoolNode(t *testing.T) {
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
	if node, err := tree.GetNodeByNs(poolNode + nodeDeli + rootNode); err != nil || node.MachineReg != "^$" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", *node, err)
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

	childNs, err := tree.GetLeaf(rootNode, NsFormat)
	t.Log("result of NS GetLeafChild:", childNs)
	if err != nil || len(childNs) != 4 {
		t.Fatal("GetLeafChild not match with expect")
	}
	if !checkStringInList(childNs, "0-2-1.0-2.loda") ||
		!checkStringInList(childNs, "0-2-2-1.0-2-2.0-2.loda") ||
		!checkStringInList(childNs, "0-3-2-1.0-3-2.0-3.loda") ||
		!checkStringInList(childNs, "0-4.loda") {
		t.Fatal("GetLeafChild not match with expect")
	}

	childNs, err = tree.GetLeaf(rootNode, IDFormat)
	t.Log("result of ID GetLeafChild:", childNs)
	if err != nil || len(childNs) != 4 {
		t.Fatal("GetLeafChild not match with expect")
	}
	if !checkStringInList(childNs, "0-2-1") ||
		!checkStringInList(childNs, "0-2-2-1") ||
		!checkStringInList(childNs, "0-3-2-1") ||
		!checkStringInList(childNs, "0-4") {
		t.Fatal("GetLeafChild not match with expect")
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
