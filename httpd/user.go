package httpd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lodastack/registry/authorize"
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
	s.router.GET("/api/v1/user/wework/signin", s.HandlerWeworkSignin)
	s.router.GET("/api/v1/user/signout", s.HandlerSignout)

	s.router.GET("/api/v1/perm/group", s.HandlerGroupGet)
	s.router.GET("/api/v1/perm/group/list", s.HandlerGroupList)
	s.router.POST("/api/v1/perm/group", s.HandlerGroupCreate)
	s.router.PUT("/api/v1/perm/group/item", s.HandlerUpdateGroupItem)
	s.router.PUT("/api/v1/perm/group/member", s.HandlerUpdateGroupMember)
	s.router.DELETE("/api/v1/perm/group", s.HandlerRemoveGroup)

	s.router.GET("/api/v1/event/group", s.HandlerGroupGet)

	s.router.GET("/api/v1/perm/user", s.HandlerUserGet)
	s.router.GET("/api/v1/perm/user/list", s.HandlerUserListGet)
	s.router.PUT("/api/v1/perm/user", s.HandlerUserSet)
	s.router.DELETE("/api/v1/perm/user", s.HandlerRemoveUser)

	// response ok if the request pass permission check.
	s.router.GET("/api/v1/perm/check", s.nilHandler)
	s.router.POST("/api/v1/perm/check", s.nilHandler)
	s.router.PUT("/api/v1/perm/check", s.nilHandler)
	s.router.DELETE("/api/v1/perm/check", s.nilHandler)
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

	ok, err := s.perm.CheckUserExist(user)
	if err != nil {
		s.logger.Errorf("check user fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	} else if !ok {
		// return 500 if the user is first login.
		ReturnServerError(w, errors.New("You have no permission, contact the administrators"))
		return
	}

	key := common.GenUUID()
	if err := s.cluster.SetSession(key, user); err != nil {
		ReturnServerError(w, errors.New("set session failed"))
		return
	}
	ReturnJson(w, 200, UserToken{User: user, Token: key})
}

const WCtokenAPI = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"

var WCTOKEN string
var WCTOKEN_CREATETIME int64

/*
{
   "access_token": "accesstoken000001",
   "expires_in": 7200
}
*/

type tokenResp struct {
	Token   string `json:"access_token"`
	Expires int    `json:"expires_in"`
}

func token(cid string, csec string, refresh bool) error {
	var diff int64
	diff = 7000
	now := time.Now().Unix()
	if now-WCTOKEN_CREATETIME < diff && WCTOKEN != "" && !refresh {
		return nil
	}
	urlencoded := fmt.Sprintf(WCtokenAPI, url.QueryEscape(cid), url.QueryEscape(csec))
	res, err := http.Get(urlencoded)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	WCTOKEN_CREATETIME = time.Now().Unix()
	if res.StatusCode != 200 {
		return fmt.Errorf("remote server not 200", res.StatusCode)
	}
	var response tokenResp
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&response); err != nil {
		return err
	}
	if response.Token != "" {
		WCTOKEN = response.Token
		return nil
	} else {
		return fmt.Errorf("empty token", response.Token)
	}
}

// HandlerWeworkSignin handler wechat server request
// https://work.weixin.qq.com/api/doc
func (s *Service) HandlerWeworkSignin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code := r.FormValue("code")
	// state := r.FormValue("state")
	if code == "" {
		http.Redirect(w, r, "/", 302)
		return
	}

	if !config.C.WeworkConf.Enable {
		ReturnServerError(w, fmt.Errorf("wework login not enabled"))
		return
	}

	token(config.C.WeworkConf.CorpID, config.C.WeworkConf.CorpSecret, false)

	u := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", WCTOKEN, code)
	resp, err := http.Get(u)
	if err != nil {
		ReturnServerError(w, err)
		return
	}
	defer resp.Body.Close()

	type wwResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		UserID  string `json:"UserId"`
	}

	var wwr wwResp
	if resp.StatusCode == 200 {
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&wwr)
	} else {
		ReturnServerError(w, fmt.Errorf("status not OK"))
		return
	}

	if wwr.ErrCode != 0 {
		ReturnServerError(w, fmt.Errorf("error code not 0, got "+string(wwr.ErrCode)))
		return
	}

	key := common.GenUUID()
	if err := s.cluster.SetSession(key, wwr.UserID); err != nil {
		ReturnServerError(w, errors.New("set session failed"))
		return
	}

	ur := fmt.Sprintf(config.C.WeworkConf.Redirect, key, wwr.UserID)
	http.Redirect(w, r, ur, 302)
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

// HandlerGroupList handle query group list resquest
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
	name := strings.ToLower(r.FormValue("gname"))
	ns := r.FormValue("ns")
	itemStr := r.FormValue("items")
	managers := r.FormValue("managers")
	members := r.FormValue("members")

	for _, gnameLetter := range name {
		if gnameLetter >= 'a' && gnameLetter <= 'z' {
			continue
		}
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	gName := ""
	if ns != "" {
		gName = authorize.GetGNameByNs(ns, name)
	}
	if gName == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	// TODO: auto members
	err := s.perm.CreateGroup(gName,
		strings.Split(managers, ","),
		strings.Split(members, ","),
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
	managerStr := r.FormValue("managers")
	memberStr := r.FormValue("members")
	managers, members := []string{}, []string{}
	if managerStr != "" {
		managers = strings.Split(managerStr, ",")
	}
	if memberStr != "" {
		members = strings.Split(memberStr, ",")
	}

	allUser := append(managers, members...)
	for _, user := range allUser {
		ok, err := s.perm.CheckUserExist(user)
		if err != nil {
			ReturnNotFound(w, "check user fail")
			return
		} else if !ok {
			if ldapExist := LDAPUserExist(user); !ldapExist {
				ReturnNotFound(w, "unknow user "+user)
				return
			}
			if err = s.perm.SetUser(user, "", "enable", ""); err != nil {
				s.logger.Errorf("set user fail: %s", err.Error())
				ReturnNotFound(w, "set user fail")
				return
			}
		}
	}

	if err := s.perm.UpdateMember(gName, managers, members); err != nil {
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

// HandlerUserListGet handle query user list resquest
func (s *Service) HandlerUserListGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameStr := r.FormValue("usernames")
	usernames := strings.Split(usernameStr, ",")

	userData, err := s.perm.GetUserList(usernames)
	if err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, userData)
}

// HandlerUserSet handle set user resquest
func (s *Service) HandlerUserSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := strings.ToLower(r.FormValue("username"))
	mobile := r.FormValue("mobile")
	alert := r.FormValue("alert")
	accessToken := r.FormValue("accesstoken")
	if username == "" ||
		(r.Header.Get(`UID`) != "" && username != r.Header.Get(`UID`)) {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.perm.SetUser(username, mobile, alert, accessToken); err != nil {
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

func (s *Service) nilHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ReturnJson(w, 200, "success")
}
