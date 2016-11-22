package node

import (
	"errors"
	"strings"
	"sync"
)

const (
	Leaf    = iota // leaf type of node
	NonLeaf        // non-leaf type of node
	Root
)

var (
	ErrInitNodeBucket      = errors.New("init node bucket fail")
	ErrInitNodeKey         = errors.New("init node bucket k-v fail")
	ErrGetNode             = errors.New("get node fail")
	ErrNodeNotFound        = errors.New("node not found")
	ErrGetParent           = errors.New("get parent node error")
	ErrCreateNodeUnderLeaf = errors.New("can not create node under leaf node")
	ErrSetResourceToLeaf   = errors.New("can not set resource to leaf node")
	ErrGetNodeID           = errors.New("get nodeid fail")
	ErrInvalidParam        = errors.New("invalid param")
	ErrNilChildNode        = errors.New("get none child node")
	ErrNodeAlreadyExist    = errors.New("node already exist")
	ErrNoLeafChild         = errors.New("have no leaf child node")
	ErrNotAllowDel         = errors.New("not allow to be delete")
)

type NodeProperty struct {
	ID   string
	Name string
	Type int

	// regexp of machine in one node,
	// used to auto add a machine into nodes
	MachineReg string
}

type Node struct {
	NodeProperty
	Children []*Node
}

func (n *Node) IsLeaf() bool {
	return n.Type == Leaf
}

func (n *Node) Exist(ns string) bool {
	if _, err := n.GetByNs(ns); err == nil {
		return true
	}
	return false
}

// update node property, do not change children.
func (n *Node) update(name, machineReg string) {
	if name != "" {
		n.Name = name
	}
	if machineReg != "" {
		n.MachineReg = machineReg
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

// delChild delete one child node by ID.
func (n *Node) delChild(childId string) error {
	for index, child := range n.Children {
		if child.ID != childId {
			continue
		}
		if !child.allowDel() {
			return ErrNotAllowDel
		}
		copy(n.Children[index:], n.Children[index+1:])
		n.Children = n.Children[:len(n.Children)-1]
		return nil
	}
	return ErrNodeNotFound
}

// AllowSetResource checks if the node could be set a resource.
// Leaf node could set/get resource;
// NonLeaf node could be only set/get template resource.
//
// If check fail, the node is NonLeaf and resType not template,
// maybe need get/set at its leaf child node.
func (n *Node) AllowResource(resType string) bool {
	if n.IsLeaf() {
		return true
	}
	if len(resType) > len(template) &&
		string(resType[:len(template)]) == template {
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

// WalkFunc is the type of the function for each node visited by Walk.
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

func (n *Node) leafNs() ([]string, error) {
	nsMap, err := n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		result := map[string]string{}
		if node.Type == Leaf {
			result[node.Name] = ""
		} else {
			for chindNs := range childReturn {
				result[chindNs+nodeDeli+node.Name] = ""
			}
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return getKeysOfMap(nsMap), nil
}

// getLeafChild return the leaf id list of the Node.
func (n *Node) leafID() ([]string, error) {
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
		return nil, ErrNoLeafChild
	}
	return getKeysOfMap(IDMap), nil
}

// leafMachineReg return the ns-MachineReg Map.
func (n *Node) leafMachineReg() (map[string]string, error) {
	return n.Walk(func(node *Node, childReturn map[string]string) (map[string]string, error) {
		result := map[string]string{}
		if node.Type == Leaf {
			result[node.Name] = node.MachineReg
		} else {
			for relativeNs, reg := range childReturn {
				result[relativeNs+nodeDeli+node.Name] = reg
			}
		}
		return result, nil
	})
}

// GetById return exact node and ns which with nodeid.
func (n *Node) GetByID(nodeId string) (*Node, string, error) {
	if n.ID == nodeId {
		return n, n.Name, nil
	} else {
		for index := range n.Children {
			if detNode, ns, err := n.Children[index].GetByID(nodeId); err == nil {
				return detNode, ns + nodeDeli + n.Name, nil
			}
		}
	}
	return nil, "", ErrNodeNotFound
}

// GetByName return exact node by nodename.
func (n *Node) GetByNs(ns string) (*Node, error) {
	nsSplit := strings.Split(ns, nodeDeli)
	if len(nsSplit) == 1 && ns == rootNode {
		// return tree if get root.
		return n, nil
	} else if len(nsSplit) < 2 {
		// the query is invalid.
		return nil, ErrNodeNotFound
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

	if rootNode != nsSplit[len(nsSplit)-1] {
		return nil, ErrNodeNotFound
	}

	checkNode := n
	var ok bool
	// Seach each part of the ns, finally get the node of the ns.
	for index := range nsSplit {
		checkNode, ok = checkChild(checkNode, nsSplit[0:len(nsSplit)-index])
		// Return error if not match.
		if !ok {
			return nil, ErrNodeNotFound
		}
		// If each part of the ns is match, return.
		if index+1 == len(nsSplit)-1 {
			break
		}
	}
	return checkNode, nil
}

type nodeCache struct {
	Cache  *map[string]string
	RWsync *sync.RWMutex
}

func (i *nodeCache) Get(name string) (string, bool) {
	i.RWsync.RLock()
	defer i.RWsync.RUnlock()
	v, ok := (*i.Cache)[name]
	return v, ok
}

func (i *nodeCache) Set(name, v string) {
	i.RWsync.Lock()
	defer i.RWsync.Unlock()
	(*i.Cache)[name] = v
}

func (i *nodeCache) Purge() {
	i.RWsync.Lock()
	defer i.RWsync.Unlock()
	for k := range *i.Cache {
		delete((*i.Cache), k)
	}
}
