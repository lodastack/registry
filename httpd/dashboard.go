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
	s.router.POST("/api/v1/dashboard/add", s.handlerDashboardAdd)
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

func (s *Service) handlerDashboardAdd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var dashboard model.Dashboard
	if err := json.Unmarshal(buf.Bytes(), &dashboard); err != nil {
		s.logger.Errorf("unmarshal dashboard fail: %s", err.Error())
		ReturnBadRequest(w, err)
		return
	}

	ns := r.FormValue("ns")
	if err := s.tree.AddDashboard(ns, dashboard); err != nil {
		s.logger.Errorf("handlerDashboardGet SetDashboard fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerDashboardSet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	var dashboards []model.Dashboard
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
	ns, dIndex, title := r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("title")
	i, err := strconv.Atoi(dIndex)
	if ns == "" || title == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.UpdateDashboard(ns, i, title); err != nil {
		s.logger.Errorf("handlerDashboardPut GetDashboard fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerDashboardDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, dIndex := r.FormValue("ns"), r.FormValue("dindex")
	i, err := strconv.Atoi(dIndex)
	if ns == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DeleteDashboard(ns, i); err != nil {
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

	ns, dIndex := r.FormValue("ns"), r.FormValue("dindex")
	i, err := strconv.Atoi(dIndex)
	if ns == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.AddPanel(ns, i, panel); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, dIndex, title, graphType, pIndex :=
		r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("title"), r.FormValue("type"), r.FormValue("pindex")
	dI, errD := strconv.Atoi(dIndex)
	pI, errP := strconv.Atoi(pIndex)
	if ns == "" || errD != nil || errP != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.UpdatePanel(ns, dI, pI, title, graphType); err != nil {
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

	ns, dIndex := r.FormValue("ns"), r.FormValue("dindex")
	i, err := strconv.Atoi(dIndex)
	if ns == "" || err != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.ReorderPanel(ns, i, newOrder); err != nil {
		s.logger.Errorf("AddPanel fail: %s", err.Error())
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerPanelDel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, dIndex, pIndex := r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("pindex")
	dI, errD := strconv.Atoi(dIndex)
	pI, errP := strconv.Atoi(pIndex)
	if ns == "" || errD != nil || errP != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DelPanel(ns, dI, pI); err != nil {
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

	ns, dIndex, pIndex := r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("pindex")
	dI, errD := strconv.Atoi(dIndex)
	pI, errP := strconv.Atoi(pIndex)
	if ns == "" || errD != nil || errP != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.AppendTarget(ns, dI, pI, target); err != nil {
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

	ns, dIndex, pIndex, tIndex := r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("pindex"), r.FormValue("tindex")
	dI, errD := strconv.Atoi(dIndex)
	pI, errP := strconv.Atoi(pIndex)
	tI, errT := strconv.Atoi(tIndex)
	if ns == "" || errD != nil || errP != nil || errT != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}

	if err := s.tree.UpdateTarget(ns, dI, pI, tI, target); err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, "OK")
}

func (s *Service) handlerTargetDelete(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ns, dIndex, pIndex, tIndex := r.FormValue("ns"), r.FormValue("dindex"), r.FormValue("pindex"), r.FormValue("tindex")
	dI, errD := strconv.Atoi(dIndex)
	pI, errP := strconv.Atoi(pIndex)
	tI, errT := strconv.Atoi(tIndex)
	if ns == "" || errD != nil || errP != nil || errT != nil {
		ReturnBadRequest(w, ErrInvalidParam)
		return
	}
	if err := s.tree.DelTarget(ns, dI, pI, tI); err != nil {
		ReturnServerError(w, err)
	}
	ReturnJson(w, 200, "OK")
}
