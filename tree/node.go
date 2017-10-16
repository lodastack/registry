package tree

import (
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/tree/node"
)

// AllNodes return the root node.
func (t *Tree) AllNodes() (n *node.Node, err error) {
	if n, err = t.node.AllNodes(); err != nil {
		t.logger.Errorf("AllNodes fail, node %v, error: %s", *n, err.Error())
	}
	return
}

// GetNodeByNS return node by ns.
func (t *Tree) GetNodeByNS(ns string) (n *node.Node, err error) {
	if ns == "" {
		t.logger.Errorf("GetNodeByNS donot allow to query empty ns")
		return nil, common.ErrInvalidParam
	}
	if n, err = t.node.GetNodeByNS(ns); err != nil {
		t.logger.Errorf("GetNode fail, ns: %s, node: %v, err: %s", ns, n, err.Error())
		return nil, common.ErrInvalidParam
	}
	return
}

// getNodeNSByID return node by node ID.
func (t *Tree) getNodeNSByID(id string) (ns string, err error) {
	if ns, err = t.node.GetNodeNSByID(id); err != nil {
		t.logger.Errorf("GetNodeByNS fail: %s", err.Error())
	}
	return
}

func (t *Tree) getNodeIDByNS(ns string) (id string, err error) {
	if id, err = t.node.GetNodeIDByNS(ns); err != nil {
		t.logger.Errorf("GetNodeIDByNS fail: %s", err.Error())
	}
	return
}

// LeafChildIDs return leaf node of the ns.
func (t *Tree) LeafChildIDs(ns string) (l []string, err error) {
	if l, err = t.node.LeafChildIDs(ns); err != nil {
		t.logger.Errorf("LeafChildIDs fail: %s", err.Error())
	}
	return
}
