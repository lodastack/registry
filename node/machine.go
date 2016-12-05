package node

import (
	"errors"
	"regexp"

	"github.com/lodastack/registry/model"
)

var (
	HostnameProp = "hostname"

	ErrInvalidMachine = errors.New("invalid machine resource")
)

// Search hostname on the tree.
// Return map[ns]resourceID.
func (t *Tree) SearchMachine(hostname string) (map[string]string, error) {
	searchHostname := model.ResourceSearch{
		Key:   HostnameProp,
		Value: []byte(hostname),
		Fuzzy: false,
	}
	resMap, err := t.SearchResource(rootNode, "machine", searchHostname)
	if err != nil {
		t.logger.Errorf("SearchResource fail, error:%s", err.Error())
		return nil, err
	}

	machineRes := map[string]string{}
	for ns, machines := range resMap {
		if len(*machines) == 0 {
			continue
		}
		machineID, _ := (*machines)[0].ID()
		if machineID == "" {
			t.logger.Errorf("machine search resout have no ID, ns: %s, machine: %+v", ns, machines)
			continue
		}
		machineRes[ns] = machineID
	}
	return machineRes, nil
}

func (t *Tree) MachineRename(oldName, newName string) error {
	location, err := t.SearchMachine(oldName)
	if err != nil {
		t.logger.Error("SearchMachine fail: %s", err.Error())
		return err
	}
	if len(location) == 0 {
		return errors.New("machine not found")
	}
	updateMap := map[string]string{HostnameProp: newName}
	for ns, resId := range location {
		if err := t.UpdateResource(ns, "machine", resId, updateMap); err != nil {
			t.logger.Error("MachineRename fail and skip, oldname: %s, newname: %s, fail ns: %s, error: %s",
				oldName, newName, ns, err.Error())
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
		UUID, err := t.AppendResource(ns, "machine", newMachine)
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
