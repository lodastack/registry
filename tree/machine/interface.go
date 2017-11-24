package machine

// Every leaf node could has machine resource which use hostname as primary key.
// Unlike other resources, machine has something special logic and method.
// Agent on the machine will register the machine to the tree.
// Tree would update the machine status(online/dead) according by the agent report.
// Tree could update/remove machine by hostname in all node on the tree.

import (
	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/node"
	"github.com/lodastack/registry/tree/resource"
)

// Inf is the machine resource method.
type Inf interface {
	// RegisterMachine register the machine to the ns which MatchNs return.
	// Return the ns and resource ID map which it registered.
	RegisterMachine(newMachine model.Resource) (map[string]string, error)

	// CheckMachineStatusByReport check the machine is online or dead by its report, update the machine status.
	CheckMachineStatusByReport(reports map[string]model.Report) error

	// SearchMachine search the hostname in all node.
	// Return the result at form of ns-resourceID map if the node has this hostname.
	SearchMachine(hostname string) (map[string][2]string, error)

	// MachineUpdate search the hostname and update the machine resource by updateMap.
	MachineUpdate(sn string, oldHostName string, updateMap map[string]string) error

	// MatchNs walk the all node and check the hostname match the ns or not, return the ns list.
	// If not match any ns, will return the pool node.
	MatchNs(hostname string) ([]string, error)
}

type machine struct {
	node     node.Inf
	resource resource.Inf
	logger   *log.Logger
}

// NewMachine return the obj which has machine interface.
func NewMachine(node node.Inf, resource resource.Inf, logger *log.Logger) Inf {
	return &machine{node: node, resource: resource, logger: logger}
}
