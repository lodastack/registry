package node

import (
	"errors"
	"regexp"

	"github.com/lodastack/registry/model"
)

var (
	HostnameProp = "hostname"
	IpProp       = "ip"

	ErrInvalidMachine = errors.New("invalid machine resource")
)

// Search hostname on the tree.
// Return map[ns]resourceID.
func (t *Tree) SearchMachine(hostname string) (map[string]string, error) {
	searchHostname, err := model.NewSearch(false, HostnameProp, hostname)
	if err != nil {
		return nil, err
	}
	resMap, err := t.SearchResource(rootNode, "machine", searchHostname)
	if err != nil {
		t.logger.Errorf("SearchResource fail, error: %s", err.Error())
		return nil, err
	}

	machineRes := map[string]string{}
	for ns, machines := range resMap {
		if len(*machines) == 0 {
			t.logger.Errorf("machine search resout have no ID, ns: %s, machine: %+v", ns, *machines)
			return nil, errors.New("search machine error")
		}
		machineID, _ := (*machines)[0].ID()
		if machineID == "" {
			t.logger.Errorf("machine search resout have no ID, ns: %s, machine: %+v", ns, *machines)
			return nil, errors.New("search machine fail")
		}
		machineRes[ns] = machineID
	}
	return machineRes, nil
}

func (t *Tree) MachineUpdate(oldName string, updateMap map[string]string) error {
	hostname, ok := updateMap[HostnameProp]
	if ok && hostname == "" {
		return ErrInvalidMachine
	}
	ip, ok := updateMap[IpProp]
	if ok && ip == "" {
		return ErrInvalidMachine
	}

	location, err := t.SearchMachine(oldName)
	if err != nil {
		t.logger.Error("SearchMachine fail: %s", err.Error())
		return err
	}
	if len(location) == 0 {
		return errors.New("machine not found")
	}
	for ns, resId := range location {
		if err := t.UpdateResource(ns, "machine", resId, updateMap); err != nil {
			t.logger.Errorf("MachineRename fail and skip, oldname: %s, ns: %s, update: %+v, error: %s",
				oldName, ns, updateMap, err.Error())
			return err
		}
	}
	return nil
}

// Return the ns which MachineReg match the hostname.
// If there is not ns be match, return the pool ns.
func (t *Tree) MatchNs(hostname string) ([]string, error) {
	nodes, err := t.AllNodes()
	if err != nil {
		return nil, err
	}
	leafReg, err := nodes.leafMachineReg()
	if err != nil {
		return nil, err
	}

	nsList := []string{}
	for ns, reg := range leafReg {
		// Skip the ^$ regular expressions.
		if reg == NoMachineMatch {
			continue
		}
		match, err := regexp.MatchString(reg, hostname)
		if err != nil || !match {
			continue
		}
		nsList = append(nsList, ns)
	}
	if len(nsList) == 0 {
		nsList = append(nsList, poolNode+nodeDeli+rootNode)
	}
	return nsList, nil
}

// Register NewMachine on the tree.
// NewMachine mast have property "hostname", it will be used to judge which ns to register.
func (t *Tree) RegisterMachine(newMachine model.Resource) (map[string]string, error) {
	hostname, ok := newMachine.ReadProperty(HostnameProp)
	if !ok {
		t.logger.Errorf("RegisterMachine fail: not provide hostname")
		return nil, ErrInvalidMachine
	}

	nsList, err := t.MatchNs(hostname)
	if err != nil {
		t.logger.Errorf("RegisterMachine fail, MatchNs fail: %s", err.Error())
		return nil, err
	}

	NsIDMap := map[string]string{}
	for _, ns := range nsList {
		UUID := newMachine.InitID()
		err := t.AppendResource(ns, "machine", newMachine)
		if err != nil {
			t.logger.Errorf("append machine %+v to ns %s fail when register, the whole ns list: %+v error: %+v",
				newMachine, ns, nsList, err)
			// TODO: rollback by RmResByMap(NsIDMap, "machine")
			return nil, err
		}
		NsIDMap[ns] = UUID
	}
	return NsIDMap, nil
}
