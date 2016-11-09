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

var nodes Node = Node{
	NodeProperty{ID: rootNode, Name: rootNode, Type: NonLeaf, MachineReg: "*"},
	[]*Node{
		{NodeProperty{ID: "0-1", Name: "0-1", Type: NonLeaf, MachineReg: "*"}, []*Node{}},
		{NodeProperty{ID: "0-2", Name: "0-2", Type: NonLeaf, MachineReg: "*"}, []*Node{
			{NodeProperty{ID: "0-2-1", Name: "0-2-1", Type: Leaf, MachineReg: "*"}, []*Node{}},
			{NodeProperty{ID: "0-2-2", Name: "0-2-2", Type: NonLeaf, MachineReg: "*"}, []*Node{
				{NodeProperty{ID: "0-2-2-1", Name: "0-2-2-1", Type: Leaf, MachineReg: "*"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-2", Name: "0-2-2-2", Type: NonLeaf, MachineReg: "*"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-3", Name: "0-2-2-3", Type: NonLeaf, MachineReg: "*"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-4", Name: "0-2-2-4", Type: NonLeaf, MachineReg: "*"}, []*Node{}},
			}},
		}},
		{NodeProperty{ID: "0-3", Name: "0-3", Type: NonLeaf, MachineReg: "*"}, []*Node{
			{NodeProperty{ID: "0-3-1", Name: "0-3-1", Type: NonLeaf, MachineReg: "*"}, []*Node{}},
			{NodeProperty{ID: "0-3-2", Name: "0-3-2", Type: NonLeaf, MachineReg: "*"}, []*Node{
				{NodeProperty{ID: "0-3-2-1", Name: "0-3-2-1", Type: Leaf, MachineReg: "*"}, []*Node{}},
			}},
		}},
		{NodeProperty{ID: "0-4", Name: "0-4", Type: Leaf, MachineReg: "*"}, []*Node{}},
	},
}

var nodeMap map[string]int = map[string]int{rootNode: 4, "0-1": 0,
	"0-2": 2, "0-2-1": 0, "0-2-2": 4, "0-2-2-1": 0, "0-2-2-2": 0, "0-2-2-3": 0, "0-2-2-4": 0,
	"0-3": 2, "0-3-1": 0, "0-3-2": 1, "0-3-2-1": 0,
	"0-4": 0}
var nodeNsMap map[string]int = map[string]int{"0-1." + rootNode: 0,
	"0-2." + rootNode: 2, "0-2-1.0-2." + rootNode: 0, "0-2-2.0-2." + rootNode: 4, "0-2-2-1.0-2-2.0-2." + rootNode: 0, "0-2-2-2.0-2-2.0-2." + rootNode: 0, "0-2-2-3.0-2-2.0-2." + rootNode: 0, "0-2-2-4.0-2-2.0-2." + rootNode: 0,
	"0-3." + rootNode: 2, "0-3-1.0-3." + rootNode: 0, "0-3-2.0-3." + rootNode: 1, "0-3-2-1.0-3-2.0-3." + rootNode: 0,
	"0-4." + rootNode: 0}

var resMap1 []map[string]string = []map[string]string{
	{"host": "127.0.0.1", "application": "loda"},
	{"host": "127.0.0.2", "application": "loda"}}
var resMap2 []map[string]string = []map[string]string{
	{"host": "127.0.0.2", "application": "loda"},
	{"host": "127.0.0.3", "application": "loda"}}

func getNodesByte() ([]byte, error) {
	return nodes.MarshalJSON()
}

func TestNodeMarshalJSON(t *testing.T) {
	if byteData, err := nodes.MarshalJSON(); err != nil || len(byteData) == 0 {
		t.Fatalf("nodes MarshalJSON fail")
	}
}

// Test get node by ns.
func TestGetByNs(t *testing.T) {
	if node, err := nodes.GetByNs(rootNode); err != nil || node == nil {
		t.Fatalf("nodes GetByNs \"root\" is valid, not match with expect\n")
	}
	if node, err := nodes.GetByNs("0-1." + rootNode); err != nil || node.ID != "0-1" {
		t.Fatalf("nodes GetByNs \"0-1.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get GetByNs \"0-1.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNs("0-2." + rootNode); err != nil || node.ID != "0-2" || len(node.Children) != 2 {
		t.Fatalf("nodes GetByNs \"0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get GetByNs \"0-2.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNs("0-2-1.0-2." + rootNode); err != nil || node.ID != "0-2-1" {
		t.Fatalf("nodes GetByNs \"0-2-1.0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get GetByNs \"0-2-1.0-2.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNs("0-2-2-2.0-2-2.0-2." + rootNode); err != nil || node.ID != "0-2-2-2" {
		t.Fatalf("nodes GetByNs \"0-2-2-2.0-2-2.0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get GetByNs \"0-2-2-2.0-2-2.0-2.root\" return right: %+v\n", node)
	}
}

// Test get node by ID.
func TestGetById(t *testing.T) {
	if node, _, err := nodes.GetByID(rootNode); err != nil || node == nil {
		t.Fatalf("nodes GetByID \"0\" is invalid, not match with expect\n")
	}
	if node, ns, err := nodes.GetByID("0-1"); err != nil || node.ID != "0-1" || ns != "0-1."+rootNode {
		t.Fatalf("nodes GetByID \"0-1.root\" not match with expect %+v, ns: %s,error: %s\n", node, ns, err)
	} else {
		t.Logf("get GetByID \"0-1.root\" return right: %+v, ns:%s\n", node, ns)
	}

	if node, ns, err := nodes.GetByID("0-2"); err != nil || node.ID != "0-2" || ns != "0-2."+rootNode || len(node.Children) != 2 {
		t.Fatalf("nodes GetByID \"0-2.root\" not match with expect %+v, ns: %s,error: %s\n", node, ns, err)
	} else {
		t.Logf("get GetByID \"0-2.root\" return right: %+v, ns:%s\n", node, ns)
	}

	if node, ns, err := nodes.GetByID("0-2-1"); err != nil || node.ID != "0-2-1" || ns != "0-2-1.0-2."+rootNode {
		t.Fatalf("nodes GetByID \"0-2-1.0-2.root\" not match with expect %+v, ns: %s,error: %s\n", node, ns, err)
	} else {
		t.Logf("get GetByID \"0-2-1.0-2.root\" return right: %+v, ns:%s\n", node, ns)
	}

	if node, ns, err := nodes.GetByID("0-2-2-2"); err != nil || node.ID != "0-2-2-2" || ns != "0-2-2-2.0-2-2.0-2."+rootNode {
		t.Fatalf("nodes GetByID \"0-2-2-2.0-2-2.0-2.root\" not match with expect %+v, ns: %s,error: %s\n", node, ns, err)
	} else {
		t.Logf("get GetByID \"0-2-2-2.0-2-2.0-2.root\" return right: %+v, ns:%s\n", node, ns)
	}
}

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

// Test node.UnmarshalJSON.
func TestNodeUnmarshalJSON(t *testing.T) {
	byteData, err := nodes.MarshalJSON()
	if err != nil || len(byteData) == 0 {
		t.Fatal("nodes MarshalJSON fail")
	}
	allNode := Node{}
	if err = allNode.UnmarshalJSON(byteData); err != nil {
		t.Fatalf("unmarshal error happen:%s\n", err.Error())
	}
	if ok := checkNodeForUnmarshal(allNode, nodeMap, t); !ok {
		t.Fatalf("unmarshal result not match witch expect")
	}
}

// check if nodes match expect or not.
func checkNodeForUnmarshal(allNode Node, nodeMap map[string]int, t *testing.T) bool {
	for id, childNum := range nodeMap {
		t.Log("GetByID", id, childNum)
		n, _, err := allNode.GetByID(id)
		if err != nil || n == nil || len(n.Children) != childNum {
			t.Log("unmarshal result not match witch expect")
			return false
		}
	}
	for name, childNum := range nodeNsMap {
		t.Log("GetByName", name, childNum)
		n, err := allNode.GetByNs(name)
		if err != nil || n == nil || len(n.Children) != childNum {
			t.Log("unmarshal result not match witch expect", name, childNum)
			return false
		}
	}
	return true
}

func TestSetResourceByID(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	resourceByte, _ := json.Marshal(resMap1)

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	// Set resource to leaf.
	leafID, err := tree.NewNode("test", rootID, Leaf)
	if err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNodeID(leafID, "machine", resourceByte)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	if res, err := tree.GetResourceByNodeID(leafID, "machine"); err != nil || len(*res) != 2 {
		t.Fatalf("get resource fail after set: %s\n", err.Error())
	} else {
		if (*res)[0]["host"] != "127.0.0.1" || (*res)[1]["host"] != "127.0.0.2" {
			t.Fatalf("resource not match with expect: %+v\n", *res)
		}
	}

	// Set resource to nonLeaf.
	nonLeafID, err := tree.NewNode("testNonLeaf", rootID, NonLeaf)
	if err != nil {
		t.Fatalf("create nonLeaf behind root fail: %s", err.Error())
	}
	if err = tree.SetResourceByNodeID(nonLeafID, "machine", resourceByte); err == nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
}

func TestSetResourceByNs(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	resourceByte, _ := json.Marshal(resMap1)

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	// Set resource to leaf.
	if _, err := tree.NewNode("test", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNs("test."+rootNode, "machine", resourceByte)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	if res, err := tree.GetResourceByNs("test."+rootNode, "machine"); err != nil || len(*res) != 2 {
		t.Fatalf("get resource fail after set: %s\n", err.Error())
	} else {
		if (*res)[0]["host"] != "127.0.0.1" || (*res)[1]["host"] != "127.0.0.2" {
			t.Fatalf("resource not match with expect: %+v\n", *res)
		}
	}

	// Set resource to nonLeaf.
	if _, err := tree.NewNode("testNonLeaf", rootID, NonLeaf); err != nil {
		t.Fatalf("create nonLeaf behind root fail: %s", err.Error())
	}
	if err = tree.SetResourceByNs("testNonLeaf."+rootNode, "machine", resourceByte); err == nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
}

func TestSearchResource(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	resourceByte1, _ := json.Marshal(resMap1)
	resourceByte2, _ := json.Marshal(resMap2)

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	// Set resource to leaf.
	if _, err := tree.NewNode("test1", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNs("test1."+rootNode, "machine", resourceByte1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	if _, err := tree.NewNode("test2", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNs("test2."+rootNode, "machine", resourceByte2)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	// search 127.0.0.1 show get 1 node each has one resource.
	search1_1 := model.ResourceSearch{
		Key:   "host",
		Value: []byte("127.0.0.1"),
		Fuzzy: false,
	}
	search1_2 := search1_1
	search1_2.Fuzzy = true
	res, err := tree.SearchResourceByNs(rootNode, "machine", search1_1)
	if resMachine, ok := res["test1."+rootNode]; err != nil || len(res) != 1 || !ok {
		t.Fatalf("search host 127.0.0.1 by not fuzzy type not match with expect")
	} else {
		if ip, ok := (*resMachine)[0].ReadProperty("host"); !ok || ip != "127.0.0.1" {
			t.Fatalf("search host 127.0.0.1 by not fuzzy type not match with expect")
		}
	}
	res, err = tree.SearchResourceByNs(rootNode, "machine", search1_2)
	if resMachine, ok := res["test1."+rootNode]; err != nil || len(res) != 1 || !ok {
		t.Fatalf("search host 127.0.0.1 by fuzzy type not match with expect")
	} else {
		if ip, ok := (*resMachine)[0].ReadProperty("host"); !ok || ip != "127.0.0.1" {
			t.Fatalf("search host 127.0.0.1 by fuzzy type not match with expect")
		}
	}

	// search 127.0.0.2 show get 2 node each has one resource.
	search2_1 := model.ResourceSearch{
		Key:   "host",
		Value: []byte("127.0.0.2"),
		Fuzzy: false,
	}
	search2_2 := search2_1
	search2_2.Fuzzy = true
	if res, err = tree.SearchResourceByNs(rootNode, "machine", search2_1); err != nil || len(res) != 2 {
		t.Fatalf("search host 127.0.0.2 by not fuzzy type not match with expect")
	}
	if res, err = tree.SearchResourceByNs(rootNode, "machine", search2_2); err != nil || len(res) != 2 {
		t.Fatalf("search host 127.0.0.2 by fuzzy type not match with expect")
	}

	// search 127.0.0. with not fuzzy type should get none result.
	search3_1 := model.ResourceSearch{
		Key:   "host",
		Value: []byte("127.0.0."),
		Fuzzy: false,
	}
	// search 127.0.0. with fuzzy type should get two node, and each has two resource.
	search3_2 := search3_1
	search3_2.Fuzzy = true
	if res, err = tree.SearchResourceByNs(rootNode, "machine", search3_1); err != nil || len(res) != 0 {
		t.Fatalf("search host 127.0.0. by not fuzzy type not match with expect")
	}
	if res, err = tree.SearchResourceByNs(rootNode, "machine", search3_2); len(res) != 2 {
		t.Fatalf("search host 127.0.0. by fuzzy type not match with expect")
	}
	for _, resMachine := range res {
		if len(*resMachine) != 2 {
			t.Fatalf("search host 127.0.0.3 by fuzzy type not match with expect")
		}
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

func TestGetResAfterSetOtherNs(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	resourceByte1, _ := json.Marshal(resMap1)
	resourceByte2, _ := json.Marshal(resMap2)

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	// Set resource to leaf.
	if _, err := tree.NewNode("leaf1", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNs("leaf1."+rootNode, "machine", resourceByte1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	if res, err := tree.GetResourceByNs("leaf1."+rootNode, "machine"); err != nil || len(*res) != 2 {
		t.Fatalf("get resource fail after set: %s\n", err.Error())
	} else {
		if (*res)[0]["host"] != "127.0.0.1" || (*res)[1]["host"] != "127.0.0.2" {
			t.Fatalf("resource not match with expect: %+v\n", *res)
		}
	}

	// Set resource to leaf.
	if _, err := tree.NewNode("leaf2", rootID, Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	err = tree.SetResourceByNs("leaf2."+rootNode, "machine", resourceByte2)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	if res, err := tree.GetResourceByNs("leaf2."+rootNode, "machine"); err != nil || len(*res) != 2 {
		t.Fatalf("get resource fail after set: %s\n", err.Error())
	} else {
		if (*res)[0]["host"] != "127.0.0.2" || (*res)[1]["host"] != "127.0.0.3" {
			t.Fatalf("resource not match with expect: %+v\n", *res)
		}
	}
}

func checkStringInList(ori []string, dest string) bool {
	for _, item := range ori {
		if item == dest {
			return true
		}
	}
	return false
}

func TestNodeGetLeafChild(t *testing.T) {
	childNs, err := nodes.getLeafNs()
	t.Log("result of GetLeafChild:", childNs)
	if err != nil || len(childNs) != 4 {
		t.Fatal("GetLeafChild not match with expect")
	}
	if !checkStringInList(childNs, "0-2-1.0-2.loda") ||
		!checkStringInList(childNs, "0-2-2-1.0-2-2.0-2.loda") ||
		!checkStringInList(childNs, "0-3-2-1.0-3-2.0-3.loda") ||
		!checkStringInList(childNs, "0-4.loda") {
		t.Fatal("GetLeafChild not match with expect")
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

func BenchmarkNodeJsonUnmarshal(b *testing.B) {
	var allNode Node
	nodeMapByte, err := getNodesByte()
	if err != nil {
		b.Fatal("getNodesByte error:%s\n", err.Error())
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := allNode.Unmarshal(nodeMapByte); err != nil {
			b.Fatalf("unmarshal error happen")
		}
	}
}

func (n *Node) Marshal() ([]byte, error) {
	return json.Marshal(*n)
}

func (n *Node) Unmarshal(v []byte) error {
	return json.Unmarshal(v, n)
}

func BenchmarkNodeFFJsonUnmarshal(b *testing.B) {
	var allNode Node
	nodeMapByte, err := getNodesByte()
	if err != nil {
		b.Fatal("getNodesByte error:%s\n", err.Error())
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// TODO: ffjson
		if err := allNode.UnmarshalJSON(nodeMapByte); err != nil {
			b.Fatalf("unmarshal error happen")
		}
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
