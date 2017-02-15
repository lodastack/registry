package httpd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lodastack/log"
	m "github.com/lodastack/models"
	"github.com/lodastack/registry/authorize"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/node"
	"github.com/lodastack/registry/utils"

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

	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string]string, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// RemoveKey removes the key from the bucket.
	RemoveKey(bucket, key []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// GetSession returns the sression value for the given key.
	GetSession(key interface{}) interface{}

	// SetSession sets the value for the given key, via distributed consensus.
	SetSession(key, value interface{}) error

	// DelSession delete the value for the given key, via distributed consensus.
	DelSession(key interface{}) error

	// Backup database.
	Backup() ([]byte, error)

	// Restore restores backup data file.
	Restore(backupfile string) error

	Peers() (map[string]map[string]string, error)
}

// Service provides HTTP service.
type Service struct {
	addr  string
	ln    net.Listener
	https bool
	cert  string
	key   string

	router *httprouter.Router

	cluster Cluster
	tree    node.TreeMethod
	perm    authorize.Perm

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

var ErrInvalidParam = errors.New("invalid infomation")

// New returns an uninitialized HTTP service.
func New(c config.HTTPConfig, cluster Cluster) (*Service, error) {
	// init Tree
	tree, err := node.NewTree(cluster)
	if err != nil {
		fmt.Println("init tree fail: %s", err.Error())
		return nil, err
	}

	// init authorize
	perm, err := authorize.NewPerm(cluster)
	if err != nil {
		fmt.Printf("init authorize fail: %s\n", err.Error())
		return nil, err
	}

	return &Service{
		addr:    c.Bind,
		https:   c.Https,
		cert:    c.Cert,
		key:     c.Key,
		cluster: cluster,
		tree:    tree,
		perm:    perm,
		router:  httprouter.New(),
		logger:  log.New("INFO", "http", model.LogBackend),
	}, nil
}

// Start the server
func (s *Service) Start() error {
	s.initHandler()

	server := http.Server{}
	if config.C.LDAPConf.Enable {
		server.Handler = s.accessLog(cors(s.auth(s.router)))
	} else {
		server.Handler = s.accessLog(cors(s.router))
	}

	// Open listener.
	if s.https {
		cert, err := tls.LoadX509KeyPair(s.cert, s.key)
		if err != nil {
			return err
		}

		listener, err := tls.Listen("tcp", s.addr, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		if err != nil {
			return err
		}

		s.logger.Println(fmt.Sprint("Listening on HTTPS:", s.addr))
		s.ln = listener
	} else {

		ln, err := net.Listen("tcp", s.addr)
		if err != nil {
			return err
		}

		s.ln = ln
	}

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
	s.router.DELETE("/api/v1/resource/collect", s.handleCollectDel)

	s.router.POST("/api/v1/ns", s.handlerNsNew)
	s.router.PUT("/api/v1/ns", s.handlerNsUpdate)
	s.router.GET("/api/v1/ns", s.handlerNsGet)
	s.router.DELETE("/api/v1/ns", s.handlerNsDel)

	s.router.POST("/api/v1/agent/ns", s.handlerRegister)
	s.router.GET("/api/v1/agent/resource", s.handlerResourceGet)
	s.router.POST("/api/v1/agent/report", s.handlerAgentReport)

	s.router.GET("/api/v1/router/resource", s.handlerResourceGet)
	s.router.GET("/api/v1/router/ns", s.handlerNsGet)

	s.router.GET("/api/v1/peer", s.handlerPeers)
	s.router.POST("/api/v1/peer", s.handlerJoin)
	s.router.DELETE("/api/v1/peer", s.handlerRemove)
	s.router.GET("/api/v1/db/backup", s.handlerBackup)
	s.router.GET("/api/v1/db/restore", s.handlerRestore)

	s.initPermissionHandler()
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
				`NS`,
				`Resource`,
			}, ", "))
		}

		if r.Method == "OPTIONS" {
			return
		}

		inner.ServeHTTP(w, r)
	})
}

