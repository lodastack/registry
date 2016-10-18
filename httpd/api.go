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

	"github.com/julienschmidt/httprouter"
)

// Cluster is the interface op must implement.
type Cluster interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// Remove removes a node from the store, specified by addr.
	Remove(addr string) error

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

	router *httprouter.Router

	cluster Cluster

	logger *log.Logger
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
		router:  httprouter.New(),
		logger:  log.New("INFO", "http", model.LogBackend),
	}
}

// Start the server
func (s *Service) Start() error {
	s.initHandler()

	server := http.Server{
		Handler: s.router,
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

// FormRedirect returns the value for the "Location" header for a 301 response.
func (s *Service) FormRedirect(r *http.Request, host string) string {
	protocol := "http"
	// if s.credentialStore != nil {
	// 	protocol = "https"
	// }
	return fmt.Sprintf("%s://%s%s", protocol, host, r.URL.Path)
}

// all handlers just for test

func (s *Service) initHandler() {
	s.router.POST("/key", s.handlerKeySet)
	s.router.GET("/key", s.handlerKeyGet)

	s.router.POST("/batch", s.handlerBatch)

	s.router.POST("/bucket", s.handlerCreateBucket)
	s.router.DELETE("/bucket", s.handlerRemoveBucket)

	s.router.POST("/peer", s.handlerJoin)
	s.router.DELETE("/peer", s.handlerRemove)

	s.router.GET("/backup", s.handlerBackup)
}

func (s *Service) handlerKeySet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (s *Service) handlerKeyGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (s *Service) handlerBatch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (s *Service) handlerCreateBucket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")

	err := s.cluster.CreateBucket([]byte(name))
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerRemoveBucket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")

	err := s.cluster.RemoveBucket([]byte(name))
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerJoin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (s *Service) handlerRemove(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	if err := s.cluster.Remove(remoteAddr); err != nil {
		b := bytes.NewBufferString(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b.Bytes())
		return
	}
}

func (s *Service) handlerBackup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var data []byte
	if data, err = s.cluster.Backup(); err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}
