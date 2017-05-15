package httpd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"github.com/lodastack/registry/model"
)

func (s *Service) initDashboardHandler() {
	s.router.GET("/api/v1/dashboard", s.handlerDashboardGet)
	s.router.POST("/api/v1/dashboard", s.handlerDashboardSet)
	s.router.PUT("/api/v1/dashboard", s.handlerDashboardPut)
	s.router.DELETE("/api/v1/dashboard", s.handlerDashboardDel)

	s.router.POST("/api/v1/dashboard/panel", s.handlerPanelPost)
	s.router.PUT("/api/v1/dashboard/panel", s.handlerPanelPut)
	s.router.PUT("/api/v1/dashboard/panel/order", s.handlerPanelReorder)
	s.router.DELETE("/api/v1/dashboard/panel", s.handlerPanelDel)

	s.router.POST("/api/v1/dashboard/target", s.handlerTargetPost)
	s.router.PUT("/api/v1/dashboard/target", s.handlerTargetPut)
	s.router.DELETE("/api/v1/dashboard/target", s.handlerTargetDelete)
}

func (s *Service) handlerDashboardGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns := r.FormValue("ns")
	if ns == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	dashboards, err := s.tree.GetDashboard(ns)
	if err != nil {
		s.logger.Errorf("handlerDashboardGet GetDashboard fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, dashboards)
}

func (s *Service) handlerDashboardSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	dashboards := make(map[string]model.Dashboard)
	if err := json.Unmarshal(buf.Bytes(), &dashboards); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns := r.FormValue("ns")
	if err := s.tree.SetDashboard(ns, dashboards); err != nil {
		s.logger.Errorf("handlerDashboardGet SetDashboard fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerDashboardPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, name, title := r.FormValue("ns"), r.FormValue("name"), r.FormValue("title")
	if ns == "" || name == "" || title == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.UpdateDashboard(ns, name, title); err != nil {
		s.logger.Errorf("handlerDashboardPut GetDashboard fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerDashboardDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, name := r.FormValue("ns"), r.FormValue("name")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DeleteDashboard(ns, name); err != nil {
		s.logger.Errorf("delete dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var panel model.Panel
	if err := json.Unmarshal(buf.Bytes(), &panel); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns, name := r.FormValue("ns"), r.FormValue("name")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.AddPanel(ns, name, panel); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, name, title, graphType, index :=
		r.FormValue("ns"), r.FormValue("name"), r.FormValue("title"), r.FormValue("type"), r.FormValue("index")
	i, err := strconv.Atoi(index)
	if ns == "" || name == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.UpdatePanel(ns, name, i, title, graphType); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelReorder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var newOrder []int
	if err := json.Unmarshal(buf.Bytes(), &newOrder); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns, name := r.FormValue("ns"), r.FormValue("name")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.ReorderPanel(ns, name, newOrder); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, name, index := r.FormValue("ns"), r.FormValue("name"), r.FormValue("index")
	i, err := strconv.Atoi(index)
	if ns == "" || name == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DelPanel(ns, name, i); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerTargetPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var target model.Target
	if err := json.Unmarshal(buf.Bytes(), &target); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns, name, panelIndex := r.FormValue("ns"), r.FormValue("name"), r.FormValue("pIndex")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	pIndex, err := strconv.Atoi(panelIndex)
	if err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.AppendTarget(ns, name, pIndex, target); err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, "OK")
}
func (s *Service) handlerTargetPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var target model.Target
	if err := json.Unmarshal(buf.Bytes(), &target); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns, name, panelIndex, targetIndex := r.FormValue("ns"), r.FormValue("name"), r.FormValue("pIndex"), r.FormValue("tIndex")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	pIndex, err := strconv.Atoi(panelIndex)
	if err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	tIndex, err := strconv.Atoi(targetIndex)
	if err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.UpdateTarget(ns, name, pIndex, tIndex, target); err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, "OK")
}
func (s *Service) handlerTargetDelete(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, name, panelIndex, targetIndex := r.FormValue("ns"), r.FormValue("name"), r.FormValue("pIndex"), r.FormValue("tIndex")
	if ns == "" || name == "" {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	pIndex, err := strconv.Atoi(panelIndex)
	if err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	tIndex, err := strconv.Atoi(targetIndex)
	if err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DelTarget(ns, name, pIndex, tIndex); err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, "OK")
}
