package node

import (
	"strings"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
)

const (
	// RootNode is the root node name.
	RootNode = "loda"
	// PoolNode is the global pool node.
	PoolNode = "pool"
	// NodeDeli join node to ns.
	// e.g The global pool node is the the leaf child of root node, its ns is pool.loda.
	NodeDeli = "."

	// NodeDataBucketID is the bucketID to save node data.
	NodeDataBucketID = "loda"
	// NodeDataKey is the node data key.
	NodeDataKey = "node"

	// NsFormat
	NsFormat = "ns"
	// IDFormat
	IDFormat = "id"
	// NotMatchMachine is defaul not match machine.
	NotMatchMachine = "^$"
)

const (
	// Leaf node type
	Leaf = iota // leaf type of node
	// NonLeaf node type
	NonLeaf
	// Root type of node
	Root
)

// InitNodes auto creates nodes when registry init.
type nodeMeta struct {
	Name    string
	Tp      int
	Comment string
}

// InitNodes auto creates nodes when registry init.
var InitNodes = []nodeMeta{
	{Name: "pool.loda", Tp: Leaf, Comment: "pool"},
	{Name: "monitor.loda", Tp: NonLeaf, Comment: "monitor system"},
	{Name: "db.monitor.loda", Tp: NonLeaf, Comment: "monitor system"},
	{Name: "common.db.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "alarm.monitor.loda", Tp: NonLeaf, Comment: "monitor system"},
	{Name: "kapacitor.alarm.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "adapter.alarm.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "nodata.alarm.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "event.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "router.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "registry.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "mq.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "etcd.monitor.loda", Tp: Leaf, Comment: "monitor system"},
	{Name: "ui.monitor.loda", Tp: Leaf, Comment: "monitor system"},
}

// NodeProperty is node should has.
type NodeProperty struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Type    int    `json:"type"`

	// regexp of machine in one node,
	// used to auto add a machine into nodes
	MachineReg string `json:"machinereg"`
}

// Node is the item of tree, it has machine match stregy and resource.
type Node struct {
	NodeProperty
	Children []*Node `json:"children"`
}

// IsLeaf return the node is leaf or not.
func (n *Node) IsLeaf() bool {
	return n.Type == Leaf
}

// Exist check the ns is exist already or not.
func (n *Node) Exist(ns string) bool {
	if _, err := n.GetByNS(ns); err == nil {
		return true
	}
	return false
}

// Copy return a copy of a Node.
func (n *Node) Copy(ori *Node) *Node {
	n = &Node{
		NodeProperty{
			ID:         ori.ID,
			Name:       ori.Name,
			Comment:    ori.Comment,
			Type:       ori.Type,
			MachineReg: ori.MachineReg,
		},
		make([]*Node, len(ori.Children)),
	}

	for i := range ori.Children {
		n.Children[i] = (&Node{}).Copy(ori.Children[i])
	}
	return n
}

// Update update node machineMatchStrategy property.
func (n *Node) Update(name, comment, machineMatchStrategy string) {
	if name != "" {
		n.Name = name
	}
	if comment != "" {
		n.Comment = comment
	}
	if machineMatchStrategy != "" {
		n.MachineReg = machineMatchStrategy
	}
}

// allowDel check if the node allow be delete.
// only leaf or no child nonleaf allow to be delete.
func (n *Node) allowDel() bool {
	if n.Type == Leaf {
		return true
	}
	if len(n.Children) == 0 {
		return true
	}
	return false
}

// RemoveChildNode remove the child by node ID.
func (n *Node) RemoveChildNode(childID string) error {
	for index, child := range n.Children {
		if child.ID != childID {
			continue
		}
		if !child.allowDel() {
			return common.ErrNotAllowDel
		}
		copy(n.Children[index:], n.Children[index+1:])
		n.Children = n.Children[:len(n.Children)-1]
		return nil
	}
	return common.ErrNodeNotFound
}

// AllowResource checks if the node could be set a resource.
// Leaf node could set/get resource;
// NonLeaf node could be only set/get template resource.
//
// If check fail, the node is NonLeaf and resType not template,
// maybe need get/set at its leaf child node.
func (n *Node) AllowResource(resType string) bool {
	if n.IsLeaf() {
		return true
	}
	if len(resType) > len(model.TemplatePrefix) &&
		string(resType[:len(model.TemplatePrefix)]) == model.TemplatePrefix {
		return true
	}
	return false
}

func getKeysOfMap(ori map[string]string) []string {
	keys := make([]string, len(ori))
	i := 0
	for key := range ori {
		keys[i] = key
		i++
	}
	return keys
}

// WalkfFun is the type of the function for each node visited by Walk.
// The node argument is the node the walkFunc will process.
// The childReturn argument will pass the nodes's childNode return.
//
// If an error was returned, processing stops.
type WalkfFun func(node *Node, childReturn map[string]string) (map[string]string, error)

// Walk the node.
func (n *Node) Walk(walkFun WalkfFun) (map[string]string, error) {
	if n.Type == Leaf {
		return walkFun(n, nil)
	}

	childReturn := map[string]string{}
	for index := range n.Children {
		oneChild, err := n.Children[index].Walk(walkFun)
		if err != nil {
			return nil, err
		}
		for k, v := range oneChild {
			childReturn[k] = v
		}
	}
	return walkFun(n, childReturn)
}

// LeafNs return all leaf child ns.
func (n *Node) LeafNs() ([]string, error) {
	nsMap, err := n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		result := map[string]string{}
		if node.Type == Leaf {
			result[node.Name] = ""
		} else {
			for chindNs := range childReturn {
				result[chindNs+NodeDeli+node.Name] = ""
			}
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return getKeysOfMap(nsMap), nil
}

// LeafChildIDs return the leaf id list of this Node.
func (n *Node) LeafChildIDs() ([]string, error) {
	IDMap, err := n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		result := map[string]string{}
		if node.Type == Leaf {
			result[node.ID] = ""
		} else {
			for chindID := range childReturn {
				result[chindID] = ""
			}
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	} else if len(IDMap) == 0 {
		return nil, common.ErrNoLeafChild
	}
	return getKeysOfMap(IDMap), nil
}

// LeafMachineReg return the ns-MachineReg Map.
func (n *Node) LeafMachineReg() (map[string]string, error) {
	return n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		result := map[string]string{}
		if node.Type == Leaf {
			result[node.Name] = node.MachineReg
		} else {
			for relativeNs, reg := range childReturn {
				result[relativeNs+NodeDeli+node.Name] = reg
			}
		}
		return result, nil
	})
}

