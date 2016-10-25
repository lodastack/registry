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
}

// TreeMethod is the interface tree must implement.
type TreeMethod interface {
	// GetAllNodes return all nodes.
	GetAllNodes() (*Node, error)

	// GetNodesById return exact node by nodeid.
	GetNodesByID(id string) (*Node, error)

	// NewNode create node.
	NewNode(name, parentId string, nodeType int, property ...string) (string, error)

	GetChild(nodeId string, leaf bool) []string

	GetNsResource(NodeId string, ResourceType string) (*model.Resources, error)
}
