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
	resMap, err := t.SearchResourceByNs(rootNode, "machine", searchHostname)
	if err != nil {
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
		return nil, ErrInvalidMachine
	}

	nsList, err := t.MatchNs(hostname)
	if err != nil {
		return nil, err
	}

	NsIDMap := map[string]string{}
	for _, ns := range nsList {
		nodeID, err := t.getIDByNs(ns)
		if err != nil {
			t.logger.Errorf("getID of ns %s fail when register machine, the whole ns is: %+v, error: %+v", ns, nsList, err)
			// TODO: rollback by RmResByMap(NsIDMap, "machine")
			return nil, err
		}
		UUID, err := t.appendResourceByNodeID(nodeID, "machine", newMachine)
		if err != nil {
			t.logger.Errorf("append machine %+v to ns %s fail, ns list: %+v error: %+v", newMachine, ns, nsList, err)
			// TODO: rollback by RmResByMap(NsIDMap, "machine")
			return nil, err
		}
		NsIDMap[ns] = UUID
	}
	return NsIDMap, nil
}
