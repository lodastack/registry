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
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)
}

// TreeMethod is the interface tree must implement.
type TreeMethod interface {
	// AllNodes return all nodes.
	AllNodes() (*Node, error)

	// GetNodesById return exact node by nodeid.
	GetNodeByID(id string) (*Node, string, error)

	// NewNode create node.
	NewNode(name, parentId string, nodeType int, property ...string) (string, error)

	// Get resource by NodeID and resour type
	GetResourceByNodeID(NodeId string, ResourceType string) (*model.Resources, error)

	// Get resource by NodeName and resour type
	GetResourceByNs(NodeName string, ResourceType string) (*model.Resources, error)

	// Set Resource to node with nodeid.
	SetResourceByNodeID(nodeId, resType string, ResByte []byte) error

	// Set resource to node with nodename.
	SetResourceByNs(nodeName, resType string, ResByte []byte) error

	// SearchResourceByNs return the map[ns]resources which match the search.
	SearchResourceByNs(ns, resType string, search model.ResourceSearch) (map[string]*model.Resources, error)

	// Return leaf child node of one ns.
	Leaf(ns string, format string) ([]string, error)

	// Search Machine on tree.
	SearchMachine(hostname string) (map[string]string, error)

	// Regist machine on the tree.
	RegisterMachine(newMachine model.Resource) (map[string]string, error)

	// Update the node property.
	UpdateNode(ns string, name, machineReg string) error

	// Delete the node with delID from parentNs.
	DelNode(parentNs, delID string) error
}
