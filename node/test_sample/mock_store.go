package test_sample

import (
	"github.com/lodastack/log"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/store"
)

func MustNewStore(t *testing.T) *store.Store {
	path := mustTempDir()
	var err error
	// Ugly
	model.LogBackend, err = log.NewFileBackend(path)
	if err != nil {
		log.Errorf("new store error: create logger fail at %s, error: %s \n", path, err.Error())
	}
	s := store.New(path, mustMockTransport())
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func MustNewStoreB() *store.Store {
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
