package tree

import (
	"fmt"

	m "github.com/lodastack/models"
	"github.com/lodastack/registry/model"
)

func (t *Tree) RegisterMachine(newMachine model.Resource) (map[string]string, error) {
	return t.m.RegisterMachine(newMachine)
}

func (t *Tree) SearchMachine(hostname string) (map[string]string, error) {
	return t.m.SearchMachine(hostname)
}

func (t *Tree) MachineUpdate(oldName string, updateMap map[string]string) error {
	return t.m.MachineUpdate(oldName, updateMap)
}

func (t *Tree) CheckMachineStatusByReport(reports map[string]m.Report) error {
	return t.m.CheckMachineStatusByReport(reports)
}

// UpdateStatusByHostname search the machine and update the status.
// updateMap is map[string]string{HostStatusProp: status}
func (t *Tree) UpdateStatusByHostname(hostname string, updateMap map[string]string) error {
	machineRecord, err := t.m.SearchMachine(hostname)
	if err != nil {
		t.logger.Errorf("UpdateStatusByHostname search machine fail: %s", err.Error())
		return fmt.Errorf("update machine fail, invalid hostname: %s, error: %s", hostname, err.Error())
	}
	for _ns, resourceID := range machineRecord {
		if err := t.r.UpdateResource(_ns, model.Machine, resourceID, updateMap); err != nil {
			t.logger.Errorf("UpdateStatusByHostname update machine fail, ns: %s, resourceID: %s, new status: %+v, error: %s",
				_ns, resourceID, updateMap, err.Error())
			return fmt.Errorf("update machine status fail, hostname %s, error: %s", hostname, err.Error())
		}
	}
	return nil
}
