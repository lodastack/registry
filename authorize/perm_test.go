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
