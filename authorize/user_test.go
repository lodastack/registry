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
	err = perm.SetUser("loda-manager", []string{})
	if err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	// get User
	user, err := perm.GetUser("loda-manager")
	if err != nil || len(user.Groups) != 1 || user.Groups[0] != "loda-defaultgroup" {
		t.Fatal("GetUser fail:", err.Error())
	}

}
