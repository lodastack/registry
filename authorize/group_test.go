package authorize

import (
	"os"
	"testing"
	"time"
)

func TestCreateGroup(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}

	// new Group
	if err = perm.CreateGroup("", []string{}, []string{}, []string{"ns-resource-method"}); err == nil {
		t.Fatal("SetGroup success, not match with expect")
	}
	if err = perm.CreateGroup("test", []string{}, []string{}, []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	// get Group
	_, err = perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}

	if err = perm.CreateGroup("test", []string{}, []string{}, []string{"ns-resource-method"}); err == nil {
		t.Fatal("SetGroup success not match expect")
	}
}

func TestUpdateGroup(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}
	if err = perm.CreateGroup("test", []string{}, []string{}, []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	if err = perm.UpdateItems("test", []string{""}); err == nil {
		t.Fatal("UpdateItems fail:", err.Error())
	}
	g, err := perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if g.Items[0] != "ns-resource-method" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}
	// update Group
	err = perm.UpdateItems("test", []string{"ns1-resource-method", "ns2-resource-method"})
	if err != nil {
		t.Fatal("UpdateItems fail:", err.Error())
	}
	// get Group
	g, err = perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Items) != 2 ||
		g.Items[0] != "ns1-resource-method" ||
		g.Items[1] != "ns2-resource-method" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}
}

func TestListGroup(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}

	// case 1
	nsList := []string{"server1.product1.com", "server2.product1.com"}
	for _, ns := range nsList {
		if err := perm.CreateGroup(GetGNameByNs(ns, "group1"), []string{}, []string{}, []string{"ns-resource-method"}); err != nil {
			t.Fatal("SetGroup fail:", err)
		}
	}
	if err := perm.CreateGroup(GetGNameByNs(nsList[0], "group2"), []string{}, []string{}, []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}
	if gList, err := perm.ListNsGroup("server1.product1.com"); err != nil || len(gList) != 2 {
		t.Fatalf("ListGroup not match with expect: %+v", gList)
	}
	if gList, err := perm.ListNsGroup("server2.product1.com"); err != nil || len(gList) != 1 {
		t.Fatalf("ListGroup not match with expect: %+v", gList)
	}

	// case 2
	nsList = []string{"server1.product2.com", "server11.product2.com", "server1-test.product2.com"}
	for _, ns := range nsList {
		if err := perm.CreateGroup(GetGNameByNs(ns, "group"), []string{}, []string{}, []string{"ns-resource-method"}); err != nil {
			t.Fatal("SetGroup fail:", err)
		}
	}
	if gList, err := perm.ListNsGroup("server1.product2.com"); err != nil || len(gList) != 1 {
		t.Fatalf("ListGroup not match with expect: %+v", gList)
	}
	if gList, err := perm.ListNsGroup("server1-test.product2.com"); err != nil || len(gList) != 1 {
		t.Fatalf("ListGroup not match with expect: %+v", gList)
	}
}
