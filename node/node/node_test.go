package node

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/node/test_sample"
)

var testPath string = "../test_sample/"
var nodes Node
var nodeMap, nodeNsMap map[string]int
var leafMachineReg map[string]string

func init() {
	if err := test_sample.LoadJsonFromFile(testPath+"node.json", &nodes); err != nil {
		fmt.Println("load node.json fail:", err.Error())
	}
	if err := test_sample.LoadJsonFromFile(testPath+"nodemap.json", &nodeMap); err != nil {
		fmt.Println("load nodemap.json fail:", err.Error())
	}
	if err := test_sample.LoadJsonFromFile(testPath+"nodeNsMap.json", &nodeNsMap); err != nil {
		fmt.Println("load nodeNsMap.json fail:", err.Error())
	}
	if err := test_sample.LoadJsonFromFile(testPath+"leafMachineReg.json", &leafMachineReg); err != nil {
		fmt.Println("load leafMachineReg.json fail:", err.Error())
	}
}

func getNodesByte() ([]byte, error) {
	return nodes.MarshalJSON()
}
func TestNodeMarshalJSON(t *testing.T) {
	if byteData, err := nodes.MarshalJSON(); err != nil || len(byteData) == 0 {
		t.Fatalf("nodes MarshalJSON fail")
	}
}

// Test get node by ns.
func TestGet(t *testing.T) {
	if node, err := nodes.GetByNS(rootNode); err != nil || node == nil {
		t.Fatalf("nodes Get \"root\" is valid, not match with expect\n")
	}
	if node, err := nodes.GetByNS("0-1." + rootNode); err != nil || node.ID != "0-1" {
		t.Fatalf("nodes Get \"0-1.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get Get \"0-1.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNS("0-2." + rootNode); err != nil || node.ID != "0-2" || len(node.Children) != 2 {
		t.Fatalf("nodes Get \"0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get GetByNs \"0-2.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNS("0-2-1.0-2." + rootNode); err != nil || node.ID != "0-2-1" {
		t.Fatalf("nodes Get \"0-2-1.0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get Get \"0-2-1.0-2.root\" return right: %+v\n", node)
	}
	if node, err := nodes.GetByNS("0-2-2-2.0-2-2.0-2." + rootNode); err != nil || node.ID != "0-2-2-2" {
		t.Fatalf("nodes Get \"0-2-2-2.0-2-2.0-2.root\" not match with expect %+v, error: %s\n", node, err)
	} else {
		t.Logf("get Get \"0-2-2-2.0-2-2.0-2.root\" return right: %+v\n", node)
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
		n, err := allNode.GetByNS(name)
		if err != nil || n == nil || len(n.Children) != childNum {
			t.Log("unmarshal result not match witch expect", name, childNum)
			return false
		}
	}
	return true
}

func TestNodeGetLeafChild(t *testing.T) {
	childNs, err := nodes.LeafNs()
	t.Log("result of GetLeafChild:", childNs)
	if err != nil || len(childNs) != 4 {
		t.Fatal("GetLeafChild not match with expect")
	}
	if !common.CheckStringInList(childNs, "0-2-1.0-2.loda") ||
		!common.CheckStringInList(childNs, "0-2-2-1.0-2-2.0-2.loda") ||
		!common.CheckStringInList(childNs, "0-3-2-1.0-3-2.0-3.loda") ||
		!common.CheckStringInList(childNs, "0-4.loda") {
		t.Fatal("GetLeafChild not match with expect")
	}
}

func TestLeafMachineReg(t *testing.T) {
	machineRegMap, err := nodes.LeafMachineReg()
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
	testNode.Update("newname", "*")
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

	if err := nodeTest.DelChild("noChild"); err != nil {
		t.Fatal("node delChild return false")
	}
	if nodeTest.Children[0].ID != "haveChild" {
		t.Fatalf("node after del children node not match with expect: %+v", nodeTest)
	}

	if err := nodeTest.DelChild("haveChild"); err == nil {
		t.Fatal("node delChild success, not match with expect")
	}

	nodeParent, err := nodeTest.GetByNS("haveChild." + rootNode)
	if err != nil || nodeParent == nil {
		t.Fatalf("get node haveChild fail, error: %v", err)
	}
	if err := nodeParent.DelChild("child"); err != nil {
		t.Fatalf("del node child return false, error:%s", err.Error())
	}
	if err := nodeTest.DelChild("haveChild"); err != nil {
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
