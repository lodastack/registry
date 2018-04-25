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
	err = perm.SetUser("loda-manager", "", "enable")
	if err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	// get User
	user, err := perm.GetUser("loda-manager")
	if err != nil || len(user.Groups) != 1 || user.Groups[0] != "loda-defaultgroup" || user.Alert != "enable" {
		t.Fatal("GetUser fail:", err.Error())
	}

}

func TestGetUserList(t *testing.T) {
	perm, s := openPerm(*t)
	defer closePerm(s)

	err1 := perm.SetUser("user1", "user1 mobile", "enable")
	err2 := perm.SetUser("user2", "user2 mobile", "enable")
	if err1 != nil || err2 != nil {
		t.Fatal("SetUser fail:", err1, err2)
	}

	// case1
	usernames := []string{"user1", "user2"}
	users, err := perm.GetUserList(usernames)
	if err != nil {
		t.Fatal("GetUserList case1 fail:", users, err)
	} else {
		if user1, ok := users["user1"]; !ok || user1.Mobile != "user1 mobile" {
			t.Fatalf("user1 not match with expect, %+v, %v", user1, err)
		}
		if user2, ok := users["user2"]; !ok || user2.Mobile != "user2 mobile" {
			t.Fatalf("user2 not match with expect, %+v, %v", user2, err)
		}
	}

	// case2
	usernames = []string{"user1", "user2", "user3"}
	users, err = perm.GetUserList(usernames)
	if err != nil {
		t.Fatal("GetUserList case1 fail:", users, err)
	} else {
		user1, ok1 := users["user1"]
		user2, ok2 := users["user2"]
		if !ok1 ||
			!ok2 ||
			user1.Mobile != "user1 mobile" ||
			user2.Mobile != "user2 mobile" {
			t.Fatalf("get user1/2 not match with expect, %+v,%+v, %v", user1, user2, err)
		}
		if _, ok := users["user3"]; ok {
			t.Fatalf("get user3 success, not match with expect")
		}
	}
}
