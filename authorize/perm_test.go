package authorize

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/store"
)

// Test create
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

	_, err = perm.GetUser(defaultUser)
	if err != nil {
		t.Fatal("GetUser fail:", err.Error())
	}
	if defaultGid == "" {
		t.Fatal("defaultGid is invalid:")
	}
	g, err := perm.GetGroup(defaultGid)
	if err != nil {
		t.Fatal("GetGroup fail:", err.Error())
	} else if len(g.Items) != len(model.Templates) {
		t.Fatal("default Group items not match with expect, %+v:", g)
	}
}

func BenchmarkCheck(b *testing.B) {
	s := mustNewStoreB(b)
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		b.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	perm, err := NewPerm(s)
	if err != nil {
		b.Fatal("NewPerm fail:", err.Error())
	}

	var gids []string
	for i := 0; i < 1000; i++ {
		if gid, err := perm.SetGroup("", []string{"manager"}, []string{fmt.Sprintf("loda-ns-%d", i)}); err != nil {
			b.Fatal("SetGroup fail:", err.Error())
		} else {
			gids = append(gids, gid)
		}
	}
	for i := 0; i < 1000; i++ {
		if err := perm.SetUser(fmt.Sprintf("loda-manager-%d", i), []string{gids[i]}, []string{}); err != nil {
			b.Fatal("SetUser fail:", err.Error())
		}
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		perm.Check(fmt.Sprintf("loda-manager-%d", i), "ns", "resource", " method")
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
