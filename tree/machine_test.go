package tree

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/node"
	"github.com/lodastack/registry/tree/test_sample"
)

var testPath string = "./test_sample/"
var nodes node.Node
var resMap1, resMap2 []map[string]string

func init() {
	if err := test_sample.LoadJsonFromFile(testPath+"node.json", &nodes); err != nil {
		fmt.Println("load node.json fail:", err.Error())
	}
	if err := test_sample.LoadJsonFromFile(testPath+"resMap1.json", &resMap1); err != nil {
		fmt.Println("load resMap1.json fail:", err.Error())
	}
	if err := test_sample.LoadJsonFromFile(testPath+"resMap2.json", &resMap2); err != nil {
		fmt.Println("load resMap2.json fail:", err.Error())
	}
}

func TestMatchNs(t *testing.T) {
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

	// test example
	// ns:                       MachineReg
	// "0-2-1.0-2.loda"        : "0-2-"
	// "0-2-2-1.0-2-2.0-2.loda": "0-2-2-1"
	// "0-3-2-1.0-3-2.0-3.loda": "0-3-2-1"
	// "0-4.loda":               "0-4"
	tree.Nodes = &nodes
	if err := tree.saveTree(); err != nil {
		t.Fatal("savetree error")
	}

	// case 1: hostname "0-2-" only machine ns "0-2-1.0-2.loda"
	if nsList, err := tree.machine.MatchNs("0-2-"); err != nil || len(nsList) != 1 || nsList[0] != "0-2-1.0-2."+node.RootNode {
		t.Fatalf("match ns 0-2- not match with expect, error, %v, result %+v ", err, nsList)
	}

	// case 2: hostname "0-2-host" only machine ns "0-2-1.0-2.loda"
	if nsList, err := tree.machine.MatchNs("0-2-host"); err != nil || len(nsList) != 1 || nsList[0] != "0-2-1.0-2."+node.RootNode {
		t.Fatalf("match ns 0-2-host not match with expect, error, %v, result %+v ", err, nsList)
	}

	// case 3: hostname "0-2-2-1-host" machine two ns: "0-2-2-1.0-2-2.0-2.loda" and "0-2-1.0-2.loda"
	if nsList, err := tree.machine.MatchNs("0-2-2-1-host"); err != nil || len(nsList) != 2 {
		t.Fatalf("match ns 0-2-2-1-host not match with expect, error, %v, result %+v ", err, nsList)
	} else {
		if !common.CheckStringInList(nsList, "0-2-1.0-2."+node.RootNode) ||
			!common.CheckStringInList(nsList, "0-2-2-1.0-2-2.0-2."+node.RootNode) {
			t.Fatalf("match ns 0-2-2-1-host not match with expect, error, %v, result %+v ", err, nsList)
		}
	}

	// case 4: hostname not machine any ns, so get ns pool.
	if nsList, err := tree.machine.MatchNs("0-5-host"); err != nil || len(nsList) != 1 || nsList[0] != node.PoolNode+node.NodeDeli+node.RootNode {
		t.Fatalf("match ns 0-5-host match with expect, error, %v, result %+v ", err, nsList)
	}
}

