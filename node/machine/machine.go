package machine

import (
	"errors"
	"regexp"
	"time"

	m "github.com/lodastack/models"
	"github.com/lodastack/registry/model"
	n "github.com/lodastack/registry/node/node"
)

var (
	// unit: hour
	MachineReportTimeout float64 = 48
	MachineReportAlive   float64 = 24

	ErrInvalidMachine = errors.New("invalid machine resource")
)

// Search hostname on the tree.
// Return map[ns]resourceID.
func (m *Machine) SearchMachine(hostname string) (map[string]string, error) {
	searchHostname, err := model.NewSearch(false, model.HostnameProp, hostname)
	if err != nil {
		return nil, err
	}
	resMap, err := m.r.SearchResource(n.RootNode, "machine", searchHostname)
	if err != nil {
		m.logger.Errorf("SearchResource fail, error: %s", err.Error())
		return nil, err
	}

	machineRes := map[string]string{}
	for ns, machines := range resMap {
		if len(*machines) == 0 {
			m.logger.Errorf("machine search resout have no ID, ns: %s, machine: %+v", ns, *machines)
			return nil, errors.New("search machine error")
		}
		machineID, _ := (*machines)[0].ID()
		if machineID == "" {
			m.logger.Errorf("machine search resout have no ID, ns: %s, machine: %+v", ns, *machines)
			return nil, errors.New("search machine fail")
		}
		machineRes[ns] = machineID
	}
	return machineRes, nil
}

func (m *Machine) MachineUpdate(oldName string, updateMap map[string]string) error {
	hostname, ok := updateMap[model.HostnameProp]
	if ok && hostname == "" {
		return ErrInvalidMachine
	}
	ip, ok := updateMap[model.IpProp]
	if ok && ip == "" {
		return ErrInvalidMachine
	}

	location, err := m.SearchMachine(oldName)
	if err != nil {
		m.logger.Error("SearchMachine fail: %s", err.Error())
		return err
	}
	if len(location) == 0 {
		return errors.New("machine not found")
	}
	for ns, resId := range location {
		if err := m.r.UpdateResource(ns, "machine", resId, updateMap); err != nil {
			m.logger.Errorf("MachineRename fail and skip, oldname: %s, ns: %s, update: %+v, error: %s",
				oldName, ns, updateMap, err.Error())
			return err
		}
	}
	return nil
}

// Return the ns which MachineReg match the hostname.
// If there is not ns be match, return the pool ns.
func (m *Machine) MatchNs(hostname string) ([]string, error) {
	nodes, err := m.n.AllNodes()
	if err != nil {
		return nil, err
	}
	leafReg, err := nodes.LeafMachineReg()
	if err != nil {
		return nil, err
	}

	nsList := []string{}
	for ns, reg := range leafReg {
		// Skip the ^$ regular expressions.
		if reg == n.NotMatchMachine {
			continue
		}
		match, err := regexp.MatchString(reg, hostname)
		if err != nil || !match {
			continue
		}
		nsList = append(nsList, ns)
	}
	if len(nsList) == 0 {
		nsList = append(nsList, n.PoolNode+n.NodeDeli+n.RootNode)
	}
	return nsList, nil
}

// RegisterMachine registry a machine to the tree.
// NewMachine mast have property "hostname", it will be used to judge which ns to register.
func (m *Machine) RegisterMachine(newMachine model.Resource) (map[string]string, error) {
	hostname, ok := newMachine.ReadProperty(model.HostnameProp)
	if !ok {
		m.logger.Errorf("RegisterMachine fail: not provide hostname")
		return nil, ErrInvalidMachine
	}

	nsList, err := m.MatchNs(hostname)
	if err != nil {
		m.logger.Errorf("RegisterMachine fail, MatchNs fail: %s", err.Error())
		return nil, err
	}

	NsIDMap := map[string]string{}
	for _, ns := range nsList {
		UUID := newMachine.InitID()
		err := m.r.AppendResource(ns, "machine", newMachine)
		if err != nil {
			m.logger.Errorf("append machine %+v to ns %s fail when register, the whole ns list: %+v error: %+v",
				newMachine, ns, nsList, err)
			// TODO: rollback by RmResByMap(NsIDMap, "machine")
			return nil, err
		}
		NsIDMap[ns] = UUID
	}
	return NsIDMap, nil
}

func (m *Machine) CheckMachineStatusByReport(reports map[string]m.Report) error {
	nodes, err := m.n.AllNodes()
	if err != nil {
		return err
	}
	allLeaf, err := nodes.LeafNs()
	if err != nil {
		return err
	}

	for _, _ns := range allLeaf {
		machineList, err := m.r.GetResourceList(_ns, "machine")
		if err != nil {
			m.logger.Errorf("get machine of ns %s status fail", _ns)
			continue
		}

		var update bool
		for i := range *machineList {
			hostname, _ := (*machineList)[i].ReadProperty(model.HostnameProp)
			hostStatus, _ := (*machineList)[i].ReadProperty(model.HostStatusProp)
			if hostStatus != model.Online || hostStatus != model.Dead {
				continue
			}

			reportInfo, ok := reports[hostname]
			if !ok {
				// set dead if not report.
				if hostStatus != model.Dead {
					update = true
					(*machineList)[i].SetProperty(model.HostStatusProp, model.Dead)
				}
			} else {
				// set Online/Dead status by report time.
				if time.Now().Sub(reportInfo.UpdateTime).Hours() >= MachineReportTimeout && hostStatus != model.Dead {
					update = true
					(*machineList)[i].SetProperty(model.HostStatusProp, model.Dead)
				} else if time.Now().Sub(reportInfo.UpdateTime).Hours() < MachineReportAlive && hostStatus != model.Online {
					update = true
					(*machineList)[i].SetProperty(model.HostStatusProp, model.Online)
				}
			}
		}
		if update {
			err = m.r.SetResource(_ns, "machine", *machineList)
		}
	}
	return nil
}
