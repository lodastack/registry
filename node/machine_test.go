package node

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lodastack/registry/model"
)

func TestMatchNs(t *testing.T) {
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
	if nsList, err := tree.MatchNs("0-2-"); err != nil || len(nsList) != 1 || nsList[0] != "0-2-1.0-2."+rootNode {
		t.Fatalf("match ns 0-2- not match with expect, error, %v, result %+v ", err, nsList)
	}

	// case 2: hostname "0-2-host" only machine ns "0-2-1.0-2.loda"
	if nsList, err := tree.MatchNs("0-2-host"); err != nil || len(nsList) != 1 || nsList[0] != "0-2-1.0-2."+rootNode {
		t.Fatalf("match ns 0-2-host not match with expect, error, %v, result %+v ", err, nsList)
	}

	// case 3: hostname "0-2-2-1-host" machine two ns: "0-2-2-1.0-2-2.0-2.loda" and "0-2-1.0-2.loda"
	if nsList, err := tree.MatchNs("0-2-2-1-host"); err != nil || len(nsList) != 2 {
		t.Fatalf("match ns 0-2-2-1-host not match with expect, error, %v, result %+v ", err, nsList)
	} else {
		if !checkStringInList(nsList, "0-2-1.0-2."+rootNode) ||
			!checkStringInList(nsList, "0-2-2-1.0-2-2.0-2."+rootNode) {
			t.Fatalf("match ns 0-2-2-1-host not match with expect, error, %v, result %+v ", err, nsList)
		}
	}

	// case 4: hostname not machine any ns, so get ns pool.
	if nsList, err := tree.MatchNs("0-5-host"); err != nil || len(nsList) != 1 || nsList[0] != poolNode+nodeDeli+rootNode {
		t.Fatalf("match ns 0-5-host match with expect, error, %v, result %+v ", err, nsList)
	}
}

func TestSearchMachine(t *testing.T) {
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
	resourceByte1, _ := json.Marshal(resMap1)
	// 127.0.0.2 and 127.0.0.3
	resourceByte2, _ := json.Marshal(resMap2)

	// test1.loda have 127.0.0.1 and 127.0.0.2
	err = tree.SetResource("test1."+rootNode, "machine", resourceByte1)
	if err != nil {
		t.Fatalf("set resource fail: %s, not match with expect\n", err.Error())
	}
	// test2.loda have 127.0.0.2 and 127.0.0.3
	err = tree.SetResource("test2."+rootNode, "machine", resourceByte2)
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
		if _, ok := result["test1."+rootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
	}

	// case 3: search 127.0.0.2, exist in two ns
	if result, err := tree.SearchMachine("127.0.0.2"); err != nil {
		t.Fatal("SearchMachine 127.0.0.2 fail", err.Error())
	} else {
		if _, ok := result["test1."+rootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
		if _, ok := result["test2."+rootNode]; !ok {
			t.Fatalf("SearchMachine 127.0.0.1 not match with expect, result: %+v", result)
		}
	}
}

func TestRegisterMachine(t *testing.T) {
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
	_, err = tree.NewNode("test3", rootNode, Leaf, "test-mu")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}
	_, err = tree.NewNode("test4", rootNode, Leaf, "test-multi")
	if err != nil {
		t.Fatalf("create leaf fail: %s", err.Error())
	}

	machine1 := model.NewResource(map[string]string{"ip": "10.10.10.1", "hostname": "test1-machine"})
	regMap, err := tree.RegisterMachine(machine1)
	if err != nil || len(regMap) != 1 {
		t.Fatalf("register machine case 1 not match with expect, regMap: %+v, error: %v", regMap, err)
	}
	for regNs, resID := range regMap {
		if regNs != "test1"+nodeDeli+rootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []byte("10.10.10.1"),
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
		if regNs != "test2"+nodeDeli+rootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []byte("10.10.10.2"),
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
		if regNs != "test3"+nodeDeli+rootNode && regNs != "test4"+nodeDeli+rootNode {
			t.Fatal("resgier ns not match with expect:", regNs)
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []byte("10.10.10.3"),
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
		if regNs != poolNode+nodeDeli+rootNode {
			t.Fatal("resgier ns not match with expect")
		}
		searchIP := model.ResourceSearch{
			Key:   "ip",
			Value: []byte("10.10.10.4"),
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
	s := mustNewStoreB(b)

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
		_, err = tree.NewNode(nodeName, rootNode, Leaf, nodeName)
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
