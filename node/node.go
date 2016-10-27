package node

import (
	"encoding/json"
	"errors"
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
)

var (
	ErrInitNodeBucket      = errors.New("init node bucket fail")
	ErrInitNodeKey         = errors.New("init node bucket k-v fail")
	ErrGetNode             = errors.New("get node fail")
	ErrNodeNotFound        = errors.New("node not found")
	ErrGetParent           = errors.New("get parent node error")
	ErrCreateNodeUnderLeaf = errors.New("can not create node under leaf node")
	ErrGetNodeID           = errors.New("get nodeid fail")
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
	Clildren []Node
}

func (n *Node) IsLeaf() bool {
	return n.Type == Leaf
}

// GetById return exact node by nodeid.
func (n *Node) GetByID(nodeId string) (*Node, error) {
	if n.ID == nodeId {
		return n, nil
	} else {
		for index := range n.Clildren {
			if detNode, err := n.Clildren[index].GetByID(nodeId); err == nil {
				return detNode, nil
			}
		}
	}
	return nil, ErrNodeNotFound
}

// GetByName return exact node by nodename.
func (n *Node) GetByName(nodeName string) (*Node, error) {
	if n.Name == nodeName {
		return n, nil
	} else {
		for index := range n.Clildren {
			if detNode, err := n.Clildren[index].GetByName(nodeName); err == nil {
				return detNode, nil
			}
		}
	}
	return nil, ErrNodeNotFound
}

type nodeIdMap struct {
	Cache  map[string]string
	RWsync *sync.RWMutex
}

func (i *nodeIdMap) Get(name string) (string, bool) {
	i.RWsync.RLock()
	defer i.RWsync.RUnlock()
	id, ok := i.Cache[name]
	return id, ok
}

func (i *nodeIdMap) Set(name, id string) {
	i.RWsync.Lock()
	defer i.RWsync.Unlock()
	i.Cache[name] = id
}

type Tree struct {
	Nodes   *Node
	Cluster Cluster

	idMap  nodeIdMap
	RWsync *sync.RWMutex

	logger *log.Logger
}

