package httpd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
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
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)

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
	s.router.PUT("/api/v1/resource/copy", s.handleResourceCopy)
	s.router.DELETE("/api/v1/resource", s.handleResourceDel)
	s.router.DELETE("/api/v1/resource/collect", s.handleCollectDel)

	s.router.POST("/api/v1/ns", s.handlerNsNew)
	s.router.PUT("/api/v1/ns", s.handlerNsUpdate)
	s.router.GET("/api/v1/ns", s.handlerNsGet)
	s.router.DELETE("/api/v1/ns", s.handlerNsDel)

	s.router.GET("/api/v1/agents", s.handlerAgents)

	// For agent
	s.router.POST("/api/v1/agent/ns", s.handlerRegister)
	s.router.GET("/api/v1/agent/resource", s.handlerResourceGet)
	s.router.POST("/api/v1/agent/report", s.handlerAgentReport)

	// For router, just allow Get method
	s.router.GET("/api/v1/router/ns", s.handlerNsGet)
	s.router.GET("/api/v1/router/resource", s.handlerResourceGet)

	// For alarm, just allow Get method
	s.router.GET("/api/v1/alarm/ns", s.handlerNsGet)
	s.router.GET("/api/v1/alarm/resource", s.handlerResourceGet)

	// For event, just allow Get method
	s.router.GET("/api/v1/event/ns", s.handlerNsGet)
	s.router.GET("/api/v1/event/resource", s.handlerResourceGet)
	s.router.GET("/api/v1/event/resource/search", s.handlerSearch)
	s.router.GET("/api/v1/event/user/list", s.HandlerUserListGet)

	s.router.GET("/api/v1/peer", s.handlerPeers)
	s.router.POST("/api/v1/peer", s.handlerJoin)
	s.router.DELETE("/api/v1/peer", s.handlerRemove)
	s.router.GET("/api/v1/db/backup", s.handlerBackup)
	s.router.GET("/api/v1/db/restore", s.handlerRestore)

	s.initPermissionHandler()
	s.initDashboardHandler()
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
			ReturnJson(w, 401, "Not Authorized. Please login.")
			return
		}
		uid, ok := v.(string)
		if !ok {
			ReturnJson(w, 401, "Not Authorized. Please login.")
			return
		}

		ns := r.Header.Get("NS")
		res := r.Header.Get("Resource")
		if ok, err := s.perm.Check(uid, ns, res, r.Method); err != nil {
			s.logger.Errorf("check permission fail, error: %s", err.Error())
			ReturnServerError(w, err)
			return
		} else if !ok {
			ReturnJson(w, 403, "Not Authorized. Please check your permission.")
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
		"/api/v1/alarm", "/api/v1/event", "/api/v1/peer"}
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

	// check the machine status.
	if status, _ := machine.ReadProperty(node.HostStatusProp); status == "" {
		machine.SetProperty(node.HostStatusProp, "online")
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
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	if report.OldHostname == "" {
		// log.Errorf("report data invalid %+v", report)
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if report.Update {
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
			log.Errorf("update machine %s fail, data: %+v, error: %s", report.NewHostname, updateMap, err.Error())
			ReturnServerError(w, err)
			return
		}
		if report.NewHostname != "" {
			if err := clearMachineStatus(report.NewHostname, report.Ns...); err != nil {
				log.Errorf("clearMachineStatus ns %v hostname %s fail: %s",
					report.Ns, report.NewHostname, err.Error())
			}
		}
	}

	if err := s.tree.AgentReport(report); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) handleResourceCopy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromNs := r.FormValue("from")
	toNs := r.FormValue("to")
	resType := r.FormValue("type")
	resId := r.FormValue("resourceid")
	if err := s.tree.CopyResource(fromNs, toNs, resType, strings.Split(resId, ",")...); err != nil {
		ReturnServerError(w, err)
		return
	}
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

	if param.ResType == model.Alarm {
		if param.UpdateMap, err = model.NewAlarmResourceByMap(param.Ns, param.UpdateMap, param.ResId); err != nil {
			ReturnBadRequest(w, err)
			return
		}
	}

	if err := s.tree.UpdateResource(param.Ns, param.ResType, param.ResId, param.UpdateMap); err != nil {
		ReturnBadRequest(w, err)
		return
	} else if param.ResType == "machine" {
		machines, err := s.tree.GetResource(param.Ns, param.ResType, param.ResId)
		if len(machines) == 0 && err != nil {
			log.Error("clear ns %s machine %s fail", param.Ns, param.ResId)
		} else {
			hostname, _ := machines[0].ReadProperty(model.PkProperty["machine"])
			if hostname == "" {
				log.Error("clear ns %s machine %s fail: have no hostname", param.Ns, param.ResId)
			} else {
				if err := clearMachineStatus(hostname, param.Ns); err != nil {
					log.Errorf("clearMachineStatus ns %s hostname %s fail: %s",
						param.Ns, hostname, err.Error())
				}
			}
		}
	}
	ReturnOK(w, "success")
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

	if param.ResType == "collect" && model.UpdateCollectName(param.R) != nil {
		s.logger.Errorf("add invalid collect: %+v", param.R)
		ReturnBadRequest(w, ErrInvalidParam)
		return
	} else if param.ResType == "machine" {
		if status, _ := param.R.ReadProperty(node.HostStatusProp); status == "" {
			param.R.SetProperty(node.HostStatusProp, "online")
		}
	}

	// Check pk property.
	resType := param.ResType
	if strings.HasPrefix(param.ResType, model.TemplatePrefix) {
		resType = param.ResType[len(model.TemplatePrefix):]
	}
	pk := model.PkProperty[resType]
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

	if param.ResType != model.Alarm {
		delete(param.R, model.IdKey)
	} else {
		if param.R, err = model.NewAlarmResourceByMap(param.Ns, param.R, ""); err != nil {
			ReturnBadRequest(w, err)
			return
		}
	}

	if err := s.tree.AppendResource(param.Ns, param.ResType, param.R); err != nil {
		ReturnServerError(w, err)
	} else {
		if param.ResType == "collect" {
			gDevName := authorize.GetNsDevGName(param.Ns)
			gOpName := authorize.GetNsOpGName(param.Ns)
			if alarms, err := model.GetAlarmFromCollect(param.R, param.Ns, gDevName+","+gOpName); err == nil && len(alarms) != 0 {
				if err := s.tree.AppendResource(param.Ns, model.Alarm, alarms...); err != nil {
					ReturnServerError(w, err)
				}
			}
		}
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
	if ns == "" || resType == "" || v == "" {
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
				resIDs = append(resIDs, resId)
			}
		}
	}

	if len(resIDs) != 0 {
		if err := s.tree.DeleteResource(ns, model.Collect, resIDs...); err != nil {
			ReturnServerError(w, err)
			return
		}
	}

	// delete collect data
	for _, resName := range resNames {
		go func() {
			time.Sleep(90 * time.Second)
			req := utils.HttpQuery{
				Method: http.MethodDelete,
				Url: fmt.Sprintf("http://%s?ns=collect.%s&name=%s&regexp=true",
					config.C.RouterAddr, ns, resName),
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
	if r.Header.Get(`UID`) != "" {
		u, err := s.perm.GetUser(r.Header.Get(`UID`))
		if err != nil || &u == nil {
			ReturnNotFound(w, "user not found")
			return
		}

		// init a nodes.
		nodeHasPermission := &node.Node{
			node.NodeProperty{
				ID:         (*nodes).ID,
				Name:       (*nodes).Name,
				Type:       (*nodes).Type,
				MachineReg: (*nodes).MachineReg,
			},
			[]*node.Node{}}

		// check the group and set ns to nodeHasPermission.
		var gNames sort.StringSlice = u.Groups
		gNames.Sort()
		for _, gName := range gNames {
			_gNs, gName := s.perm.ReadGName(gName)
			switch gName {
			case authorize.AdminGName:
				if _gNs == "loda" {
					nodeHasPermission = nodes
					break
				}
			case authorize.DefaultGName:
				continue
			default:
				_gNsSplit := strings.Split(_gNs, ".")
				_gNsLength, lenNsRoot := len(_gNsSplit), 1
				if _gNsLength == 1 && _gNsSplit[_gNsLength-1] == "loda" {
					nodeHasPermission = nodes
					break
				}
				nodePointer := nodeHasPermission
				for i := 1 + lenNsRoot; i <= _gNsLength; i++ {
					nsToCheck := strings.Join(_gNsSplit[_gNsLength-i:_gNsLength], ".")
					nodeOnTree, err := nodes.Get(nsToCheck)
					if err != nil {
						break
					}
					if _node, err := nodeHasPermission.Get(nsToCheck); err != nil {
						if i == _gNsLength {
							nodePointer.Children = append(nodePointer.Children, nodeOnTree)
							break
						}
						newNode := &node.Node{
							node.NodeProperty{
								ID:         nodeOnTree.ID,
								Name:       nodeOnTree.Name,
								Type:       nodeOnTree.Type,
								MachineReg: nodeOnTree.MachineReg,
							},
							[]*node.Node{}}
						nodePointer.Children = append(nodePointer.Children, newNode)
						nodePointer = newNode
					} else {
						nodePointer = _node
					}
				}
			}
		}
		nodes = nodeHasPermission
	}

	// leaf NS list format handler
	if r.FormValue("format") == "list" {
		list, err := nodes.LeafNs()
		if err != nil {
			ReturnServerError(w, err)
			return
		}
		if nsSplit := strings.Split(ns, "."); len(nsSplit) > 1 {
			nsSurfix := strings.Join(nsSplit[1:], ".")
			for i := range list {
				list[i] = list[i] + "." + nsSurfix
			}
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
	devStr := r.FormValue("devs")
	opStr := r.FormValue("ops")

	nodeT, err := strconv.Atoi(nodeType)
	if name == "" || parentNs == "" || err != nil || (nodeT != node.Leaf && nodeT != node.NonLeaf) {
		ReturnServerError(w, ErrInvalidParam)
		return
	}

	var ns, gOpName, gDevName string
	var ops, devs []string
	ns = name + "." + parentNs
	if len(ns) > 64-len("collect.") {
		ReturnBadRequest(w, errors.New("The ns name is to long, please check and re-operate."))
		return
	}
	for _, nsLetter := range name {
		if nsLetter == '-' || (nsLetter >= 'a' && nsLetter <= 'z') || (nsLetter >= '0' && nsLetter <= '9') {
			continue
		}
		ReturnBadRequest(w, errors.New("The ns name only allows numbers/letters/crossed, please check and re-operate."))
		return
	}

	if _, err = s.tree.NewNode(name, parentNs, nodeT, machineMatch); err != nil {
		ReturnServerError(w, err)
		return
	}

	creater := r.Header.Get(`UID`)

	if opStr != "" {
		ops = strings.Split(opStr, ",")
	} else {
		ops = []string{creater}
	}
	err = s.perm.CreateGroup(authorize.GetNsOpGName(ns), ops, ops, s.perm.AdminGroupItems(ns))
	if err != nil {
		ReturnServerError(w, fmt.Errorf("Create op group %s fail: %s", gOpName, err.Error()))
		return
	}

	if devStr != "" {
		devs = strings.Split(devStr, ",")
	} else {
		devs = []string{creater}
	}
	gDevName = authorize.GetNsDevGName(ns)
	err = s.perm.CreateGroup(gDevName, devs, devs, s.perm.DefaultGroupItems(ns))
	if err != nil {
		ReturnServerError(w, fmt.Errorf("Create dev group %s fail: %s", gDevName, err.Error()))
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) handlerNsUpdate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	name := r.FormValue("name")
	machinereg := r.FormValue("machinereg")

	for _, nsLetter := range name {
		if nsLetter == '-' || (nsLetter >= 'a' && nsLetter <= 'z') || (nsLetter >= '0' && nsLetter <= '9') {
			continue
		}
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

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

	if err := s.perm.RemoveGroup(authorize.GetNsOpGName(ns)); err != nil {
		log.Errorf("remove ns admin group fail: %s", err.Error())
	}
	if err := s.perm.RemoveGroup(authorize.GetNsDevGName(ns)); err != nil {
		log.Errorf("remove ns admin group fail: %s", err.Error())
	}
	ReturnOK(w, "success")
}

func (s *Service) handlerAgents(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ReturnJson(w, 200, s.tree.GetReportInfo())
}
