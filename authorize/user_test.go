package authorize

import (
	"os"
	"testing"
	"time"
)

func TestSetUser(t *testing.T) {
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

	// new User
	err = perm.SetUser("loda-manager", []string{"gid"}, []string{})
	if err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	// get User
	u, err := perm.GetUser("loda-manager")
	if err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if len(u.Groups) != 1 || u.Groups[0] != "gid" {
		t.Fatalf("GetUser resoult not match with expect, %v", u)
	}

	// update User
	err = perm.SetUser("loda-manager", []string{"gid1", "gid2"}, []string{})
	if err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	// get User
	u, err = perm.GetUser("loda-manager")
	if err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if len(u.Groups) != 2 || u.Groups[0] != "gid1" || u.Groups[1] != "gid2" {
		t.Fatalf("GetUser resoult not match with expect, %v", u)
	}
}

func TestRemoveUser(t *testing.T) {
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

	// set User
	err = perm.SetUser("loda-manager", []string{"gid"}, []string{})
	if err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}

	// get User
	u, err := perm.GetUser("loda-manager")
	if err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if len(u.Groups) != 1 || u.Groups[0] != "gid" {
		t.Fatalf("GetUser resoult not match with expect, %v", u)
	}

	// remove User
	err = perm.RemoveUser("loda-manager")
	if err != nil {
		t.Fatal("DeleteUser fail:", err.Error())
	}
	if u, err = perm.GetUser("loda-manager"); err == nil {
		t.Fatal("Get removed User success, not match with expect", err.Error())
	}
}
