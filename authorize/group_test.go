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
	if err = perm.CreateGroup("", []string{"ns-resource-method"}); err == nil {
		t.Fatal("SetGroup success, not match with expect")
	}
	if err = perm.CreateGroup("test", []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	// get Group
	_, err = perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}

	if err = perm.CreateGroup("test", []string{"ns-resource-method"}); err == nil {
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
	if err = perm.CreateGroup("test", []string{"ns-resource-method"}); err != nil {
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
