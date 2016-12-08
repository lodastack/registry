package authorize

import (
	"os"
	"testing"
	"time"
)

func TestSetGroup(t *testing.T) {
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
	gid, err := perm.SetGroup("", []string{"manager"}, []string{"ns-resource-method"})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}
	// get Group
	g, err := perm.GetGroup(gid)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 1 || g.Manager[0] != "manager" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}

	// update Group
	gid, err = perm.SetGroup(gid, []string{"manager1", "manager2"}, []string{"ns1-resource-method", "ns2-resource-method"})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}
	// get Group
	g, err = perm.GetGroup(gid)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 2 ||
		g.Manager[0] != "manager1" ||
		g.Manager[1] != "manager2" ||
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
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}

	// set group
	gid, err := perm.SetGroup("", []string{"manager"}, []string{"ns-resource-method"})
	if err != nil {
		t.Fatal("SetGroup fail:", err.Error())
	}

	// get group
	g, err := perm.GetGroup(gid)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}
	if len(g.Manager) != 1 || g.Manager[0] != "manager" {
		t.Fatalf("GetGroup resoult not match with expect, %v", g)
	}

	// remove group
	err = perm.RemoveGroup(gid)
	if err != nil {
		t.Fatal("DeleteGroup fail:", err.Error())
	}
	if g, err = perm.GetGroup(gid); err == nil {
		t.Fatal("Get removed Group success, not match with expect", err.Error())
	}
}
