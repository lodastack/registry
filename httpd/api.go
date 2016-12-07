package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lodastack/log"
	m "github.com/lodastack/models"
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

	// Backup database.
	Backup() ([]byte, error)

	// Restore restores backup data file.
	Restore(backupfile string) error
}

// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener

	router  *httprouter.Router
	session *LodaSession

	cluster Cluster
	tree    node.TreeMethod

	logger *log.Logger
}

type bodyParam struct {
	Ns        string             `json:"ns"`
	ResType   string             `json:"type"`
	ResId     string             `json:"resourceid"`
	UpdateMap map[string]string  `json:"update"`
	Rl        model.ResourceList `json:"resourcelist"`
	R         model.Resource     `json:"resource"`
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster, tree node.TreeMethod) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
		tree:    tree,
		router:  httprouter.New(),
		session: NewSession(),
		logger:  log.New("INFO", "http", model.LogBackend),
	}
}

// Start the server
func (s *Service) Start() error {
	s.initHandler()

	server := http.Server{
		//Handler: accessLog(cors(s.auth(s.router))),
		Handler: accessLog(cors(s.router)),
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

func (s *Service) initHandler() {
	s.router.POST("/api/v1/resource", s.handlerResourceSet)
	s.router.POST("/api/v1/resource/add", s.handlerResourceAdd)
	s.router.GET("/api/v1/resource", s.handlerResourceGet)
	s.router.GET("/api/v1/resource/search", s.handlerSearch)
	s.router.PUT("/api/v1/resource", s.handleResourcePut)
	s.router.PUT("/api/v1/resource/move", s.handleResourceMove)
	s.router.DELETE("/api/v1/resource", s.handleResourceDel)

	s.router.POST("/api/v1/ns", s.handlerNsNew)
	s.router.PUT("/api/v1/ns", s.handlerNsUpdate)
	s.router.GET("/api/v1/ns", s.handlerNsGet)
	s.router.DELETE("/api/v1/ns", s.handlerNsDel)

	s.router.POST("/api/v1/agent/ns", s.handlerRegister)
	s.router.GET("/api/v1/agent/resource", s.handlerResourceGet)
	s.router.PUT("/api/v1/agent/report", s.handlerAgentReport)

	s.router.POST("/api/v1/peer", s.handlerJoin)
	s.router.DELETE("/api/v1/peer", s.handlerRemove)
	s.router.GET("/api/v1/backup", s.handlerBackup)
	s.router.GET("/api/v1/restore", s.handlerRestore)

	s.router.POST("/api/v1/user/signin", s.HandlerSignin)
	s.router.GET("/api/v1/user/signout", s.HandlerSignout)
}

func cors(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set(`Access-Control-Allow-Origin`, origin)
			w.Header().Set(`Access-Control-Allow-Methods`, strings.Join([]string{
				`DELETE`,
				`GET`,
				`OPTIONS`,
				`POST`,
				`PUT`,
			}, ", "))

			w.Header().Set(`Access-Control-Allow-Headers`, strings.Join([]string{
				`Accept`,
				`Accept-Encoding`,
				`Authorization`,
				`Content-Length`,
				`Content-Type`,
				`X-CSRF-Token`,
				`X-HTTP-Method-Override`,
				`Authtoken`,
				`X-Requested-With`,
			}, ", "))
		}

		if r.Method == "OPTIONS" {
			return
		}

		inner.ServeHTTP(w, r)
	})
}

func accessLog(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stime := time.Now().UnixNano() / 1e3
		inner.ServeHTTP(w, r)
		dur := time.Now().UnixNano()/1e3 - stime
		if dur <= 1e3 {
			log.Infof("access path %s in %d us\n", r.URL.Path, dur)
		} else {
			log.Infof("access path %s in %d ms\n", r.URL.Path, dur/1e3)
		}
	})
}

func (s *Service) auth(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !uriFilter(r) {
			inner.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("AuthToken")
		v := s.session.Get(key)
		log.Infof("Header AuthToken: %s - %s", key, v)
		if v == nil {
			ReturnJson(w, 401, "Not Authorized")
			return
		}
		uid, ok := v.(string)
		if !ok {
			ReturnJson(w, 401, "Not Authorized")
			return
		}

		list := strings.Split(r.URL.Path, "/")
		if len(list) < 5 {
			ReturnJson(w, 401, "Not Authorized")
			return
		}
		// TODO: auth filter
		if !permCheck(list[4], uid, r.Method) {
			ReturnJson(w, 401, "Not Authorized")
			return
		}
		w.Header().Set(`UID`, uid)
		r.Header.Set(`UID`, uid)
		inner.ServeHTTP(w, r)
	})
}