func (s *Service) accessLog(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stime := time.Now().UnixNano() / 1e3
		inner.ServeHTTP(w, r)
		dur := time.Now().UnixNano()/1e3 - stime
		if dur <= 1e3 {
			s.logger.Infof("access %s path %s in %d us\n", r.Method, r.URL.Path, dur)
		} else {
			s.logger.Infof("access %s path %s in %d ms\n", r.Method, r.URL.Path, dur/1e3)
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
		v := s.cluster.GetSession(key)
		s.logger.Infof("Header AuthToken: %s - %s", key, v)
		if v == nil {
			ReturnJson(w, 401, "Not Authorized")
			return
		}
		uid, ok := v.(string)
		if !ok {
			ReturnJson(w, 401, "Not Authorized")
			return
		}

		ns := r.Header.Get("NS")
		res := r.Header.Get("Resource")
		if ok, err := s.perm.Check(uid, ns, res, r.Method); err != nil {
			s.logger.Errorf("check permission fail, error: %s", err.Error())
			ReturnServerError(w, err)
			return
		} else if !ok {
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
	var UNAUTH_URI = []string{"/api/v1/user/signin", "/api/v1/user/signout", "/api/v1/agent", "/api/v1/router",
		"/api/v1/alarm", "/api/v1/peer"}
	for _, uri := range UNAUTH_URI {
		if strings.HasPrefix(r.RequestURI, uri) {
			return false
		}
	}
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
		ReturnBadRequest(w, ErrInvalidParam)
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
	if report.Update {
		if report.OldHostname == "" {
			ReturnBadRequest(w, ErrInvalidParam)
			return
		}
		updateMap := map[string]string{}
		if report.NewHostname != "" && report.NewHostname != report.OldHostname {
			updateMap[node.HostnameProp] = report.NewHostname
		}
		if len(report.NewIPList) != 0 &&
			len(report.OldIPList) != 0 &&
			strings.Join(report.NewIPList, ",") != strings.Join(report.OldIPList, ",") {
			updateMap[node.IpProp] = strings.Join(report.NewIPList, ",")
		}
		if err := s.tree.MachineUpdate(report.OldHostname, updateMap); err != nil {
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
	if err := s.tree.MoveResource(fromNs, toNs, resType, strings.Split(resId, ",")...); err != nil {
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
		ReturnBadRequest(w, ErrInvalidParam)
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
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err != nil {
		ReturnServerError(w, err)
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
	if param.Ns == "" || param.ResType == "" || param.R == nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if param.ResType == "collect" && !model.UpdateCollectName(param.R) {
		s.logger.Errorf("add invalid type collect: %+v", param.R)
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	// Check pk property.
	pk := model.PkProperty[param.ResType]
	pkValue, _ := param.R.ReadProperty(pk)
	if pkValue == "" {
		s.logger.Errorf("cannot append resource without pk: %+v", param.R)
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	// Check whether the pk property of the resource is already exist.
	search, _ := model.NewSearch(false, pk, pkValue)
	res, err := s.tree.SearchResource(param.Ns, param.ResType, search)
	if err != nil {
		s.logger.Errorf("check the addend resource fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	} else if len(res) != 0 {
		s.logger.Errorf("resource already exist in the ns, data: %+v", res)
		ReturnBadRequest(w, errors.New("resource already exist"))
		return
	}

	delete(param.R, model.IdKey)
	if err := s.tree.AppendResource(param.Ns, param.ResType, param.R); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
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
	if ns == "" || resType == "" || k == "" || v == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	search, _ := model.NewSearch(searchMod == "fuzzy", k, v)

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
	resIDs := r.FormValue("resourceid")
	if err := s.tree.DeleteResource(ns, resType, strings.Split(resIDs, ",")...); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

// handleCollectDel handle the delete collect request.
// delete the collect resource and clear data in db.
func (s *Service) handleCollectDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	measurements := r.FormValue("measurements")
	resNames, ok := model.GetResNameFromMeasurements(strings.Split(measurements, ","))
	if !ok {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	delDataMeasurements := make([]string, 0)
	resIDs := make([]string, 0)
	// search collect resource and get the ID.
	for _, resName := range resNames {
		search, _ := model.NewSearch(false, model.PkProperty[model.Collect], resName)
		res, err := s.tree.SearchResource(ns, model.Collect, search)
		if err != nil {
			s.logger.Errorf("check the addend resource fail: %s", err.Error())
			ReturnServerError(w, err)
			return
		} else if len(res) == 0 {
			s.logger.Errorf("cannot search collect resource %s in ns: %s, skip this", resName, ns)
			continue
		}

		for _, r := range *res[ns] {
			if resId, ok := r.ID(); ok {
				delDataMeasurements = append(delDataMeasurements, resName)
				resIDs = append(resIDs, resId)
			}
		}
	}

	if len(resIDs) == 0 {
		s.logger.Errorf("search measurement result is nil, measurements: %s,  ns: %s", measurements, ns)
		ReturnServerError(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DeleteResource(ns, model.Collect, resIDs...); err != nil {
		ReturnServerError(w, err)
		return
	}

	// delete collect data
	for _, delName := range delDataMeasurements {
		go func() {
			time.Sleep(90 * time.Second)
			req := utils.HttpQuery{
				Method: http.MethodDelete,
				Url: fmt.Sprintf("http://%s?ns=collect.%s&name=%s&regexp=true",
					config.C.RouterAddr, ns, delName),
				BodyType: utils.Form,
				Timeout:  10}
			if err := req.DoQuery(); err != nil || req.Result.Status > 299 {
				s.logger.Errorf("del data fail: %s, error: %v, result: %+v",
					req.Url, err, req.Result)
			}
		}()
	}
	ReturnOK(w, "success")
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
	// leaf NS list format handler
	// param["format"] = "list"
	if r.FormValue("format") == "list" {
		list, err := nodes.LeafNs()
		if err != nil {
			ReturnServerError(w, err)
			return
		}
		ReturnJson(w, 200, list)
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
		ReturnServerError(w, ErrInvalidParam)
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
