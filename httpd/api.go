package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
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

	cluster Cluster

	logger *log.Logger
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
		logger:  log.New("INFO", "http", model.LogBackend),
	}
}

// Start the server
func (s *Service) Start() error {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.ln = ln

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			s.logger.Errorf("Serve error: ", err.Error())
		}
	}()
	s.logger.Println("service listening on: ", s.addr)

	return nil
}

// ServeHTTP allows Service to serve HTTP requests.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch {
	case strings.HasPrefix(r.URL.Path, "/key") && r.Method == http.MethodPost:
		s.handlerKeySet(w, r)
	case strings.HasPrefix(r.URL.Path, "/key") && r.Method == http.MethodGet:
		s.handlerKeyGet(w, r)
	case strings.HasPrefix(r.URL.Path, "/batch") && r.Method == http.MethodPost:
		s.handlerBatch(w, r)
	case strings.HasPrefix(r.URL.Path, "/bucket") && r.Method == http.MethodPost:
		s.handlerCreateBucket(w, r)
	case strings.HasPrefix(r.URL.Path, "/bucket") && r.Method == http.MethodDelete:
		s.handlerRemoveBucket(w, r)
	case strings.HasPrefix(r.URL.Path, "/join") && r.Method == http.MethodPost:
		s.handleJoin(w, r)
	case strings.HasPrefix(r.URL.Path, "/backup") && r.Method == http.MethodGet:
		s.handlerBackup(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// Close closes the service.
func (s *Service) Close() error {
	s.ln.Close()
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

func (s *Service) handlerBatch(w http.ResponseWriter, r *http.Request) {
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
