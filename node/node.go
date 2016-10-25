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

type nodeIdMap struct {
	Cache  map[string]string
	RWsync *sync.RWMutex
}

func (i *nodeIdMap) Get(id string) (string, bool) {
	i.RWsync.RLock()
	defer i.RWsync.RUnlock()
	name, ok := i.Cache[id]
	return name, ok
}

func (i *nodeIdMap) Set(id, name string) {
	i.RWsync.Lock()
	defer i.RWsync.Unlock()
	i.Cache[id] = name
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
		// TODO: ffjson
		initByte, err := json.Marshal(Node{NodeProperty{ID: "0", Name: rootNode, Type: NonLeaf, MachineReg: "*"}, []Node{}})
		if err != nil {
			return ErrInitNodeKey
		}
		if err = t.Cluster.Update([]byte(nodeBucket), []byte(nodeDataKey), initByte); err != nil {
			return ErrInitNodeKey
		}
	case nodeIdKey:
		// Initialization NodeId Map to store.
		// TODO: ffjson
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

func (t *Tree) GetAllNodes() (*Node, error) {
	v, err := t.Cluster.View([]byte(nodeBucket), []byte(nodeDataKey))
	if err != nil || len(v) == 0 {
		t.logger.Error("get nodeData fail:", err, string(v))
		return nil, ErrGetNode
	}

	nodeData := Node{}
	// TODO: ffjson
	if err := json.Unmarshal(v, &nodeData); err != nil {
		t.logger.Error("GetAllNodes fail:", v)
		t.logger.Error("unmarshal byte to node fail:", err, string(v))
		return nil, ErrGetNode
	}
	return &nodeData, nil
}

// GetNodesById return exact node by nodeid.
func (t *Tree) GetNodesByID(id string) (*Node, error) {
	// TODO: use nodeidKey as cache
	nodes, err := t.GetAllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, err
	}
	return nodes.GetByID(id)
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// TODO:
// 1. Create bucket
// 2. copy template
// 3. Permission Check
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
	return nodeId, t.saveTree()
}

// GetResource return the ResourceType resource belong to the node with NodeId.
// TODO: Permission Check
func (t *Tree) GetNsResource(NodeId string, ResourceType string) (*model.Resources, error) {
	resByte, err := t.Cluster.View([]byte(NodeId), []byte(ResourceType))
	if err != nil {
		return nil, err
	}
	resources, err := model.NewResources(resByte)
	if err != nil {
	}

	return resources, nil
}

// Return nodeid of child nodes of a node.
// If leaf is true, GetChild return nodeid of all leaf children node.
// If leaf is false, GetChild return nodeid of non-leaf childtre node.
func (t *Tree) GetChild(nodeId string, leaf bool) []string {
	// TODO
	return []string{string(nodeId)}
}

func (t *Tree) saveTree() error {
	// TODO: ffjson
	treeByte, _ := json.Marshal(*t.Nodes)
	return t.Cluster.Update([]byte(nodeBucket), []byte(nodeDataKey), treeByte)
}

// Return NodeId if pretty is id, else return resource data.
// Params contain: nodeid/resource key/resource value
func (t *Tree) SearchResource(NodeId string, ResourceType string, pretty string, params ...string) {
}
