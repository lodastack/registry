package httpd

import (
	"sync"
)

// Session interface
type Session interface {
	// Get returns the session value associated to the given key.
	Get(key interface{}) interface{}
	// Set sets the session value associated to the given key.
	Set(key interface{}, val interface{})
	// Delete removes the session value associated to the given key.
	Delete(key interface{})
}

func NewSession() *LodaSession {
	m := make(map[interface{}]interface{})
	return &LodaSession{SessionMap: m}
}

// LodaSession struct
type LodaSession struct {
	//SessionMap store all client token
	SessionMap map[interface{}]interface{}
	// Mux locks SessionMap
	Mux sync.Mutex
}

// Get returns the session value associated to the given key.
func (ls *LodaSession) Get(i interface{}) interface{} {
	ls.Mux.Lock()
	defer ls.Mux.Unlock()
	return ls.SessionMap[i]
}

// Set sets the session value associated to the given key.
func (ls *LodaSession) Set(k, v interface{}) {
	ls.Mux.Lock()
	defer ls.Mux.Unlock()
	ls.SessionMap[k] = v
}

// Delete removes the session value associated to the given key.
func (ls *LodaSession) Delete(k interface{}) {
	ls.Mux.Lock()
	defer ls.Mux.Unlock()
	ls.SessionMap[k] = nil
}
