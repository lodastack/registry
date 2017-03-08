package httpd

import (
	"net/http"
	"strings"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"

	"github.com/julienschmidt/httprouter"
)

// UserToken struct
type UserToken struct {
	User  string `json:"user"`
	Token string `json:"token"`
}

func (s *Service) initPermissionHandler() {
	s.router.POST("/api/v1/user/signin", s.HandlerSignin)
	s.router.GET("/api/v1/user/signout", s.HandlerSignout)

	s.router.GET("/api/v1/perm/group", s.HandlerGroupGet)
	s.router.GET("/api/v1/perm/group/list", s.HandlerGroupList)
	s.router.POST("/api/v1/perm/group", s.HandlerGroupCreate)
	s.router.PUT("/api/v1/perm/group/item", s.HandlerUpdateGroupItem)
	s.router.PUT("/api/v1/perm/group/member", s.HandlerUpdateGroupMember)
	s.router.DELETE("/api/v1/perm/group", s.HandlerRemoveGroup)

	s.router.GET("/api/v1/event/group", s.HandlerGroupGet)

	s.router.GET("/api/v1/perm/user", s.HandlerUserGet)
	s.router.PUT("/api/v1/perm/user", s.HandlerUserSet)
	s.router.DELETE("/api/v1/perm/user", s.HandlerRemoveUser)
}

// SigninHandler handler signin request
func (s *Service) HandlerSignin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := strings.ToLower(r.FormValue("username"))
	pass := r.FormValue("password")
	if user == "" || pass == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if config.C.LDAPConf.Enable {
		if err := LDAPAuth(user, pass); err != nil {
			ReturnServerError(w, err)
			return
		}
	}
	key := common.GenUUID()
	if err := s.cluster.SetSession(key, user); err != nil {
		ReturnJson(w, 500, "set session failed")
		return
	}

	ok, err := s.perm.CheckUserExist(user)
	if err != nil {
		s.logger.Errorf("check user fail: %s", err.Error())
	} else if !ok {
		// create user if first login.
		if err = s.perm.SetUser(user, nil); err != nil {
			s.logger.Errorf("set user fail: %s", err.Error())
		}
	}

	if err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, UserToken{User: user, Token: key})
}

//SignoutHandler handler signout request
func (s *Service) HandlerSignout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user string
	key := r.Header.Get("AuthToken")
	v := s.cluster.GetSession(key)
	if v == nil {
		ReturnJson(w, 200, UserToken{Token: key})
		return
	}
	user = v.(string)
	s.cluster.DelSession(key)
	ReturnJson(w, 200, UserToken{User: user, Token: key})
}

// HandlerGroupGet handle query group resquest
func (s *Service) HandlerGroupGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gName := strings.ToLower(r.FormValue("gname"))
	if gName == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	g, err := s.perm.GetGroup(gName)
	if err != nil || &g == nil {
		ReturnNotFound(w, "group not found")
		return
	}
	ReturnJson(w, 200, g)
}

// HandlerGroupGet handle query group list resquest
func (s *Service) HandlerGroupList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := strings.ToLower(r.FormValue("ns"))
	if ns == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	gList, err := s.perm.ListNsGroup(ns)
	if err != nil {
		ReturnNotFound(w, "group not found")
		return
	}
	ReturnJson(w, 200, gList)
}

func (s *Service) HandlerGroupCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gName := strings.ToLower(r.FormValue("gname"))
	ns := r.FormValue("ns")
	itemStr := r.FormValue("items")
	if ns != "" {
		gName = s.perm.GetGNameByNs(ns) + "-" + gName
	}
	if gName == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	// TODO: auto members
	err := s.perm.CreateGroup(gName,
		strings.Split(itemStr, ","))
	if err != nil {
		s.logger.Errorf("set group fail: %s", err.Error())
		ReturnNotFound(w, err.Error())
		return
	}
	ReturnOK(w, "success")
}

// HandlerGroupGet handle update group resquest
func (s *Service) HandlerUpdateGroupItem(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gName := strings.ToLower(r.FormValue("gname"))
	itemStr := r.FormValue("items")
	if gName == "" || itemStr == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	err := s.perm.UpdateItems(gName, strings.Split(itemStr, ","))
	if err != nil {
		s.logger.Errorf("set group fail: %s", err.Error())
		ReturnNotFound(w, "set group fail")
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) HandlerUpdateGroupMember(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gName := strings.ToLower(r.FormValue("gname"))
	action := r.FormValue("action")
	managerStr := r.FormValue("managers")
	memberStr := r.FormValue("members")
	managers, members := []string{}, []string{}
	if managerStr != "" {
		managers = strings.Split(managerStr, ",")
	}
	if memberStr != "" {
		members = strings.Split(memberStr, ",")
	}
	if err := s.perm.UpdateMember(gName, managers, members, action); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnOK(w, "success")
}

// HandlerRemoveGroup handle remove group request
func (s *Service) HandlerRemoveGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gName := strings.ToLower(r.FormValue("gname"))
	if gName == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.perm.RemoveGroup(gName); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "success")
}

// HandlerGroupGet handle query user resquest
func (s *Service) HandlerUserGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := strings.ToLower(r.FormValue("username"))
	if username == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	u, err := s.perm.GetUser(username)
	if err != nil || &u == nil {
		ReturnNotFound(w, "user not found")
		return
	}
	ReturnJson(w, 200, u)
}

// HandlerGroupGet handle set user resquest
func (s *Service) HandlerUserSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := strings.ToLower(r.FormValue("username"))
	dashboardStr := r.FormValue("dashboards")
	if username == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	dashboards := strings.Split(dashboardStr, ",")

	if err := s.perm.SetUser(username, dashboards); err != nil {
		s.logger.Errorf("set user fail: %s", err.Error())
		ReturnNotFound(w, "set user fail")
		return
	}
	ReturnOK(w, "success")
}

func (s *Service) HandlerRemoveUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := strings.ToLower(r.FormValue("username"))
	if username == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.perm.RemoveUser(username); err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "success")
}
