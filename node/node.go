package node

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
)

const (
	Leaf    = iota // leaf type of node
	NonLeaf        // non-leaf type of node

	nodeBucket  = "loda"
	nodeDataKey = "node"
	nodeIdKey   = "nodeid"
	rootNode    = "loda"
	nodeDeli    = "."

	NsFormat = "ns"
	IDFormat = "id"
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

	ErrHappen = errors.New("error happen")
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

// getLeafChild return the leaf ns list of the Node.
func (n *Node) getLeafNs() ([]string, error) {
	childNs := []string{}
	for index := range n.Children {
		if n.Children[index].Type == Leaf {
			childNs = append(childNs, n.Children[index].Name)
		} else {
			if childLeaf, err := n.Children[index].getLeafNs(); err != nil {
				return nil, err
			} else {
				for childIndex := range childLeaf {
					childNs = append(childNs, childLeaf[childIndex]+nodeDeli+n.Children[index].Name)
				}
			}
		}
	}
	return childNs, nil
}

// getLeafChild return the leaf id list of the Node.
func (n *Node) getLeafID() ([]string, error) {
	childNs := []string{}
	for index := range n.Children {
		if n.Children[index].Type == Leaf {
			childNs = append(childNs, n.Children[index].ID)
		} else {
			if childLeaf, err := n.Children[index].getLeafID(); err != nil {
				return nil, err
			} else {
				childNs = append(childNs, childLeaf...)
			}
		}
	}
	return childNs, nil
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
		return nil, ErrInvalidParam
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

type Tree struct {
	Nodes   *Node
	Cluster Cluster

	nsIDCache nodeCache
	nsNSCache nodeCache
	RWsync    *sync.RWMutex

	logger *log.Logger
}

func NewTree(cluster Cluster) (*Tree, error) {
	t := Tree{
		Nodes:     &Node{NodeProperty{ID: "0", Name: rootNode, Type: NonLeaf, MachineReg: "*"}, []*Node{}},
		Cluster:   cluster,
		RWsync:    &sync.RWMutex{},
		nsIDCache: nodeCache{&map[string]string{}, &sync.RWMutex{}},
		nsNSCache: nodeCache{&map[string]string{}, &sync.RWMutex{}},
		logger:    log.New("INFO", "tree", model.LogBackend),
	}
	err := t.Cluster.CreateBucketIfNotExist([]byte(nodeBucket))
	if err != nil {
		t.logger.Error("itree CreateBucketIfNotExist fail:", err.Error())
		return nil, err
	}

	if err := t.initIfNotExist(nodeDataKey); err != nil {
		t.logger.Error("init nodeDataKey fail:", err.Error())
		return nil, err
	}
	if err := t.initIfNotExist(nodeIdKey); err != nil {
		t.logger.Error("init nodeidKey fail:", err.Error())
		return nil, err
	}
	return &t, nil
}

// initIfNotExist initialization tree data and idmap if they are nil.
func (t *Tree) initIfNotExist(key string) error {
	v, err := t.Cluster.View([]byte(nodeBucket), []byte(key))
	if err != nil {
		return err
	}
	if len(v) != 0 {
		return nil
	}

	t.logger.Info(key, "is not inited, begin to init")
	switch key {
	case nodeDataKey:
		// Initialization node Data to store.
		initNodes := Node{NodeProperty{ID: "0", Name: rootNode, Type: NonLeaf, MachineReg: "*"}, []*Node{}}
		initByte, err := initNodes.MarshalJSON()
		if err != nil {
			return ErrInitNodeKey
		}
		if err = t.Cluster.Update([]byte(nodeBucket), []byte(nodeDataKey), initByte); err != nil {
			return ErrInitNodeKey
		}
	case nodeIdKey:
		// Initialization NodeId Map to store.
		initByte, _ := json.Marshal(map[string]string{"0": rootNode})
		if err != nil {
			return ErrInitNodeKey
		}
		if err = t.Cluster.Update([]byte(nodeBucket), []byte(nodeIdKey), initByte); err != nil {
			return ErrInitNodeKey
		}
	}
	return nil
}

// Save Nodes to store.
func (t *Tree) saveTree() error {
	treeByte, err := t.Nodes.MarshalJSON()
	if err != nil {
		t.logger.Errorf("Tree save fail: %s\n", err.Error())
		return err
	}
	t.nsIDCache.Purge()
	t.nsNSCache.Purge()
	return t.Cluster.Update([]byte(nodeBucket), []byte(nodeDataKey), treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeId string) error {
	return t.Cluster.CreateBucket([]byte(nodeId))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) getResByteOfNode(bucketId, resType string) ([]byte, error) {
	return t.Cluster.View([]byte(bucketId), []byte(resType))
}

// Get []byte of allnodes.
func (t *Tree) getAllNodeByte() ([]byte, error) {
	return t.getResByteOfNode(nodeBucket, nodeDataKey)
}

// Set resource to node bucket.
func (t *Tree) setResourceByNodeID(nodeId, resType string, resByte []byte) error {
	return t.Cluster.Update([]byte(nodeId), []byte(resType), resByte)
}

// GetAllNodes return the whole nodes.
func (t *Tree) GetAllNodes() (*Node, error) {
	v, err := t.getAllNodeByte()
	if err != nil || len(v) == 0 {
		t.logger.Errorf("get allNode fail: %s", err.Error())
		return nil, ErrGetNode
	}

	var allNode Node
	if err := allNode.UnmarshalJSON(v); err != nil {
		t.logger.Errorf("GetAllNodes unmarshal byte to node fail: %s", err.Error())
		return nil, ErrGetNode
	}
	return &allNode, nil
}

// GetNodesById return node and its ns which have the nodeid.
func (t *Tree) GetNodeByID(id string) (*Node, string, error) {
	// Return GetNodeByNs if read ns from cache, becasue GetNodeByNs donot need read the whole tree.
	// Update tree will purge cache, so cache can be trust.
	NodeNs, ok := t.nsIDCache.Get(id)
	if ok {
		node, err := t.GetNodeByNs(NodeNs)
		return node, NodeNs, err
	}
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, "", err
	}
	node, ns, err := nodes.GetByID(id)

	if _, ok := t.nsIDCache.Get(id); !ok {
		t.nsIDCache.Set(id, ns)
	}

	return node, ns, err
}

// getNsByID return id of node with name nodeName.
func (t *Tree) getNsByID(id string) (string, error) {
	var err error
	ns, ok := t.nsIDCache.Get(id)

	if !ok {
		if _, ns, err = t.GetNodeByID(id); err != nil {
			t.logger.Errorf("GetNodeByID fail when get id:%s, error: %s\n", id, err.Error)
			return "", err
		}

	}
	return ns, nil
}

// GetNodesById return exact node with name.
func (t *Tree) GetNodeByNs(ns string) (*Node, error) {
	// TODO: use nodeidKey as cache
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, err
	}
	return nodes.GetByNs(ns)
}

func (t *Tree) getIDByNs(ns string) (string, error) {
	id, ok := t.nsNSCache.Get(ns)
	// If cannot get Node from cache, get from tree and set cache.
	if !ok {
		node, err := t.GetNodeByNs(ns)
		if err != nil {
			t.logger.Errorf("GetNodeByNs fail when get ns:%s, error: %s\n", ns, err.Error)
			return "", err
		}
		id = node.ID
		t.nsNSCache.Set(ns, node.ID)
	}
	return id, nil
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// TODO:
// 1. copy template
// 2. Permission Check
func (t *Tree) NewNode(name, parentId string, nodeType int, property ...string) (string, error) {
	nodeId := common.GenUUID()
	newNode := Node{NodeProperty{ID: nodeId, Name: name, Type: nodeType, MachineReg: "-"}, []*Node{}}

	t.RWsync.Lock()
	defer t.RWsync.Unlock()
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Errorf("get all nodes error, parent id: %s, error: %s", parentId, err.Error())
		return "", err
	}
	parent, _, err := nodes.GetByID(parentId)
	if err != nil {
		t.logger.Error("get parent id error:", parentId)
		return "", ErrGetParent
	}
	if parent.IsLeaf() {
		t.logger.Error("cannot create node under leaf, leaf nodeid:", parentId)
		return "", ErrCreateNodeUnderLeaf
	}

	parent.Children = append(parent.Children, &newNode)
	t.Nodes = nodes

	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return "", err
	}
	// Create a now bucket fot this node.
	if err := t.createBucketForNode(nodeId); err != nil {
		t.logger.Errorf("NewNode createNodeBucket fail, nodeid:%s\n", nodeId)
		// Delete the new node in tree to rollback.
		parent.Children = parent.Children[:len(parent.Children)-1]
		if err := t.saveTree(); err != nil {
			t.logger.Error("Rollback tree node fail!")
		}
		return "", err
	}
	return nodeId, nil
}