func TestSearchMachine(t *testing.T) {
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

	_, err = tree.NewNode("test1", "comment1", node.RootNode, node.Leaf, "test1")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test2", "comment2", node.RootNode, node.Leaf, "test2")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	// 127.0.0.1 and 127.0.0.2
	resourceByte1, _ := model.NewResourceList(resMap1)
	// 127.0.0.2 and 127.0.0.3
	resourceByte2, _ := model.NewResourceList(resMap2)

	// test1.loda have 127.0.0.1 and 127.0.0.2
	err = tree.SetResource("test1."+node.RootNode, "machine", *resourceByte1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	// test2.loda have 127.0.0.2 and 127.0.0.3
	err = tree.SetResource("test2."+node.RootNode, "machine", *resourceByte2)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	// case 1: search not exist machine 127.0.0.0
	if result, err := tree.SearchMachine("127.0.0.0"); err != nil || len(result) != 0 {
		t.Fatalf("SearchMachine 127.0.0.0 fail, error: %s, result: %+v", err.Error(), result)
	}

	// case 2: search 127.0.0.1, exist in one ns
	if result, err := tree.SearchMachine("127.0.0.1"); err != nil {
		t.Fatal("SearchMachine 127.0.0.1 fail", err.Error())
	} else {
		if _, ok := result["test1."+node.RootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
	}

	// case 3: search 127.0.0.2, exist in two ns
	if result, err := tree.SearchMachine("127.0.0.2"); err != nil {
		t.Fatal("SearchMachine 127.0.0.2 fail", err.Error())
	} else {
		if _, ok := result["test1."+node.RootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
		if _, ok := result["test2."+node.RootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
	}
}

func TestUpdateStatusByHostname(t *testing.T) {
	s := test_sample.MustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test1", "comment1", node.RootNode, node.Leaf, "test1")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test2", "comment2", node.RootNode, node.Leaf, "test2")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	// 127.0.0.1 and 127.0.0.2
	resourceByte1, _ := model.NewResourceList(resMap1)
	// 127.0.0.2 and 127.0.0.3
	resourceByte2, _ := model.NewResourceList(resMap2)

	// test1.loda have 127.0.0.1 and 127.0.0.2
	err = tree.SetResource("test1."+node.RootNode, model.Machine, *resourceByte1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	// test2.loda have 127.0.0.2 and 127.0.0.3
	err = tree.SetResource("test2."+node.RootNode, model.Machine, *resourceByte2)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}

	if err := tree.UpdateStatusByHostname("127.0.0.1", map[string]string{model.HostStatusProp: "test"}); err != nil {
		t.Fatalf("UpdateStatusByHostname 127.0.0.1 fail: %s, ", err.Error())
	}
	if err := tree.UpdateStatusByHostname("127.0.0.2", map[string]string{model.HostStatusProp: "test"}); err != nil {
		t.Fatalf("UpdateStatusByHostname 127.0.0.2 fail: %s, ", err.Error())
	}
	if err := tree.UpdateStatusByHostname("127.0.0.3", map[string]string{model.HostStatusProp: "test"}); err != nil {
		t.Fatalf("UpdateStatusByHostname 127.0.0.3 fail: %s, ", err.Error())
	}

	if l, err := tree.resource.GetResourceList("test1."+node.RootNode, model.Machine); err != nil {
		t.Fatalf("read node test1 fail: %s ", err.Error())
	} else {
		for _, r := range *l {
			hostname, _ := r.ReadProperty(model.HostnameProp)
			status, _ := r.ReadProperty(model.HostStatusProp)
			if hostname != "127.0.0.1" && hostname != "127.0.0.2" {
				t.Fatalf("read node test1.loda machine not match as expect")
			}
			if status != "test" {
				t.Fatalf("read node test1.loda machine not match as expect")
			}
		}
	}

	if l, err := tree.resource.GetResourceList("test2."+node.RootNode, model.Machine); err != nil {
		t.Fatalf("read node test1 fail: %s ", err.Error())
	} else {
		for _, r := range *l {
			hostname, _ := r.ReadProperty(model.HostnameProp)
			status, _ := r.ReadProperty(model.HostStatusProp)
			if hostname != "127.0.0.2" && hostname != "127.0.0.3" {
				t.Fatalf("read node test1.loda machine not match as expect")
			}
			if status != "test" {
				t.Fatalf("read node test1.loda machine not match as expect")
			}
		}
	}

}

func TestRegisterMachine(t *testing.T) {
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
	_, err = tree.NewNode("test1", "comment1", node.RootNode, node.Leaf, "test1")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test2", "comment2", node.RootNode, node.Leaf, "test2")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test3", "comment3", node.RootNode, node.Leaf, "test-mu")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test4", "comment4", node.RootNode, node.Leaf, "test-multi")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	machine1 := model.NewResource(map[string]string{"ip": "10.10.10.1", "hostname": "test1-machine"})
	regMap, err := tree.RegisterMachine(machine1)
	if err != nil || len(regMap) != 1 {
		t.Fatalf("register machine case 1 not match with expect, regMap: %+v, error: %v", regMap, err)
	}
	for regNs, resID := range regMap {
		if regNs != "test1"+node.NodeDeli+node.RootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []string{"10.10.10.1"},
			Fuzzy: false,
		}
		if resMap, err := tree.SearchResource(regNs, "machine", searchIP); err != nil || len(resMap) != 1 {
			t.Fatal("cannot search machine1 after register")
		} else {
			rs := resMap[regNs]
			if rID, _ := (*rs)[0].ID(); rID != resID {
				t.Fatal("cannot search register resID")
			}
		}
	}

	machine2 := model.NewResource(map[string]string{"ip": "10.10.10.2", "hostname": "test2-machine"})
	regMap, err = tree.RegisterMachine(machine2)
	if err != nil || len(regMap) != 1 {
		t.Fatalf("register machine case 2 not match with expect, regMap: %+v, error: %v", regMap, err)
	}
	for regNs, resID := range regMap {
		if regNs != "test2"+node.NodeDeli+node.RootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []string{"10.10.10.2"},
			Fuzzy: false,
		}
		if resMap, err := tree.SearchResource(regNs, "machine", searchIP); err != nil || len(resMap) != 1 {
			t.Fatal("cannot search machine2 after register")
		} else {
			rs := resMap[regNs]
			if rID, _ := (*rs)[0].ID(); rID != resID {
				t.Fatal("cannot search register resID")
			}
		}
	}

	machine3 := model.NewResource(map[string]string{"ip": "10.10.10.3", "hostname": "test-multi-machine"})
	regMap, err = tree.RegisterMachine(machine3)
	if err != nil || len(regMap) != 2 {
		t.Fatalf("regist machine case 3 not match with expect, regMap: %+v, error: %v", regMap, err)
	}
	for regNs, resID := range regMap {
		if regNs != "test3"+node.NodeDeli+node.RootNode && regNs != "test4"+node.NodeDeli+node.RootNode {
			t.Fatal("resgier ns not match with expect:", regNs)
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []string{"10.10.10.3"},
			Fuzzy: false,
		}
		if resMap, err := tree.SearchResource(regNs, "machine", searchIP); err != nil || len(resMap) != 1 {
			t.Fatal("cannot search machine3 after register")
		} else {
			rs := resMap[regNs]
			if rID, _ := (*rs)[0].ID(); rID != resID {
				t.Fatal("cannot search register resID")
			}
		}
	}

	machine4 := model.NewResource(map[string]string{"ip": "10.10.10.4", "hostname": "no-machine"})
	regMap, err = tree.RegisterMachine(machine4)
	if err != nil || len(regMap) != 1 {
		t.Fatalf("regist machine case 4 not match with expect, regMap: %+v, error: %v", regMap, err)
	}
	for regNs, resID := range regMap {
		if regNs != node.PoolNode+node.NodeDeli+node.RootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []string{"10.10.10.4"},
			Fuzzy: false,
		}
		if resMap, err := tree.SearchResource(regNs, "machine", searchIP); err != nil || len(resMap) != 1 {
			t.Fatal("cannot search machine4 after register")
		} else {
			rs := resMap[regNs]
			if rID, _ := (*rs)[0].ID(); rID != resID {
				t.Fatal("cannot search register resID")
			}
		}
	}
}

func BenchmarkRegisterNewMachine(b *testing.B) {
	s := test_sample.MustNewStoreB(b)

	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		b.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	tree, err := NewTree(s)
	if err != nil {
		b.Fatal("NewTree error")
	}
	cnt := 100
	for i := 0; i < cnt; i++ {
		nodeName := fmt.Sprintf("test-%d", i)
		_, err = tree.NewNode(nodeName, "comment", node.RootNode, node.Leaf, nodeName)
		if err != nil {
			b.Fatalf("create leaf fail: %s", err.Error())
		}
	}
	b.ResetTimer()
	b.ReportAllocs()
	hostname := ""
	for i := 0; i < b.N; i++ {
		loop := i / 100
		num := i % 100
		hostname = fmt.Sprintf("test-%d-%d", num, loop)
		machine := model.NewResource(map[string]string{"ip": "10.10.10.3", "hostname": hostname})
		if regMap, err := tree.RegisterMachine(machine); err != nil || len(regMap) == 0 {
			b.Fatalf("register machine fail, error: %v", err)
		}
	}
}
