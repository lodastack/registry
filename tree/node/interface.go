package node

import (
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/tree/cluster"
)

// Inf is the interface node have.
type Inf interface {
	// GetNodeByNS return the node by ns.
	// e.g return root node if get ns $RootNode, default is "loda".
	GetNodeByNS(ns string) (*Node, error)

	// LeafChildIDs return leaf child node ID list of the ns.
	LeafChildIDs(ns string) ([]string, error)

	// GetNodeIDByNS return the NS of the node ID.
	GetNodeIDByNS(ns string) (string, error)

	// GetNodeNSByID returh the node ID of the ns.
	GetNodeNSByID(id string) (string, error)

	// AllNodes return the root node.
	AllNodes() (*Node, error)

	// Save the []byte of the nodes.
	Save(nodeData []byte) error
}

type node struct {
	cluster cluster.Inf
}

// return a node interface object.
func NewNode(cluster cluster.Inf) Inf {
	return &node{cluster: cluster}
}

func (n *node) Save(nodeByte []byte) error {
	return n.cluster.Update([]byte(NodeDataBucketID), []byte(NodeDataKey), nodeByte)
}

// Get value from cluster by bucketID and resType.
func (m *node) getByteFromStore(bucketID, resType string) ([]byte, error) {
	return m.cluster.View([]byte(bucketID), []byte(resType))
}

// get allnodes from cluster.
func (m *node) allNodeByte() ([]byte, error) {
	return m.getByteFromStore(NodeDataBucketID, NodeDataKey)
}

// AllNodes return the whole nodes.
func (m *node) AllNodes() (*Node, error) {
	v, err := m.allNodeByte()
	if err != nil || len(v) == 0 {
		return nil, common.ErrGetNode
	}

	allNode, err := getAllNodeByByte(v)
	if err != nil {
		return nil, common.ErrGetNode
	}
	return allNode, nil
}

// GetNSByID return NS by NodeID.
func (m *node) GetNodeNSByID(id string) (string, error) {
	nodes, err := m.AllNodes()
	if err != nil {
		return "", err
	}
	_, ns, err := nodes.GetByID(id)
	if err != nil {
		return "", err
	}
	return ns, nil
}

func (m *node) GetNodeIDByNS(ns string) (string, error) {
	node, err := m.GetNodeByNS(ns)
	if err != nil {
		return "", err
	}
	return node.ID, nil
}

// GetNodesById return exact node with name.
func (m *node) GetNodeByNS(ns string) (*Node, error) {
	if ns == "" {
		return nil, common.ErrInvalidParam
	}
	// TODO: use nodeidKey as cache
	nodes, err := m.AllNodes()
	if err != nil {
		return nil, err
	}
	return nodes.GetByNS(ns)
}

// Return leaf IDs of the ns.
func (m *node) LeafChildIDs(ns string) ([]string, error) {
	// check the ns exist and valid or not.
	nodeID, err := m.GetNodeIDByNS(ns)
	if nodeID == "" || err != nil {
		return nil, err
	}

	// read the tree if not get from cache.
	node, err := m.GetNodeByNS(ns)
	if err != nil {
		return nil, err
	}
	if node.Type == Leaf {
		return []string{node.ID}, nil
	}
	return node.LeafChildIDs()
}
