package node

import (
	"encoding/json"
	"testing"
)

var nodes Node = Node{
	NodeProperty{ID: rootNode, Name: rootNode, Type: NonLeaf, MachineReg: "*"},
	[]*Node{
		{NodeProperty{ID: "0-1", Name: "0-1", Type: NonLeaf, MachineReg: "0-1"}, []*Node{}},
		{NodeProperty{ID: "0-2", Name: "0-2", Type: NonLeaf, MachineReg: "0-2"}, []*Node{
			{NodeProperty{ID: "0-2-1", Name: "0-2-1", Type: Leaf, MachineReg: "0-2-"}, []*Node{}},
			{NodeProperty{ID: "0-2-2", Name: "0-2-2", Type: NonLeaf, MachineReg: "0-2-2"}, []*Node{
				{NodeProperty{ID: "0-2-2-1", Name: "0-2-2-1", Type: Leaf, MachineReg: "0-2-2-1"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-2", Name: "0-2-2-2", Type: NonLeaf, MachineReg: "0-2-2-2"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-3", Name: "0-2-2-3", Type: NonLeaf, MachineReg: "0-2-2-3"}, []*Node{}},
				{NodeProperty{ID: "0-2-2-4", Name: "0-2-2-4", Type: NonLeaf, MachineReg: "0-2-2-4"}, []*Node{}},
			}},
		}},
		{NodeProperty{ID: "0-3", Name: "0-3", Type: NonLeaf, MachineReg: "0-3"}, []*Node{
			{NodeProperty{ID: "0-3-1", Name: "0-3-1", Type: NonLeaf, MachineReg: "0-3-1"}, []*Node{}},
			{NodeProperty{ID: "0-3-2", Name: "0-3-2", Type: NonLeaf, MachineReg: "0-3-2"}, []*Node{
				{NodeProperty{ID: "0-3-2-1", Name: "0-3-2-1", Type: Leaf, MachineReg: "0-3-2-1"}, []*Node{}},
			}},
		}},
		{NodeProperty{ID: "0-4", Name: "0-4", Type: Leaf, MachineReg: "0-4"}, []*Node{}},
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
var leafMachineReg map[string]string = map[string]string{
	"0-2-1.0-2." + rootNode: "0-2-", "0-2-2-1.0-2-2.0-2." + rootNode: "0-2-2-1",
	"0-3-2-1.0-3-2.0-3." + rootNode: "0-3-2-1", "0-4." + rootNode: "0-4"}

var resMap1 []map[string]string = []map[string]string{
	{"host": "127.0.0.1", "hostname": "127.0.0.1", "application": "loda"},
	{"host": "127.0.0.2", "hostname": "127.0.0.2", "application": "loda"}}
var resMap2 []map[string]string = []map[string]string{
	{"host": "127.0.0.2", "hostname": "127.0.0.2", "application": "loda"},
	{"host": "127.0.0.3", "hostname": "127.0.0.3", "application": "loda"}}

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

func checkStringInList(ori []string, dest string) bool {
	for _, item := range ori {
		if item == dest {
			return true
		}
	}
	return false
}

func TestNodeGetLeafChild(t *testing.T) {
	childNs, err := nodes.leafNs()
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

func TestLeafMachineReg(t *testing.T) {
	machineRegMap, err := nodes.leafMachineReg()
	if err != nil {
		t.Fatal("leafMachineReg error:", err.Error())
	}
	if len(machineRegMap) != 4 {
		t.Fatal("leafMachineReg not match with expect")
	}
	for ns, reg := range leafMachineReg {
		if machineRegMap[ns] != reg {
			t.Fatal("leafMachineReg not match with expect")
		}
	}
}

func TestUpdateNode(t *testing.T) {
	testNode := new(Node)
	*testNode = nodes
	testNode.update("newname", "*")
	if testNode.Name != "newname" || len(testNode.Children) != 4 {
		t.Fatalf("node update not match with expect: %v", testNode)
	}
}

func TestDeleteNode(t *testing.T) {
	var nodeTest Node = Node{
		NodeProperty{ID: rootNode, Name: rootNode, Type: NonLeaf, MachineReg: "*"},
		[]*Node{
			{NodeProperty{ID: "noChild", Name: "noChild", Type: NonLeaf, MachineReg: "-"}, []*Node{}},
			{NodeProperty{ID: "haveChild", Name: "haveChild", Type: NonLeaf, MachineReg: "-"}, []*Node{
				{NodeProperty{ID: "child", Name: "child", Type: Leaf, MachineReg: "-"}, []*Node{}},
			}},
		},
	}

	if err := nodeTest.delChild("noChild"); err != nil {
		t.Fatal("node delChild return false")
	}
	if nodeTest.Children[0].ID != "haveChild" {
		t.Fatalf("node after del children node not match with expect: %+v", nodeTest)
	}

	if err := nodeTest.delChild("haveChild"); err == nil {
		t.Fatal("node delChild success, not match with expect")
	}

	nodeParent, err := nodeTest.GetByNs("haveChild." + rootNode)
	if err != nil || nodeParent == nil {
		t.Fatalf("get node haveChild fail, error: %v", err)
	}
	if err := nodeParent.delChild("child"); err != nil {
		t.Fatalf("del node child return false, error:%s", err.Error())
	}
	if err := nodeTest.delChild("haveChild"); err != nil {
		t.Fatal("del node haveChild fail")
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
