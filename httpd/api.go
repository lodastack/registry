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

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)

	// Restore restores backup data file.
	Restore(backupfile string) error
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
	s.router.POST("/api/v1/resource", s.handlerResourceSet)
	s.router.GET("/api/v1/resource", s.handlerResourceGet)
	s.router.GET("/api/v1/resource/search", s.handlerSearch)

	s.router.POST("/api/v1/ns", s.handlerNsNew)
	s.router.PUT("/api/v1/ns", s.handlerNsUpdate)
	s.router.GET("/api/v1/ns", s.handlerNsGet)
	s.router.DELETE("/api/v1/ns", s.handlerNsDel)

	s.router.POST("/api/v1/agent/ns", s.handlerRegister)

	s.router.POST("/api/v1/batch", s.handlerBatch)

	s.router.POST("/api/v1/bucket", s.handlerCreateBucket)
	s.router.DELETE("/api/v1/bucket", s.handlerRemoveBucket)

	s.router.POST("/api/v1/peer", s.handlerJoin)
	s.router.DELETE("/api/v1/peer", s.handlerRemove)

	s.router.GET("/api/v1/backup", s.handlerBackup)
	s.router.GET("/api/v1/restore", s.handlerRestore)
}

// Handle handlerRegister search hostname on the tree first,
// and register it if the machine not on the tree.
func (s *Service) handlerRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	machine := model.Resource{}
	if err := json.Unmarshal(buf.Bytes(), &machine); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	hostname, _ := machine.ReadProperty(node.HostnameProp)
	if hostname == "" {
		ReturnBadRequest(w, fmt.Errorf("invalid infomation"))
		return
	}

	if matchineMap, err := s.tree.SearchMachine(hostname); err != nil {
		s.logger.Errorf("SearchMachine fail, error: %s", err.Error())
		ReturnServerError(w, err)
		return
	} else if len(matchineMap) != 0 {
		ReturnJson(w, 200, matchineMap)
		return
	}

	regMap, err := s.tree.RegisterMachine(machine)
	if err != nil {
		s.logger.Errorf("RegisterMachine fail, error: %s", err.Error())
		ReturnServerError(w, err)
	} else {
		ReturnJson(w, 200, regMap)
	}
}

func (s *Service) handlerResourceSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryString := r.URL.Query()
	res := queryString.Get("resource")
	id := queryString.Get("nodeid")
	ns := queryString.Get("ns")

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	var err error
	if id != "" {
		err = s.tree.SetResourceByNodeID(id, res, buf.Bytes())
	} else if ns != "" {
		err = s.tree.SetResourceByNs(ns, res, buf.Bytes())
	} else {
		ReturnBadRequest(w, fmt.Errorf("invalid infomation"))
		return
	}

	if err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
	}
}

func (s *Service) handlerResourceGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var resource *model.Resources
	res := r.FormValue("resource")
	id := r.FormValue("nodeid")
	ns := r.FormValue("ns")

	if id != "" {
		resource, err = s.tree.GetResourceByNodeID(id, res)
	} else if ns != "" {
		resource, err = s.tree.GetResourceByNs(ns, res)
	} else {
		ReturnBadRequest(w, fmt.Errorf("invalid infomation"))
		return
	}
	if err != nil {
		ReturnServerError(w, err)
		return
	}
	if len(*resource) == 0 {
		ReturnNotFound(w, "No resources found.")
		return
	}
	ReturnJson(w, 200, resource)
}

func (s *Service) handlerNsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var nodes *node.Node
	var err error
	nodeid := r.FormValue("nodeid")
	// nodename := r.FormValue("nodename")

	if nodeid == "" {
		nodes, err = s.tree.AllNodes()
	} else {
		nodes, _, err = s.tree.GetNodeByID(nodeid)
	}
	if err != nil && err != node.ErrNodeNotFound {
		ReturnServerError(w, err)
		return
	}
	if nodes == nil {
		ReturnNotFound(w, "No node found.")
		return
	}

	ReturnJson(w, 200, nodes)
}

func (s *Service) handlerNsNew(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var id string
	queryString := r.URL.Query()
	name := queryString.Get("name")
	parent := queryString.Get("parent")
	nodeType := queryString.Get("type")
	machineMatch := queryString.Get("machinereg")

	nodeT, err := strconv.Atoi(nodeType)
	if name == "" || parent == "" || err != nil || (nodeT != node.Leaf && nodeT != node.NonLeaf) {
		ReturnServerError(w, fmt.Errorf("invalid information"))
		return
	}

	if id, err = s.tree.NewNode(name, parent, nodeT, machineMatch); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, id)
}

func (s *Service) handlerNsUpdate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	name := r.FormValue("name")
	machinereg := r.FormValue("machinereg")

	if err := s.tree.UpdateNode(ns, name, machinereg); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) handlerNsDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	parentNs := r.FormValue("parentns")
	delID := r.FormValue("delid")

	if err := s.tree.DelNode(parentNs, delID); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

// search bucket by nodes/key(resource)/resource_property
// TODO: return only or preperty ns or some property of resource from res.
func (s *Service) handlerSearch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	resource := r.FormValue("resource")
	k := r.FormValue("k")
	v := r.FormValue("v")
	searchType := r.FormValue("type")
	search := model.ResourceSearch{
		Key:   k,
		Value: []byte(v),
		Fuzzy: searchType == "fuzzy",
	}
	res, err := s.tree.SearchResourceByNs(ns, resource, search)
	if err != nil {
		s.logger.Errorf("handlerSearch SearchResourceByNs fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}

	if len(res) == 0 {
		ReturnNotFound(w, "No resources found.")
		return
	}
	ReturnJson(w, 200, res)
}

func (s *Service) handlerBatch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var rows []model.Row
	rows = append(rows, model.Row{[]byte("k1"), []byte("v1"), []byte("bucket-test")})
	rows = append(rows, model.Row{[]byte("k2"), []byte("v2"), []byte("bucket-test-no")})
	if err := s.cluster.Batch(rows); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) handlerCreateBucket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")

	err := s.cluster.CreateBucket([]byte(name))
	if err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
	}
}

func (s *Service) handlerRemoveBucket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")

	err := s.cluster.RemoveBucket([]byte(name))
	if err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
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
		ReturnBadRequest(w, fmt.Errorf("unmarshal fail"))
		return
	}

	if len(m) != 1 {
		ReturnBadRequest(w, fmt.Errorf("only allow 1 addr to join one time"))
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		ReturnBadRequest(w, fmt.Errorf("ihave no addr to join"))
		return
	}

	if err := s.cluster.Join(remoteAddr); err != nil {
		ReturnServerError(w, err)
		return
	}
}

func (s *Service) handlerRemove(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ReturnBadRequest(w, fmt.Errorf("read body fail"))
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		ReturnBadRequest(w, fmt.Errorf("unmarshal fail"))
		return
	}

	if len(m) != 1 {
		ReturnBadRequest(w, fmt.Errorf("only allow 1 addr to remove one time"))
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		ReturnBadRequest(w, fmt.Errorf("have no addr to join"))
		return
	}

	if err := s.cluster.Remove(remoteAddr); err != nil {
		ReturnServerError(w, err)
		return
	}
}

func (s *Service) handlerBackup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var data []byte
	if data, err = s.cluster.Backup(); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, string(data))
	}
}

func (s *Service) handlerRestore(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	file := r.FormValue("file")
	var err error
	if err = s.cluster.Restore(file); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
	}
}
