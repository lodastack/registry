package tree

import (
	"os"
	"sort"
	"testing"
	"time"

	"github.com/lodastack/models"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
	n "github.com/lodastack/registry/tree/node"
	"github.com/lodastack/registry/tree/test_sample"
)

// Test create
func TestCreateNodeAndLeafCache(t *testing.T) {
	s := test_sample.MustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)

	var leafID, nonLeafID, childNonID, childLeafID string
	// Test reate Leaf node and create bucket.
	if leafID, err = tree.NewNode("l1", n.RootNode, n.Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if err := tree.setByteToStore(leafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to leafID fail: %s", err.Error())
	}
	// Test reate NonLeaf node and create bucket.
	if nonLeafID, err = tree.NewNode("n1", n.RootNode, n.NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}
	if err := tree.setByteToStore(nonLeafID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to nonLeafID fail: %s", err.Error())
	}
	// Test reate node under leaf node and create bucket.
	if _, err := tree.NewNode("n1", "1", n.NonLeaf); err == nil {
		t.Fatalf("create node under unexist root success, not match with expect")
	}
	if _, err := tree.NewNode("n1", "l1."+n.RootNode, n.NonLeaf); err == nil {
		t.Fatalf("create node under leaf success, not match with expect")
	}
	// Test reate node under nonleaf node and create bucket.
	if childNonID, err = tree.NewNode("nn1", "n1."+n.RootNode, n.NonLeaf); err != nil {
		t.Fatalf("create node behind nonLeaf node fail: %s\n", err.Error())
	}
	if childLeafID, err = tree.NewNode("nnl1", "nn1.n1."+n.RootNode, n.Leaf); err != nil {
		t.Fatalf("create node under nonleaf fail: %s", err.Error())
	}
	if err := tree.setByteToStore(childNonID, "test", []byte("test")); err != nil {
		t.Fatalf("set k-v to childID fail: %s", err.Error())
	}
	poolnode, err := tree.GetNodeByNS("pool.loda")

	if err != nil {
		t.Fatalf("get pool node fail: %s", err.Error())
	}
	leafIDs, err := tree.LeafChildIDs(n.RootNode)
	if err != nil || len(leafIDs) != 3 {
		t.Fatalf("get leaf of root fail not match with expect:%s,%s %v", leafID, childLeafID, leafIDs)
	}
	sort.Strings(leafIDs)
	expectedIDs := []string{leafID, childLeafID, poolnode.ID}
	sort.Strings(expectedIDs)
	for i := 0; i < len(leafIDs); i++ {
		if leafIDs[i] != expectedIDs[i] {
			t.Fatalf("leaf id not match with expect")
		}
	}
}

