package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/lodastack/registry/model"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
)

// Cluster is the interface op must implement.
type Cluster interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// Create a bucket, via distributed consensus.
	CreateBucket(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// Backup database.
	Backup() ([]byte, error)
}

// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener
	// TODO: need fix, don't use classic martini, now just test
	m *martini.ClassicMartini

	cluster Cluster

	logger *log.Logger
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
		logger:  log.New(os.Stderr, "[http] ", log.LstdFlags),
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
	// batch update keys
	s.m.Post("/batch", s.handlerUpdate)
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

// NormalizeAddr ensures that the given URL has a HTTP protocol prefix.
// If none is supplied, it prefixes the URL with "http://".
func NormalizeAddr(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return fmt.Sprintf("http://%s", addr)
	}
	return addr
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
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

	if err := s.cluster.Join(remoteAddr); err != nil {
		b := bytes.NewBufferString(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b.Bytes())
		return
	}
}

// FormRedirect returns the value for the "Location" header for a 301 response.
func (s *Service) FormRedirect(r *http.Request, host string) string {
	protocol := "http"
	// if s.credentialStore != nil {
	// 	protocol = "https"
	// }
	return fmt.Sprintf("%s://%s%s", protocol, host, r.URL.Path)
}

// all handlers just for test

func (s *Service) handlerKeySet(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")
	bucket := r.FormValue("bucket")

	err := s.cluster.Update([]byte(bucket), []byte(key), []byte(value))
	if err != nil {
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
	if res, err = s.cluster.View([]byte(bucket), []byte(key)); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", string(res))
	}
}

func (s *Service) handlerUpdate(w http.ResponseWriter, r *http.Request) {
	var rows []model.Row
	rows = append(rows, model.Row{[]byte("k1"), []byte("v1"), []byte("bucket-test")})
	rows = append(rows, model.Row{[]byte("k2"), []byte("v2"), []byte("bucket-test-no")})
	if err := s.cluster.Batch(rows); err != nil {
		b := bytes.NewBufferString(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b.Bytes())
		return
	}
	fmt.Fprintf(w, "%s", "success")
}

func (s *Service) handlerCreateBucket(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	err := s.cluster.CreateBucket([]byte(name))
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerRemoveBucket(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	err := s.cluster.RemoveBucket([]byte(name))
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerBackup(w http.ResponseWriter, r *http.Request) {
	var err error
	var data []byte
	if data, err = s.cluster.Backup(); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}
