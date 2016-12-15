package node

import (
	// "github.com/lodastack/log"
	"github.com/lodastack/registry/model"
)

// Cluster is the interface op must implement.
type Cluster interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// Remove removes a node from the store, specified by addr.
	Remove(addr string) error

	// Create a bucket, via distributed consensus.
	CreateBucket(name []byte) error

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// Backup database.
	Backup() ([]byte, error)

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string]string, error)
}

// TreeMethod is the interface tree must implement.
type TreeMethod interface {
	// AllNodes return all nodes.
	AllNodes() (*Node, error)

	// GetNodesById return exact node by nodeid.
	GetNode(id string) (*Node, error)

	// NewNode create node.
	NewNode(name, parentNs string, nodeType int, property ...string) (string, error)

	// Get resource by NodeName and resour type
	GetResourceList(NodeName string, ResourceType string) (*model.ResourceList, error)

	// Set resource to node with nodename.
	SetResource(nodeName, resType string, rl model.ResourceList) error

	// SearchResourceByNs return the map[ns]resources which match the search.
	SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error)

	// Return leaf child node of the ns.
	LeafIDs(ns string) ([]string, error)

	// Search Machine on tree.
	SearchMachine(hostname string) (map[string]string, error)

	// Regist machine on the tree.
	RegisterMachine(newMachine model.Resource) (map[string]string, error)

	// Update the node property.
	UpdateNode(ns string, name, machineReg string) error

	// Delete the node with delID from parentNs.
	DelNode(ns string) error

	// Update Resource By ns and ResourceID.
	UpdateResource(ns, resType, resID string, updateMap map[string]string) error

	// Update hostname property of machine resource.
	MachineRename(oldName, newName string) error

	// Append resource to ns.
	AppendResource(ns, resType string, appendRes ...model.Resource) error

	// Delete resource from ns.
	DeleteResource(ns, resType string, resId ...string) error

	// Remove resource from one ns to another.
	MoveResource(oldNs, newNs, resType string, resourceID ...string) error
}