func TestCopyTemplateDuringCreateNode(t *testing.T) {
	s := test_sample.MustNewStore(t)
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

	if _, err = tree.NewNode("testl", n.RootNode, n.Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if _, err = tree.NewNode("testnl", n.RootNode, n.NonLeaf); err != nil {
		t.Fatalf("create nonleaf behind root fail: %s", err.Error())
	}

	if res, err := tree.GetResourceList(n.RootNode, template+"collect"); err != nil || len(*res) != model.TemplateCollectNum {
		t.Fatalf("get root collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if res, err := tree.GetResourceList("testnl.loda", template+"collect"); err != nil || len(*res) != model.TemplateCollectNum {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if alarms, err := tree.GetResourceList("testnl.loda", template+model.Alarm); err != nil || len(*alarms) != 1 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*alarms), err)
	} else {
		for _, alarm := range *alarms {
			if alarm["db"] != "" {
				t.Fatalf("get nonLeafNode alarm_template not match with expect, db: %s \n", alarm["db"])
			}
		}
	}

	if res, err := tree.GetResourceList("testl.loda", "collect"); err != nil || len(*res) != model.TemplateCollectNum {
		t.Fatalf("get LeafNode collect not match with expect, len: %d, err: %v\n", len(*res), err)
	}
	if alarms, err := tree.GetResourceList("testl.loda", model.Alarm); err != nil || len(*alarms) != 1 {
		t.Fatalf("get nonLeafNode collect_template not match with expect, len: %d, err: %v\n", len(*alarms), err)
	} else {
		for _, alarm := range *alarms {
			if alarm["db"] != models.DBPrefix+"testl.loda" {
				t.Fatalf("get nonLeafNode alarm_template not match with expect, db: %s \n", alarm["db"])
			}
			if alarm["groups"] != "loda.testl-op" {
				t.Fatalf("get nonLeafNode alarm_template not match with expect, groups: %s \n", alarm["groups"])
			}
		}
	}
}

func TestUpdateTemplate(t *testing.T) {
	s := test_sample.MustNewStore(t)
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
	err = tree.SetResource(n.RootNode, template+"collect", *resource1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	if _, err = tree.NewNode("testl", n.RootNode, n.Leaf); err != nil {
		t.Fatalf("create leaf behind root fail: %s", err.Error())
	}
	if _, err = tree.NewNode("testnl", n.RootNode, n.NonLeaf); err != nil {
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
	s := test_sample.MustNewStore(t)
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
	if node, err := tree.GetNodeByNS(n.PoolNode + n.NodeDeli + n.RootNode); err != nil || node.MachineReg != "^$" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}
}

func TestTreeGetLeaf(t *testing.T) {
	s := test_sample.MustNewStore(t)
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

	childIDs, err := tree.LeafChildIDs(n.RootNode)
	t.Log("result of ID LeafIDs:", childIDs)
	if err != nil || len(childIDs) != 4 {
		t.Fatalf("LeafIDs not match with expect, leaf: %+v, error: %v", childIDs, err)
	}
	if !common.CheckStringInList(childIDs, "0-2-1") ||
		!common.CheckStringInList(childIDs, "0-2-2-1") ||
		!common.CheckStringInList(childIDs, "0-3-2-1") ||
		!common.CheckStringInList(childIDs, "0-4") {
		t.Fatal("GetLeafChild not match with expect")
	}
}

func TestTreeUpdateNode(t *testing.T) {
	s := test_sample.MustNewStore(t)
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

	// case 1 : update leaf node name and machineReg.
	if err := tree.UpdateNode("0-4.loda", "0-5", "test update"); err != nil {
		t.Fatalf("tree UpdateNode error: %s", err.Error())
	}
	if node, err := tree.GetNodeByNS("0-5.loda"); err != nil || node.MachineReg != "test update" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}

	// case 2 : update leaf machineReg.
	if err := tree.UpdateNode("0-5.loda", "0-5", "test update-2"); err != nil {
		t.Fatalf("tree UpdateNode error: %s", err.Error())
	}
	if node, err := tree.GetNodeByNS("0-5.loda"); err != nil || node.MachineReg != "test update-2" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}

	// case 3: update nonleaf node name.
	if err := tree.UpdateNode("0-3.loda", "0-6", "test update"); err != nil {
		t.Fatalf("tree UpdateNode error: %s", err.Error())
	}
	if node, err := tree.GetNodeByNS("0-6.loda"); err != nil || node.MachineReg != "test update" {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}
	if node, err := tree.GetNodeByNS("0-3-1.0-6.loda"); err != nil {
		t.Fatalf("root pool node not match with expect, node: %+v, error: %v", node, err)
	}

	// case 4: update node name to a already exist node.
	if err := tree.UpdateNode("0-2-1.0-2.loda", "0-2-2", "test update"); err == nil {
		t.Fatal("tree UpdateNode 0-2-1.0-2.loda success, not match with expect")
	}
}

func TestRomoveNode(t *testing.T) {
	s := test_sample.MustNewStore(t)
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

	_, err = tree.NewNode("test1", n.RootNode, n.Leaf, "test1")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test2", n.RootNode, n.Leaf, "test2")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	// 127.0.0.1 and 127.0.0.2
	resource1, _ := model.NewResourceList(resMap1)

	// test1.loda have 127.0.0.1 and 127.0.0.2
	err = tree.SetResource("test1."+n.RootNode, "machine", *resource1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	if err := tree.RemoveNode("test1.loda"); err == nil {
		t.Fatal("delete ns still have machine success, not match wich expect")
	}
	if err := tree.RemoveNode("test2.loda"); err != nil {
		t.Fatalf("delete ns have no machine fail, not match wich expect, error: %s", err.Error())
	}
}