func NewTree(cluster Cluster) (*Tree, error) {
	t := Tree{
		Nodes:   &Node{NodeProperty{ID: "0", Name: rootNode, Type: NonLeaf, MachineReg: "*"}, []Node{}},
		Cluster: cluster,
		RWsync:  &sync.RWMutex{},
		idMap:   nodeIdMap{map[string]string{}, &sync.RWMutex{}},
		logger:  log.New("INFO", "tree", model.LogBackend),
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
		initNodes := Node{NodeProperty{ID: "0", Name: rootNode, Type: NonLeaf, MachineReg: "*"}, []Node{}}
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
	treeByte, _ := t.Nodes.MarshalJSON()
	return t.Cluster.Update([]byte(nodeBucket), []byte(nodeDataKey), treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeId string) error {
	return t.Cluster.CreateBucket([]byte(nodeId))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) getResByteOfNode(bucketId, resType []byte) ([]byte, error) {
	return t.Cluster.View([]byte(bucketId), []byte(resType))
}

// Get []byte of allnodes.
func (t *Tree) getAllNodeByte() ([]byte, error) {
	return t.getResByteOfNode([]byte(nodeBucket), []byte(nodeDataKey))
}

func (t *Tree) setResourceByNodeID(nodeId, resType string, resByte []byte) error {
	return t.Cluster.Update([]byte(nodeId), []byte(resType), resByte)
}

// getNodeIDByName return id of node with name nodeName.
// NOTE: if two nodes have the same name, will return the first one it find.
func (t *Tree) getNodeIDByName(nodeName string) string {
	NodeId, ok := t.idMap.Get(nodeName)
	// If cannot get Node from cache, get from tree and set tree cache.
	if !ok {
		if node, err := t.GetNodeByName(nodeName); err != nil {
			return ""
		} else {
			NodeId = node.ID
		}
		t.idMap.Set(nodeName, NodeId)
	}
	return NodeId
}

// GetAllNodes return the whole nodes.
func (t *Tree) GetAllNodes() (*Node, error) {
	v, err := t.getAllNodeByte()
	if err != nil || len(v) == 0 {
		t.logger.Error("get allNode fail:", err, string(v))
		return nil, ErrGetNode
	}

	allNode := Node{}
	if err := allNode.UnmarshalJSON(v); err != nil {
		t.logger.Errorf("GetAllNodes unmarshal byte to node fail: %s\n", err)
		return nil, ErrGetNode
	}
	return &allNode, nil
}

// GetNodesById return exact node with nodeid.
func (t *Tree) GetNodeByID(id string) (*Node, error) {
	// TODO: use nodeidKey as cache
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, err
	}
	return nodes.GetByID(id)
}

// GetNodesById return exact node with name.
func (t *Tree) GetNodeByName(name string) (*Node, error) {
	// TODO: use nodeidKey as cache
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, err
	}
	return nodes.GetByName(name)
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// TODO:
// 1. copy template
// 2. Permission Check
func (t *Tree) NewNode(name, parentId string, nodeType int, property ...string) (string, error) {
	nodeId := common.GenUUID()
	newNode := Node{NodeProperty{ID: nodeId, Name: name, Type: nodeType, MachineReg: "-"}, []Node{}}

	t.RWsync.Lock()
	defer t.RWsync.Unlock()
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error:", parentId)
		return "", err
	}
	parent, err := nodes.GetByID(parentId)
	if err != nil {
		t.logger.Error("get parent id error:", parentId)
		return "", ErrGetParent
	}
	if parent.IsLeaf() {
		t.logger.Error("cannot create node under leaf, leaf nodeid:", parentId)
		return "", ErrCreateNodeUnderLeaf
	}

	parent.Clildren = append(parent.Clildren, newNode)
	t.Nodes = nodes

	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail")
		return "", err
	}
	// Create a now bucket fot this node.
	if err := t.createBucketForNode(nodeId); err != nil {
		t.logger.Errorf("NewNode createNodeBucket fail, nodeid:%s\n", nodeId)
		// Delete node to rollback tree
		parent.Clildren = parent.Clildren[:len(parent.Clildren)-1]
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
	resByte, err := t.getResByteOfNode([]byte(NodeId), []byte(ResourceType))
	if err != nil {
		return nil, err
	}
	resources := &model.Resources{}
	err = resources.Unmarshal(resByte)
	if err != nil {
		t.logger.Error("GetResourceByNodeID fail:", err, string(resByte))
	}
	return resources, nil
}

// GetResource return the ResourceType resource belong to the node with NodeName.
// TODO: Permission Check
func (t *Tree) GetResourceByNodeName(nodeName string, ResourceType string) (*model.Resources, error) {
	nodeId := t.getNodeIDByName(nodeName)
	if nodeId == "" {
		t.logger.Error("GetResourceByNodeName GetNodeByName fail: %s \n", nodeName)
		return nil, ErrGetNodeID
	}
	return t.GetResourceByNodeID(nodeId, ResourceType)
}

func (t *Tree) SetResourceByNodeID(nodeId, resType string, ResByte []byte) error {
	var err error
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
	return t.setResourceByNodeID(nodeId, resType, resStore)
}

func (t *Tree) SetResourceByNodeName(nodeName, resType string, ResByte []byte) error {
	nodeId := t.getNodeIDByName(nodeName)
	if nodeId == "" {
		t.logger.Error("GetResourceByNodeName GetNodeByName fail: %s \n", nodeName)
		return ErrGetNodeID
	}
	return t.SetResourceByNodeID(nodeId, resType, ResByte)
}

// Return nodeid of child nodes of a node.
// If leaf is true, GetChild return nodeid of all leaf children node.
// If leaf is false, GetChild return nodeid of non-leaf childtre node.
func (t *Tree) GetChild(nodeId string, leaf bool) []string {
	// TODO
	return []string{string(nodeId)}
}

// Return NodeId if pretty is id, else return resource data.
// Params contain: nodeid/resource key/resource value
func (t *Tree) SearchResource(NodeId string, ResourceType string, pretty string, params ...string) {
}
