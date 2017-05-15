package node

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/lodastack/registry/model"
)

var (
	dashboardType = "dashboard"
)

var DashboardBuck = "dashboard"

type DashboardInf interface {
	// GetDashboard return dashboard map of the ns.
	GetDashboard(ns string) (model.DashboardData, error)

	// SetDashboard set the dashboard map to the ns.
	SetDashboard(ns string, dashboardData model.DashboardData) error

	// DeleteDashboard update the dashboard of the ns.
	DeleteDashboard(ns, dashboardName string) error

	// Update update the title of dashboard.
	UpdateDashboard(ns, dashboardName, title string) error

	PanelInf
}

type PanelInf interface {
	// ReorderPanel reorder the panel
	ReorderPanel(ns string, dashboardName string, newOrder []int) error

	// AddPanel add the panel to the dashboard.
	AddPanel(ns, dashboardName string, panel model.Panel) error

	// DelPanel delete the panel of the dashboard.
	DelPanel(ns, dashboardName string, panelIndex int) error

	// UpdatePanel update the panel of the dashboard.
	UpdatePanel(ns, dashboardName string, panelIndex int, title, graphType string) error

	// AppendTarget append a target to panel.
	AppendTarget(ns, dashboardName string, panelIndex int, target model.Target) error

	// UpdateTarget update a target.
	UpdateTarget(ns, dashboardName string, panelIndex, targetIndex int, target model.Target) error

	// DelTarget delete a target.
	DelTarget(ns, dashboardName string, panelIndex, targetIndex int) error
}

//  u.cluster.View([]byte(AuthBuck), getUKey(username))
// GetDashboard return the dashboard under the ns.
func (t *Tree) GetDashboard(ns string) (model.DashboardData, error) {
	nodeId, err := t.getID(ns)
	if err != nil {
		t.logger.Errorf("getIDByNs fail: %s", err.Error())
		return nil, err
	}

	resByte, err := t.getByteFromStore(nodeId, dashboardType)
	if err != nil {
		return nil, err
	}
	if len(resByte) == 0 {
		return nil, nil
	}
	rl := make(map[string]model.Dashboard)
	err = json.Unmarshal(resByte, &rl)
	if err != nil {
		t.logger.Errorf("unmarshal resource fail, error: %s, data: %s:", err, string(resByte))
		return nil, err
	}
	return rl, nil
}

// GetDashboard return the dashboard under the ns.
func (t *Tree) SetDashboard(ns string, dashboardData model.DashboardData) error {
	nodeId, err := t.getID(ns)
	if err != nil {
		t.logger.Errorf("getIDByNs fail: %s", err.Error())
		return err
	}
	resNewByte, err := json.Marshal(dashboardData)
	if err != nil {
		t.logger.Errorf("marshal dashboard fail: %s", err.Error())
		return err
	}
	return t.setByteToStore(nodeId, dashboardType, resNewByte)
}

func (t *Tree) UpdateDashboard(ns, dashboardName, title string) error {
	dashboardMap, err := t.GetDashboard(ns)
	if err != nil {
		return err
	}
	if dashboard, ok := dashboardMap[dashboardName]; !ok {
		return errors.New("invalid dashboard name")
	} else {
		dashboard.Title = title
		dashboardMap[dashboardName] = dashboard
	}

	return t.SetDashboard(ns, dashboardMap)
}

func (t *Tree) DeleteDashboard(ns, dashboardName string) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	delete(dashboardData, dashboardName)
	return t.SetDashboard(ns, dashboardData)
}

func (t *Tree) AddPanel(ns, dashboardName string, panel model.Panel) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok {
		return errors.New("not find dashboard " + dashboardName)
	}

	dashboard.Panels = append(dashboard.Panels, panel)
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}

func (t *Tree) DelPanel(ns, dashboardName string, panelIndex int) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) <= panelIndex {
		return errors.New("dashboard name or panel index invalid")
	}

	copy(dashboard.Panels[panelIndex:], dashboard.Panels[panelIndex+1:])
	dashboard.Panels = dashboard.Panels[:len(dashboard.Panels)-1]
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}

func (t *Tree) UpdatePanel(ns, dashboardName string, panelIndex int, title, graphType string) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) <= panelIndex {
		return errors.New("dashboard name or panel index invalid")
	}
	if title != "" {
		dashboard.Panels[panelIndex].Title = title
	}
	if graphType != "" {
		dashboard.Panels[panelIndex].GraphType = graphType
	}
	return t.SetDashboard(ns, dashboardData)
}

func invalidOrder(order sort.IntSlice) bool {
	tmp := make(sort.IntSlice, len(order))
	copy(tmp, order)
	tmp.Sort()
	for i, index := range tmp {
		if i != index {
			return true
		}
	}
	return false
}

func (t *Tree) ReorderPanel(ns string, dashboardName string, newOrder []int) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) != len(newOrder) {
		return errors.New("dashboard name or new order invalid")
	}
	if invalidOrder(newOrder) {
		return errors.New("dashboard new order invalid")
	}

	newPanels := make([]model.Panel, len(dashboard.Panels))
	for i, order := range newOrder {
		newPanels[i] = dashboard.Panels[order]
	}
	// TODO: clear code
	dashboard.Panels = newPanels
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}

// AppendTarget append a target to panel.
func (t *Tree) AppendTarget(ns, dashboardName string, panelIndex int, target model.Target) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) <= panelIndex {
		return errors.New("dashboard name or new order invalid")
	}
	dashboard.Panels[panelIndex].Targets = append(dashboard.Panels[panelIndex].Targets, target)
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}

// UpdateTarget update a target.
func (t *Tree) UpdateTarget(ns, dashboardName string, panelIndex, targetIndex int, target model.Target) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) <= panelIndex || len(dashboard.Panels[panelIndex].Targets) <= targetIndex {
		return errors.New("dashboard name or new order invalid")
	}
	dashboard.Panels[panelIndex].Targets[targetIndex] = target
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}

// DelTarget remove update a target.
func (t *Tree) DelTarget(ns, dashboardName string, panelIndex, targetIndex int) error {
	dashboardData, err := t.GetDashboard(ns)
	if err != nil || len(dashboardData) == 0 {
		t.logger.Errorf("DeleteDashboard error, data: %+v, error: %v", dashboardData, err)
		return err
	}
	dashboard, ok := dashboardData[dashboardName]
	if !ok || len(dashboard.Panels) < panelIndex || len(dashboard.Panels[panelIndex].Targets) < targetIndex {
		return errors.New("dashboard name or new order invalid")
	}
	if len(dashboard.Panels[panelIndex].Targets) == targetIndex+1 {
		tmp := dashboard.Panels[panelIndex].Targets
		tmp = tmp[:len(tmp)-1]
		dashboard.Panels[panelIndex].Targets = tmp
	} else {
		tmp := dashboard.Panels[panelIndex].Targets
		copy(tmp[len(tmp)-1:], tmp[len(tmp):])
		dashboard.Panels[panelIndex].Targets = tmp
	}
	dashboardData[dashboardName] = dashboard
	return t.SetDashboard(ns, dashboardData)
}
