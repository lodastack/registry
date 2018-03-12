package test_sample

import (
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/lodastack/store/log"
	"github.com/lodastack/store/store"
)

func MustNewStore(t *testing.T) *store.Store {
	path := mustTempDir()

	s := store.New(path, mustMockTransport(), log.New())
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func MustNewStoreB(t *testing.B) *store.Store {
	path := mustTempDir()

	s := store.New(path, mustMockTransport(), log.New())
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

type MockTransport struct {
	ln net.Listener
}

func mustMockTransport() store.Transport {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic("failed to create new transport" + err.Error())
	}
	return &MockTransport{ln}
}

func (m *MockTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}

func (m *MockTransport) Accept() (net.Conn, error) { return m.ln.Accept() }

func (m *MockTransport) Close() error { return m.ln.Close() }

func (m *MockTransport) Addr() net.Addr { return m.ln.Addr() }
