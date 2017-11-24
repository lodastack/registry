package tree

import (
	"fmt"

	"github.com/lodastack/registry/model"
)

// RegisterMachine search and register the machine to the node which match the hostname.
func (t *Tree) RegisterMachine(newMachine model.Resource) (map[string]string, error) {
	return t.machine.RegisterMachine(newMachine)
}

// SearchMachine search the hostname in all node.
func (t *Tree) SearchMachine(hostname string) (map[string][2]string, error) {
	return t.machine.SearchMachine(hostname)
}

// MachineUpdate search the hostname and update the machine resource by updateMap.
func (t *Tree) MachineUpdate(sn string, oldName string, updateMap map[string]string) error {
	return t.machine.MachineUpdate(sn, oldName, updateMap)
}

// CheckMachineStatusByReport check the machine is online or dead by its report, update the machine status.
func (t *Tree) CheckMachineStatusByReport(reports map[string]model.Report) error {
	return t.machine.CheckMachineStatusByReport(reports)
}

// UpdateStatusByHostname search the machine and update the status.
// updateMap is map[string]string{HostStatusProp: status}
func (t *Tree) UpdateStatusByHostname(hostname string, updateMap map[string]string) error {
	machineRecord, err := t.machine.SearchMachine(hostname)
	if err != nil {
		t.logger.Errorf("UpdateStatusByHostname search machine fail: %s", err.Error())
		return fmt.Errorf("update machine fail, invalid hostname: %s, error: %s", hostname, err.Error())
	}
	for _ns, resourceID := range machineRecord {
		if err := t.resource.UpdateResource(_ns, model.Machine, resourceID[0], updateMap); err != nil {
			t.logger.Errorf("UpdateStatusByHostname update machine fail, ns: %s, resourceID: %s, new status: %+v, error: %s",
				_ns, resourceID, updateMap, err.Error())
			return fmt.Errorf("update machine status fail, hostname %s, error: %s", hostname, err.Error())
		}
	}
	return nil
}

// RemoveStatusByHostname search and remove the machine by hostname.
func (t *Tree) RemoveStatusByHostname(hostname string) error {
	machineRecord, err := t.machine.SearchMachine(hostname)
	if err != nil {
		t.logger.Errorf("UpdateStatusByHostname search machine fail: %s", err.Error())
		return fmt.Errorf("update machine fail, invalid hostname: %s, error: %s", hostname, err.Error())
	}
	for _ns, resourceID := range machineRecord {
		if err := t.resource.RemoveResource(_ns, model.Machine, resourceID[0]); err != nil {
			t.logger.Errorf("UpdateStatusByHostname update machine fail, ns: %s, resourceID: %s,  error: %s",
				_ns, resourceID, err.Error())
			return fmt.Errorf("update machine status fail, hostname %s, error: %s", hostname, err.Error())
		}
	}
	return nil
}