// GetResource return the ResourceType resource belong to the node with NodeId.
// TODO: Permission Check
func (t *Tree) GetResourceByNodeID(NodeId string, ResourceType string) (*model.Resources, error) {
	resByte, err := t.getResByteOfNode(NodeId, ResourceType)
	if err != nil {
		return nil, err
	}
	resources := new(model.Resources)
	err = resources.Unmarshal(resByte)
	if err != nil {
		t.logger.Error("GetResourceByNodeID fail:", err, string(resByte))
	}
	return resources, nil
}

// GetResource return the ResourceType resource belong to the node with NodeName.
// TODO: Permission Check
func (t *Tree) GetResourceByNs(ns string, ResourceType string) (*model.Resources, error) {
	nodeId, err := t.getIDByNs(ns)
	if nodeId == "" || err != nil {
		t.logger.Error("GetResourceByNodeName GetNodeByNs fail, ns %s, error:%+v \n", ns, err)
		return nil, ErrGetNodeID
	}
	return t.GetResourceByNodeID(nodeId, ResourceType)
}

func (t *Tree) SetResourceByNodeID(nodeId, resType string, ResByte []byte) error {
	ns, err := t.getNsByID(nodeId)
	if err != nil {
		return err
	}
	return t.SetResourceByNs(ns, resType, ResByte)
}

