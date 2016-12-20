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
	s.router.PUT("/api/v1/perm/group", s.HandlerGroupPut)
	s.router.GET("/api/v1/perm/user", s.HandlerUserGet)
	s.router.PUT("/api/v1/perm/user", s.HandlerUserSet)
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
		if err = s.perm.SetUser(user, nil, nil); err != nil {
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
	gId := strings.ToLower(r.FormValue("gid"))
	if gId == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	g, err := s.perm.GetGroup(gId)
	if err != nil || &g == nil {
		ReturnNotFound(w, "group not found")
		return
	}
	ReturnJson(w, 200, g)
}

// HandlerGroupGet handle update group resquest
func (s *Service) HandlerGroupPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	gId := strings.ToLower(r.FormValue("gid"))
	managerStr := r.FormValue("managers")
	// TODO: ToLower
	itemStr := r.FormValue("items")
	if gId == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	var managers, items []string
	if managerStr != "" {
		managers = strings.Split(managerStr, ",")
	}
	if itemStr != "" {
		items = strings.Split(itemStr, ",")
	}

	_, err := s.perm.SetGroup(gId, managers, items)
	if err != nil {
		s.logger.Errorf("set group fail: %s", err.Error())
		ReturnNotFound(w, "set group fail")
		return
	}
	ReturnOK(w, "success")
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
	gidStr := r.FormValue("gids")
	dashboardStr := r.FormValue("dashboards")
	if username == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	var gids, dashboards []string
	if gidStr != "" {
		gids = strings.Split(gidStr, ",")
	}
	if dashboardStr != "" {
		dashboards = strings.Split(dashboardStr, ",")
	}

	if err := s.perm.SetUser(username, gids, dashboards); err != nil {
		s.logger.Errorf("set user fail: %s", err.Error())
		ReturnNotFound(w, "set user fail")
		return
	}
	ReturnOK(w, "success")
}