// GetByID return exact node and ns which with nodeid.
func (n *Node) GetByID(nodeID string) (*Node, string, error) {
	if n.ID == nodeID {
		return n, n.Name, nil
	}
	for index := range n.Children {
		if detNode, ns, err := n.Children[index].GetByID(nodeID); err == nil {
			return detNode, ns + NodeDeli + n.Name, nil
		}
	}
	return nil, "", common.ErrNodeNotFound
}

// GetByNS return exact node by nodename.
func (n *Node) GetByNS(ns string) (*Node, error) {
	nsSplit := strings.Split(ns, NodeDeli)
	if len(nsSplit) == 1 && ns == RootNode {
		// return tree if get root.
		return n, nil
	} else if len(nsSplit) < 2 {
		// the query is invalid.
		return nil, common.ErrNodeNotFound
	}

	// Func to check if children node match the ns.
	// Get name of next part of the ns, search it in the child nodes.
	checkChild := func(node *Node, nsSplit []string) (*Node, bool) {
		nextNsPart := nsSplit[len(nsSplit)-2]
		for index := range node.Children {
			if node.Children[index].Name == nextNsPart {
				return node.Children[index], true
			}
		}
		return nil, false
	}

	if RootNode != nsSplit[len(nsSplit)-1] {
		return nil, common.ErrNodeNotFound
	}

	checkNode := n
	var ok bool
	// Seach each part of the ns, finally get the node of the ns.
	for index := range nsSplit {
		checkNode, ok = checkChild(checkNode, nsSplit[0:len(nsSplit)-index])
		// Return error if not match.
		if !ok {
			return nil, common.ErrNodeNotFound
		}
		// If each part of the ns is match, return.
		if index+1 == len(nsSplit)-1 {
			break
		}
	}
	return checkNode, nil
}

// TODO: finish the comment
// get nodeID-childIDs map of this node.
func (n *Node) getChildMap() (map[string]string, error) {
	leafCache := map[string]string{}
	_, err := n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		if node.Type == Leaf {
			return map[string]string{node.ID: ""}, nil
		}
		childIDs := ""
		for LeafID := range childReturn {
			childIDs += LeafID + ","
		}
		leafCache[node.ID] = strings.TrimRight(childIDs, ",")
		return childReturn, nil
	})
	if err != nil {
		return nil, err
	}
	return leafCache, nil
}