func (t *Tree) SetResourceByNs(ns, resType string, ResByte []byte) error {
	node, err := t.GetNodeByNs(ns)
	if err != nil || node.ID == "" {
		t.logger.Error("Get node by ns(%s) fail\n", ns)
		return ErrGetNode
	}
	if !node.IsLeaf() {
		return ErrSetResourceToLeaf
	}

	resesStruct, err := model.NewResources(ResByte)
	if err != nil {
		t.logger.Errorf("set resource to node fail, unmarshal resource fail: %s\n", err)
		return err
	}
	var resStore []byte
	resStore, err = resesStruct.Marshal()
	if err != nil {
		t.logger.Errorf("set resource to node fail, marshal resource to byte fail: %s\n", err)
		return err
	}

	return t.setResourceByNodeID(node.ID, resType, resStore)
}

func (t *Tree) SearchResourceByNs(ns, resType string, search model.ResourceSearch) (map[string]*model.Resources, error) {
	leafIDs, err := t.GetLeaf(ns, IDFormat)
	if err != nil {
		return nil, err
	} else if len(leafIDs) == 0 {
		return nil, ErrNilChildNode
	}

	result := map[string]*model.Resources{}

	for _, leafID := range leafIDs {
		resByte, err := t.getResByteOfNode(leafID, resType)
		if err != nil {
			return nil, err
		}

		search.Init()
		if resOfOneNs, err := search.Process(resByte); err != nil {
			return nil, err
		} else if len(resOfOneNs) != 0 {
			ns, err := t.getNsByID(leafID)
			if err != nil {
				return nil, err
			}
			result[ns] = &model.Resources{}
			result[ns].AppendResources(resOfOneNs)
		}
	}

	return result, nil
}

// Return leaf node of the ns.
func (t *Tree) GetLeaf(ns string, format string) ([]string, error) {
	nodes, err := t.GetNodeByNs(ns)
	if err != nil {
		return nil, err
	}
	switch format {
	case IDFormat:
		childNsList, err := nodes.getLeafID()
		if err != nil {
			return nil, err
		}
		return childNsList, nil
	default:
		childNsList, err := nodes.getLeafNs()
		if err != nil {
			return nil, err
		}
		// add the search ns to child relative ns.
		for index := range childNsList {
			childNsList[index] = childNsList[index] + nodeDeli + ns
		}
		return childNsList, nil
	}
	return nil, ErrHappen
}

// Return NodeId if pretty is id, else return resource data.
// Params contain: nodeid/resource key/resource value
func (t *Tree) SearchResource(NodeId string, ResourceType string, pretty string, params ...string) {
}