// pass agent or router backend requests, this API shuold be almost desinged in GET method.
func uriFilter(r *http.Request) bool {
	var UNAUTH_URI = []string{"/api/v1/user/signin", "/api/v1/agent", "/api/v1/router", "/api/v1/alarm"}
	for _, uri := range UNAUTH_URI {
		if strings.HasPrefix(r.RequestURI, uri) {
			return false
		}
	}
	return true
}

func permCheck(ns string, uid string, method string) bool {
	return true
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

func (s *Service) handlerAgentReport(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	report := m.Report{}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	if err := report.UnmarshalJSON(buf.Bytes()); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if report.NewHostname != "" && report.OldHostname != "" && report.NewHostname != report.OldHostname {
		if err := s.tree.MachineRename(report.OldHostname, report.NewHostname); err != nil {
			ReturnBadRequest(w, err)
			return
		}
	}
	// TODO: process the report time/version.
	ReturnOK(w, "success")
}

func (s *Service) handleResourceMove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromNs := r.FormValue("from")
	toNs := r.FormValue("to")
	resType := r.FormValue("type")
	resId := r.FormValue("resourceid")
	if err := s.tree.MoveResource(fromNs, toNs, resType, resId); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) handlerResourceSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	param := bodyParam{}
	if err := json.Unmarshal(buf.Bytes(), &param); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if param.Ns != "" {
		err = s.tree.SetResource(param.Ns, param.ResType, param.Rl)
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
	var resList *model.ResourceList
	ns := r.FormValue("ns")
	resType := r.FormValue("type")

	if ns != "" {
		resList, err = s.tree.GetResourceList(ns, resType)
	} else {
		ReturnBadRequest(w, fmt.Errorf("invalid infomation"))
		return
	}
	if err != nil {
		ReturnServerError(w, err)
		return
	}
	if len(*resList) == 0 {
		ReturnNotFound(w, "No resources found.")
		return
	}
	ReturnJson(w, 200, resList)
}

func (s *Service) handleResourcePut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	param := bodyParam{}
	if err := json.Unmarshal(buf.Bytes(), &param); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if err := s.tree.UpdateResource(param.Ns, param.ResType, param.ResId, param.UpdateMap); err != nil {
		ReturnBadRequest(w, err)
		return
	} else {
		ReturnOK(w, "success")
	}
}

func (s *Service) handlerResourceAdd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	param := bodyParam{}
	if err := json.Unmarshal(buf.Bytes(), &param); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	delete(param.R, "_id")
	if uuid, err := s.tree.AppendResource(param.Ns, param.ResType, param.R); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, uuid)
	}
}

// search bucket by nodes/key(resource)/resource_property
// TODO: return only or preperty ns or some property of resource from res.
func (s *Service) handlerSearch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	resType := r.FormValue("type")
	k := r.FormValue("k")
	v := r.FormValue("v")
	searchMod := r.FormValue("mod")
	search := model.ResourceSearch{
		Key:   k,
		Value: []byte(v),
		Fuzzy: searchMod == "fuzzy",
	}
	res, err := s.tree.SearchResource(ns, resType, search)
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

func (s *Service) handleResourceDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	resType := r.FormValue("type")
	resID := r.FormValue("resourceid")
	if err := s.tree.DeleteResource(ns, resType, resID); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
	}
}

func (s *Service) handlerNsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var nodes *node.Node
	var err error
	ns := r.FormValue("ns")
	// nodename := r.FormValue("nodename")

	if ns == "" {
		nodes, err = s.tree.AllNodes()
		if err != nil {
			ReturnServerError(w, err)
			return
		}
	} else {
		nodes, err = s.tree.GetNode(ns)
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
	parentNs := r.FormValue("ns")
	name := r.FormValue("name")
	nodeType := r.FormValue("type")
	machineMatch := r.FormValue("machinereg")

	nodeT, err := strconv.Atoi(nodeType)
	if name == "" || parentNs == "" || err != nil || (nodeT != node.Leaf && nodeT != node.NonLeaf) {
		ReturnServerError(w, fmt.Errorf("invalid information"))
		return
	}

	if _, err = s.tree.NewNode(name, parentNs, nodeT, machineMatch); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
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
	ns := r.FormValue("ns")

	if err := s.tree.DelNode(ns); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}
