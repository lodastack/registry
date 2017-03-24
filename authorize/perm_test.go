package authorize

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/store"
)

func TestUpdateGroupMember(t *testing.T) {
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

	if err = perm.SetUser("user1", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if _, err = perm.GetUser("user1"); err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if err = perm.SetUser("user2", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if _, err = perm.GetUser("user2"); err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if err = perm.CreateGroup("group1", []string{}, []string{}, []string{""}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}
	if _, err = perm.GetGroup("group1"); err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	}

	err = perm.UpdateMember("group1", []string{"user1"}, []string{}, Add)
	if err != nil {
		t.Fatal("TestUpdateGroupMember case1 fail", err)
	}
	user, err := perm.GetUser("user1")
	if err != nil || len(user.Groups) != 2 || user.Groups[1] != "group1" {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	group, err := perm.GetGroup("group1")
	if err != nil || len(group.Managers) != 1 || group.Managers[0] != "user1" || len(group.Members) != 0 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, group: %+v", err, group)
	}

	err = perm.UpdateMember("group1", []string{}, []string{"user2"}, Add)
	if err != nil {
		t.Fatal("TestUpdateGroupMember case2 fail", err)
	}
	user, err = perm.GetUser("user1")
	if err != nil || len(user.Groups) != 2 || user.Groups[1] != "group1" {
		t.Fatalf("TestUpdateGroupMember case2 fail, %v, user: %+v", err, user)
	}
	user, err = perm.GetUser("user2")
	if err != nil || len(user.Groups) != 2 || user.Groups[1] != "group1" {
		t.Fatalf("TestUpdateGroupMember case2 fail, %v, user: %+v", err, user)
	}
	group, err = perm.GetGroup("group1")
	if err != nil || len(group.Managers) != 1 || group.Managers[0] != "user1" || len(group.Members) != 1 || group.Members[0] != "user2" {
		t.Fatalf("TestUpdateGroupMember case2 fail, %v, group: %+v", err, group)
	}

	err = perm.UpdateMember("group1", []string{"user1"}, []string{"user2"}, Remove)
	if err != nil {
		t.Fatal("TestUpdateGroupMember case3 fail", err)
	}
	user, err = perm.GetUser("user1")
	if err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case3 fail, %v, user: %+v", err, user)
	}
	user, err = perm.GetUser("user2")
	if err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case3 fail, %v, user: %+v", err, user)
	}
	group, err = perm.GetGroup("group1")
	if err != nil || len(group.Managers) != 0 || len(group.Members) != 0 {
		t.Fatalf("TestUpdateGroupMember case3 fail, %v, group: %+v", err, group)
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

	// case 1
	if err = perm.SetUser("user1", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.SetUser("user2", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.SetUser("user3", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.CreateGroup("group1", []string{}, []string{}, []string{""}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}
	if err = perm.UpdateMember("group1", []string{"user1", "user3"}, []string{"user2", "user3"}, Add); err != nil {
		t.Fatal("TestUpdateGroupMember case3 fail", err)
	}

	// case 2
	if err := perm.RemoveUser("user1"); err != nil {
		t.Fatal("remove user case1 fail")
	}
	if _, err := perm.GetUser("user1"); err == nil {
		t.Fatalf("TestUpdateGroupMember case1 fail")
	}
	if user, err := perm.GetUser("user2"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	if group, err := perm.GetGroup("group1"); err != nil || len(group.Managers) != 1 || len(group.Members) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, group: %+v", err, group)
	}

	// case 3
	if err := perm.RemoveUser("user2"); err != nil {
		t.Fatal("remove user case2 fail")
	}
	if _, err := perm.GetUser("user2"); err == nil {
		t.Fatalf("TestUpdateGroupMember case2 fail")
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case2")
	}
	if group, err := perm.GetGroup("group1"); err != nil || len(group.Managers) != 1 || len(group.Members) != 1 {
		t.Fatalf("TestUpdateGroupMember case2 fail, %v, group: %+v", err, group)
	}

	if err := perm.RemoveUser("user3"); err != nil {
		t.Fatal("remove user case2 fail")
	}
	if _, err := perm.GetUser("user3"); err == nil {
		t.Fatalf("TestUpdateGroupMember case2")
	}
	if group, err := perm.GetGroup("group1"); err != nil || len(group.Managers) != 0 || len(group.Members) != 0 {
		t.Fatalf("TestUpdateGroupMember case2 fail, %v, group: %+v", err, group)
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

	if err = perm.SetUser("user1", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.SetUser("user2", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.SetUser("user3", ""); err != nil {
		t.Fatal("SetUser fail:", err.Error())
	}
	if err = perm.CreateGroup("group1", []string{}, []string{}, []string{}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}
	if err = perm.CreateGroup("group2", []string{}, []string{}, []string{}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}
	if err = perm.CreateGroup("group3", []string{}, []string{}, []string{}); err != nil {
		t.Fatal("SetGroup fail:", err)
	}

	if err = perm.UpdateMember("group1", []string{"user1", "user3"}, []string{"user2", "user3"}, Add); err != nil {
		t.Fatal("TestUpdateGroupMember case3 fail", err)
	}
	if err = perm.UpdateMember("group2", []string{"user2", "user3"}, []string{"user2", "user3"}, Add); err != nil {
		t.Fatal("TestUpdateGroupMember case3 fail", err)
	}
	if err = perm.UpdateMember("group3", []string{"user1"}, []string{}, Add); err != nil {
		t.Fatal("TestUpdateGroupMember case3 fail", err)
	}
	if user, err := perm.GetUser("user1"); err != nil || len(user.Groups) != 3 {
		t.Fatalf("TestUpdateGroupMember fail, %v, %+v", err, user)
	}
	if user, err := perm.GetUser("user2"); err != nil || len(user.Groups) != 3 {
		t.Fatalf("TestUpdateGroupMember fail, %v, user: %+v", err, user)
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 3 {
		t.Fatalf("TestUpdateGroupMember fail, %v, user: %+v", err, user)
	}

	// case 1
	if err := perm.RemoveGroup("group1"); err != nil {
		t.Fatal("remove group case 1 fail:", err)
	}
	if user, err := perm.GetUser("user1"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail")
	}
	if user, err := perm.GetUser("user2"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}

	// case 2
	if err := perm.RemoveGroup("group2"); err != nil {
		t.Fatal("remove group case 1 fail:", err)
	}
	if user, err := perm.GetUser("user1"); err != nil || len(user.Groups) != 2 {
		t.Fatalf("TestUpdateGroupMember case1 fail")
	}
	if user, err := perm.GetUser("user2"); err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}

	// case 3
	if err := perm.RemoveGroup("group3"); err != nil {
		t.Fatal("remove group case 1 fail:", err)
	}
	if user, err := perm.GetUser("user1"); err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case1 fail")
	}
	if user, err := perm.GetUser("user2"); err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
	if user, err := perm.GetUser("user3"); err != nil || len(user.Groups) != 1 {
		t.Fatalf("TestUpdateGroupMember case1 fail, %v, user: %+v", err, user)
	}
}

func TestNewPerm(t *testing.T) {
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

	_, err = perm.GetUser(DefaultUser)
	if err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}

	g, err := perm.GetGroup(defaultGName)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	} else if len(g.Items) != len(model.Templates)+3 {
		t.Fatal("default Group items not match with expect, %+v:", g)
	}
	g, err = perm.GetGroup(adminGName)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	} else if len(g.Items) != len(model.Templates)*4 {
		t.Fatal("default Group items not match with expect, %+v:", g)
	}
}

func mustNewStore(t *testing.T) *store.Store {
	path := mustTempDir()
	var err error
	// Ugly
	model.LogBackend, err = log.NewFileBackend(path)
	if err != nil {
		t.Fatalf("new store error: create logger fail at %s, error: %s \n", path, err.Error())
	}
	s := store.New(path, mustMockTransport())
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func mustNewStoreB(b *testing.B) *store.Store {
	path := mustTempDir()
	var err error
	// Ugly
	model.LogBackend, err = log.NewFileBackend(path)
	if err != nil {
		return nil
	}
	s := store.New(path, mustMockTransport())
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func mustTempDir() string {
	var err error
	path, err := ioutil.TempDir("", "registry-test-")
	if err != nil {
		panic("failed to create temp dir")
	}
	return path
}

type mockTransport struct {
	ln net.Listener
}

func mustMockTransport() store.Transport {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic("failed to create new transport" + err.Error())
	}
	return &mockTransport{ln}
}

func (m *mockTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}

func (m *mockTransport) Accept() (net.Conn, error) { return m.ln.Accept() }

func (m *mockTransport) Close() error { return m.ln.Close() }

func (m *mockTransport) Addr() net.Addr { return m.ln.Addr() }

func openPerm(t testing.T) (Perm, *store.Store) {
	s := mustNewStore(&t)
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}

	s.WaitForLeader(10 * time.Second)
	perm, err := NewPerm(s)
	if err != nil {
		t.Fatal("NewPerm fail:", err.Error())
	}
	return perm, s
}

func closePerm(s *store.Store) {
	s.Close(true)
	os.RemoveAll(s.Path())
}
