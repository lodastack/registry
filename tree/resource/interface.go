package resource

// Node has many type resource such as machine/collect/alarm/group/dashbord...
// Resource is property of node.
// Leaf node have resource; Nonleaf node have resource template which used when create child node.

import (
	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/cluster"
	"github.com/lodastack/registry/tree/node"
)

// Inf is the interface resource have.
type Inf interface {
	// GetResourceList return a type resource list of a node.
	GetResourceList(ns string, resourceType string) (*model.ResourceList, error)

	// GetResource return the one resource of the ns.
	GetResource(ns, resType string, resourceID ...string) ([]model.Resource, error)

	// SetResource set the resource list to the ns.
	SetResource(ns, resType string, rl model.ResourceList) error

	// RemoveResource remove one resource from a node.
	RemoveResource(ns, resType string, resID ...string) error

	// UpdateResource update one resource by updateMap.
	UpdateResource(ns, resType, resID string, updateMap map[string]string) error

	// AppendResource append resources to a ns.
	AppendResource(ns, resType string, appendRes ...model.Resource) error

	// MoveResource move one resource fo an other ns, the resouce will be removed from the old ns.
	MoveResource(oldNs, newNs, resType string, resourceIDs ...string) error

	// CopyResource copy one resource from one ns to the other ns, the resource will still exist in the old ns.
	CopyResource(fromNs, toNs, resType string, resourceIDs ...string) error

	// SearchResource search any preperty resource in the ns and its child ns.
	// Set the ResourceSearch.Key zero value if search the resource all proprety.
	SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error)
}

type resourceMethod struct {
	cluster cluster.Inf
	node    node.Inf
	logger  *log.Logger
}

// NewResource return the reource interface.
func NewResource(cluster cluster.Inf, node node.Inf, logger *log.Logger) Inf {
	return &resourceMethod{cluster: cluster, node: node, logger: logger}
}
