package node

import (
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/tree/cluster"
)

type NodeInf interface {
	GetNodeByNS(ns string) (*Node, error)
	LeafChildIDs(ns string) ([]string, error)
	GetNodeIDByNS(ns string) (string, error)
	GetNodeNSByID(id string) (string, error)
	AllNodes() (*Node, error)
}

type NodeMethod struct {
	c cluster.ClusterInf
}

func NewNodeMethod(c cluster.ClusterInf) NodeInf {
	return &NodeMethod{c: c}
}

// Get value from cluster by bucketID and resType.
func (m *NodeMethod) getByteFromStore(bucketID, resType string) ([]byte, error) {
	return m.c.View([]byte(bucketID), []byte(resType))
}

// get allnodes from cluster.
func (m *NodeMethod) allNodeByte() ([]byte, error) {
	return m.getByteFromStore(NodeDataBucketID, NodeDataKey)
}

// AllNodes return the whole nodes.
func (m *NodeMethod) AllNodes() (*Node, error) {
	v, err := m.allNodeByte()
	if err != nil || len(v) == 0 {
		return nil, common.ErrGetNode
	}

	var allNode Node
	if err := allNode.UnmarshalJSON(v); err != nil {
		return nil, common.ErrGetNode
	}
	return &allNode, nil
}

// GetNSByID return NS by NodeID.
func (m *NodeMethod) GetNodeNSByID(id string) (string, error) {
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

func (m *NodeMethod) GetNodeIDByNS(ns string) (string, error) {
	node, err := m.GetNodeByNS(ns)
	if err != nil {
		return "", err
	}
	return node.ID, nil
}

// GetNodesById return exact node with name.
func (m *NodeMethod) GetNodeByNS(ns string) (*Node, error) {
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
func (m *NodeMethod) LeafChildIDs(ns string) ([]string, error) {
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
