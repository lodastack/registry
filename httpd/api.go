package httpd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/lodastack/registry/store"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
)

// Store is the interface Raft-backed key-value stores must implement.
type Store interface {
	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []store.Row) error

	// Create a bucket, via distributed consensus.
	CreateBucket(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

	// Backup database.
	Backup() ([]byte, error)

	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error
}

// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener
	// TODO: need fix, don't use classic martini, now just test
	m *martini.ClassicMartini

	store Store

	logger *log.Logger
}

// New returns an uninitialized HTTP service.
func New(addr string, store Store) *Service {
	return &Service{
		addr:   addr,
		store:  store,
		logger: log.New(os.Stderr, "[http] ", log.LstdFlags),
	}
}

// Start the server
func (s *Service) Start() error {
	s.m = martini.Classic()

	s.m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "DELETE", "GET"},
		AllowHeaders:     []string{"accept, content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	s.m.Post("/join", s.handleJoin)

	// get a key
	s.m.Get("/key", s.handlerKeyGet)
	// create key
	s.m.Post("/key", s.handlerKeySet)
	// create a bucket
	s.m.Post("/bucket", s.handlerCreateBucket)
	// remove a bucket
	s.m.Delete("/bucket", s.handlerRemoveBucket)

	// backup database
	s.m.Get("/backup", s.handlerBackup)

	go s.m.RunOnAddr(s.addr)

	return nil
}

// Close closes the service.
func (s *Service) Close() error {
	//s.ln.Close()
	return nil
}

// all handlers just for test

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) handlerKeySet(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")
	bucket := r.FormValue("bucket")

	if err := s.store.Update([]byte(bucket), []byte(key), []byte(value)); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerKeyGet(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	bucket := r.FormValue("bucket")

	var res []byte
	var err error
	if res, err = s.store.View([]byte(bucket), []byte(key)); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", string(res))
	}
}

func (s *Service) handlerCreateBucket(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	var err error
	if err = s.store.CreateBucket([]byte(name)); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerRemoveBucket(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	var err error
	if err = s.store.RemoveBucket([]byte(name)); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerBackup(w http.ResponseWriter, r *http.Request) {
	var err error
	var data []byte
	if data, err = s.store.Backup(); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}
