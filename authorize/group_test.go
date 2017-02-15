package authorize

import (
	"os"
	"testing"
	"time"

	"github.com/lodastack/registry/config"
)

func TestCreateGroup(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	config.C.Admins = []string{"loda-admin"}
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}

	// new Group
	if err = perm.CreateGroup("", []string{"manager"}, []string{"ns-resource-method"}); err == nil {
		t.Fatal("SetGroup success, not match with expect")
	}
	if err = perm.CreateGroup("test", []string{""}, []string{"ns-resource-method"}); err == nil {
		t.Fatal("SetGroup success, not match with expect")
	}
	if err = perm.CreateGroup("test", []string{"manager"}, []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	// get Group
	g, err := perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 1 || g.Manager[0] != "manager" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}

	if err = perm.CreateGroup("test", []string{"manager"}, []string{"ns-resource-method"}); err == nil {
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
	config.C.Admins = []string{"loda-admin"}
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}
	if err = perm.CreateGroup("test", []string{"manager"}, []string{"ns-resource-method"}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	err = perm.UpdateGroup("test", []string{}, []string{""})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}
	g, err := perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 1 ||
		g.Manager[0] != "manager" ||
		len(g.Items) != 1 ||
		g.Items[0] != "ns-resource-method" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}
	// update Group
	err = perm.UpdateGroup("test", []string{"manager1", "manager2"}, []string{"ns1-resource-method", "ns2-resource-method"})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}
	// get Group
	g, err = perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 2 ||
		g.Manager[0] != "manager1" ||
		g.Manager[1] != "manager2" ||
		len(g.Items) != 2 ||
		g.Items[0] != "ns1-resource-method" ||
		g.Items[1] != "ns2-resource-method" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}
}

func TestRemoveGroup(t *testing.T) {
	s := mustNewStore(t)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	config.C.Admins = []string{"loda-admin"}
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}

	// set group
	err = perm.CreateGroup("test", []string{"manager"}, []string{"ns-resource-method"})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}

	// get group
	g, err := perm.GetGroup("test")
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 1 || g.Manager[0] != "manager" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}

	// remove group
	err = perm.RemoveGroup("test")
	if err != nil {
		t.Fatal("DeleteGroup fail:", err.Error())
	}
	if g, err = perm.GetGroup("test"); err == nil {
		t.Fatal("Get removed Group success, not match with expect", err.Error())
	}
}
