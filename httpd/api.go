package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/node"

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

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

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
	tree    node.TreeMethod

	logger *log.Logger
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster, tree node.TreeMethod) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
		tree:    tree,
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
			s.logger.Fatalf("Serve error: %s\n", err.Error())
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
	s.router.POST("/resource", s.handlerResourceSet)
	s.router.GET("/resource", s.handlerResourceGet)

	s.router.POST("/ns", s.handlerNsNew)
	s.router.GET("/ns", s.handlerNsGet)

	s.router.POST("/batch", s.handlerBatch)

	s.router.POST("/bucket", s.handlerCreateBucket)
	s.router.DELETE("/bucket", s.handlerRemoveBucket)

	s.router.POST("/peer", s.handlerJoin)
	s.router.DELETE("/peer", s.handlerRemove)

	s.router.GET("/search", s.handlerSearch)

	s.router.GET("/backup", s.handlerBackup)
}

func (s *Service) handlerResourceSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryString := r.URL.Query()
	key := queryString.Get("key")
	id := queryString.Get("id")
	name := queryString.Get("name")

	buf := bytes.NewBufferString("")
	if _, err := buf.ReadFrom(r.Body); err != nil {
		fmt.Fprintf(w, "%s", "read body fail, please check and try again")
		return
	}

	var err error
	if id != "" {
		err = s.tree.SetResourceByNodeID(id, key, buf.Bytes())
	} else if name != "" {
		err = s.tree.SetResourceByNodeName(name, key, buf.Bytes())
	} else {
		err = fmt.Errorf("invalid node infomation to get resource")
	}

	if err != nil {
		fmt.Fprintf(w, "%s", err)
	} else {
		fmt.Fprintf(w, "%s", "success")
	}
}

func (s *Service) handlerResourceGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var res []byte
	var err error
	var resource *model.Resources
	key := r.FormValue("key")
	id := r.FormValue("id")
	name := r.FormValue("name")

	if id != "" {
		resource, err = s.tree.GetResourceByNodeID(id, key)
	} else if name != "" {
		resource, err = s.tree.GetResourceByNodeName(name, key)
	} else {
		err = fmt.Errorf("invalid node infomation to get resource")
	}
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}

	res, _ = json.Marshal(resource)
	fmt.Fprintf(w, "%s", string(res))
}

func (s *Service) handlerNsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var nodes *node.Node
	var err error
	var res []byte
	nodeid := r.FormValue("nodeid")
	// nodename := r.FormValue("nodename")

	if nodeid == "" {
		nodes, err = s.tree.GetAllNodes()
	} else {
		nodes, err = s.tree.GetNodeByID(nodeid)
	}
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}

	// TODO: ffjson
	res, _ = json.Marshal(nodes)
	fmt.Fprintf(w, "%s", string(res))
}

func (s *Service) handlerNsNew(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var id string
	queryString := r.URL.Query()
	name := queryString.Get("name")
	parent := queryString.Get("parent")
	nodeType := queryString.Get("type")

	nodeT, err := strconv.Atoi(nodeType)
	if name == "" || parent == "" || err != nil || (nodeT != node.Leaf && nodeT != node.NonLeaf) {
		fmt.Fprintf(w, "%s", "invalid param no create node, please check and try again")
		return
	}
	if id, err = s.tree.NewNode(name, parent, nodeT); err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}

	fmt.Fprintf(w, "%s", id)
}

// search bucket by nodes/key(resource)/resource_property
func (s *Service) handlerSearch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// ns := r.FormValue("ns")
	// key := r.FormValue("key")

	var result map[string]map[string]string = make(map[string]map[string]string, 0)
	out, _ := json.Marshal(result)
	fmt.Fprintf(w, "%s", string(out))
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
