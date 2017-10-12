package tree

import (
	"github.com/lodastack/registry/common"
	n "github.com/lodastack/registry/tree/node"
)

// GetAllNodes return the root node.
func (t *Tree) AllNodes() (n *n.Node, err error) {
	if n, err = t.n.AllNodes(); err != nil {
		t.logger.Errorf("AllNodes fail, node %v, error: %s", *n, err.Error())
	}
	return
}

// GetNodeByNS return node by ns.
func (t *Tree) GetNodeByNS(ns string) (n *n.Node, err error) {
	if ns == "" {
		t.logger.Errorf("GetNodeByNS donot allow to query empty ns")
		return nil, common.ErrInvalidParam
	}
	if n, err = t.n.GetNodeByNS(ns); err != nil {
		t.logger.Errorf("GetNode fail, ns: %s, node: %v, err: %s", ns, n, err.Error())
		return nil, common.ErrInvalidParam
	}
	return
}

// getNodeNSByID return node by node ID.
func (t *Tree) getNodeNSByID(id string) (ns string, err error) {
	if ns, err = t.n.GetNodeNSByID(id); err != nil {
		t.logger.Errorf("GetNodeByNS fail: %s", err.Error())
	}
	return
}

func (t *Tree) getNodeIDByNS(ns string) (id string, err error) {
	if id, err = t.n.GetNodeIDByNS(ns); err != nil {
		t.logger.Errorf("GetNodeIDByNS fail: %s", err.Error())
	}
	return
}

// Return leaf node of the ns.
func (t *Tree) LeafChildIDs(ns string) (l []string, err error) {
	if l, err = t.n.LeafChildIDs(ns); err != nil {
		t.logger.Errorf("LeafChildIDs fail: %s", err.Error())
	}
	return
}
