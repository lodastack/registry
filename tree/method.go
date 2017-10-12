package tree

import (
	m "github.com/lodastack/models"
	"github.com/lodastack/registry/model"
	n "github.com/lodastack/registry/tree/node"
)

// TreeMethod is the interface tree must implement.
type TreeMethod interface {
	// AllNodes return all nodes.
	AllNodes() (*n.Node, error)

	// SetAgentInfo set agent info
	AgentReport(info m.Report) error

	// GetAgents return agent info
	GetReportInfo() map[string]m.Report

	// GetNodesById return exact node by nodeid.
	GetNodeByNS(id string) (*n.Node, error)

	// NewNode create node.
	NewNode(name, parentNs string, nodeType int, property ...string) (string, error)

	// GetResource return the resourceList by ns/resource type/resource ID.
	GetResource(ns, resType string, resID ...string) ([]model.Resource, error)

	// Get resource by NodeName and resour type
	GetResourceList(NodeName string, ResourceType string) (*model.ResourceList, error)

	// Set resource to node with nodename.
	SetResource(nodeName, resType string, rl model.ResourceList) error

	// SearchResourceByNs return the map[ns]resources which match the search.
	SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error)

	// Return leaf child node of the ns.
	LeafChildIDs(ns string) ([]string, error)

	// Search Machine on tree.
	SearchMachine(hostname string) (map[string]string, error)

	// Regist machine on the tree.
	RegisterMachine(newMachine model.Resource) (map[string]string, error)

	// Update the node property.
	UpdateNode(ns string, name, machineReg string) error

	// RemoveNode remove the node with delID from parentNs.
	RemoveNode(ns string) error

	// Update Resource By ns and ResourceID.
	UpdateResource(ns, resType, resID string, updateMap map[string]string) error

	// Update hostname property of machine resource.
	MachineUpdate(oldName string, updateMap map[string]string) error

	// Append resource to ns.
	AppendResource(ns, resType string, appendRes ...model.Resource) error

	// CopyResource copy resource from fromNs to toNs.
	CopyResource(fromNs, toNs, resType string, resID ...string) error

	// Remove resource from ns.
	RemoveResource(ns, resType string, resId ...string) error

	// Remove resource from one ns to another.
	MoveResource(oldNs, newNs, resType string, resourceID ...string) error

	// UpdateStatusByHostname update machine status.
	UpdateStatusByHostname(hostname string, updateMap map[string]string) error
	DashboardInf
}
