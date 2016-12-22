package node

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
)

var (
	template    string = model.TemplatePrefix
	ErrEmptyRes        = model.ErrEmptyRes
)

const (
	nodeDataKey = "node"
	nodeIdKey   = "nodeid"
	rootNode    = "loda"
	rootID      = "0"
	nodeDeli    = "."
	poolNode    = "pool"

	NsFormat = "ns"
	IDFormat = "id"

	NoMachineMatch = "^$"
)

type Tree struct {
	Nodes   *Node
	Cluster Cluster

	Cache *nodeCache
	Mu    sync.RWMutex

	logger *log.Logger
}

func NewTree(cluster Cluster) (*Tree, error) {
	t := Tree{
		Nodes:   &Node{NodeProperty{ID: rootID, Name: rootNode, Type: NonLeaf, MachineReg: NoMachineMatch}, []*Node{}},
		Cluster: cluster,
		Mu:      sync.RWMutex{},
		logger:  log.New(config.C.LogConf.Level, "tree", model.LogBackend),
		Cache:   &nodeCache{Data: map[string]string{}},
	}
	err := t.init()
	return &t, err
}

func (t *Tree) init() error {
	err := t.Cluster.CreateBucketIfNotExist([]byte(rootNode))
	if err != nil {
		t.logger.Error("itree CreateBucketIfNotExist fail:", err.Error())
		return err
	}
	if err := t.initKey(nodeDataKey); err != nil {
		t.logger.Error("init nodeDataKey fail:", err.Error())
		return err
	}
	if err := t.initKey(nodeIdKey); err != nil {
		t.logger.Error("init nodeidKey fail:", err.Error())
		return err
	}
	go func() {
		start := time.Now()
		allNodes, err := t.AllNodes()
		if err != nil {
			panic(err)
		}

		initCache, err := allNodes.initNsCache()
		if err != nil {
			panic(err)
		}
		t.Cache.Lock()
		for k, v := range t.Cache.Data {
			initCache.Set(k, v)
		}
		t.Cache.Data = initCache.Data
		t.Cache.Unlock()
		finishCacheInit := time.Now()
		t.logger.Debugf("cache have %d item, init cost :%v",
			t.Cache.Len(),
			finishCacheInit.Sub(start))
	}()
	return nil
}

// initKey initialization tree data and idmap if they are nil.
func (t *Tree) initKey(key string) error {
	v, err := t.Cluster.View([]byte(rootNode), []byte(key))
	if err != nil {
		return err
	}
	if len(v) != 0 {
		return nil
	}

	t.logger.Info(key, "is not inited, begin to init")
	switch key {
	case nodeDataKey:
		// Create rootNode map/bucket and init template.
		if _, err := t.NewNode("", "", Root); err != nil {
			panic("create root node fail: " + err.Error())
		}
		// Create root pool node.
		if _, err := t.NewNode(poolNode, rootNode, Leaf, NoMachineMatch); err != nil {
			panic("create root pool node fail: " + err.Error())
		}
	case nodeIdKey:
		// Initialization NodeId Map to store.
		initByte, _ := json.Marshal(map[string]string{rootID: rootNode})
		if err != nil {
			return ErrInitNodeKey
		}
		if err = t.Cluster.Update([]byte(rootNode), []byte(nodeIdKey), initByte); err != nil {
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
	// TODO: purge cache or not
	return t.Cluster.Update([]byte(rootNode), []byte(nodeDataKey), treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeId string) error {
	return t.Cluster.CreateBucket([]byte(nodeId))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) getByteFromStore(bucketId, resType string) ([]byte, error) {
	return t.Cluster.View([]byte(bucketId), []byte(resType))
}

func (t *Tree) removeNodeFromStore(bucketId string) error {
	return t.Cluster.RemoveBucket([]byte(bucketId))
}

// Get []byte of allnodes.
func (t *Tree) allNodeByte() ([]byte, error) {
	return t.getByteFromStore(rootNode, nodeDataKey)
}

// Set resource to node bucket.
func (t *Tree) setByteToStore(nodeId, resType string, resByte []byte) error {
	return t.Cluster.Update([]byte(nodeId), []byte(resType), resByte)
}

func (t *Tree) templateOfNode(nodeId string) (map[string]string, error) {
	return t.Cluster.ViewPrefix([]byte(nodeId), []byte(template))
}

// GetAllNodes return the whole nodes.
func (t *Tree) AllNodes() (*Node, error) {
	v, err := t.allNodeByte()
	if err != nil || len(v) == 0 {
		t.logger.Errorf("get allNode fail: %v", err)
		return nil, ErrGetNode
	}

	var allNode Node
	if err := allNode.UnmarshalJSON(v); err != nil {
		t.logger.Errorf("GetAllNodes unmarshal byte to node fail: %s", err.Error())
		return nil, ErrGetNode
	}
	return &allNode, nil
}

// getNsByID return id of node with name nodeName.
func (t *Tree) getNs(id string) (string, error) {
	ns, ok := t.Cache.Get(id)
	if ok && ns != "" {
		return ns, nil
	}

	nodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return "", err
	}
	_, ns, err = nodes.GetByID(id)
	if err != nil {
		t.logger.Errorf("get node by ID fail: %s.", err.Error())
		return "", err
	}
	t.Cache.Set(id, ns)
	return ns, nil
}

// GetNodesById return exact node with name.
func (t *Tree) GetNode(ns string) (*Node, error) {
	if ns == "" {
		return nil, ErrInvalidParam
	}
	// TODO: use nodeidKey as cache
	nodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return nil, err
	}
	return nodes.Get(ns)
}

func (t *Tree) getID(ns string) (string, error) {
	id, ok := t.Cache.Get(ns)
	// If cannot get Node from cache, get from tree and set cache.
	if ok && id != "" {
		return id, nil
	}

	node, err := t.GetNode(ns)
	if err != nil {
		t.logger.Errorf("GetNodeByNs fail when get ns:%s, error: %s\n", ns, err.Error)
		return "", err
	}
	t.Cache.Set(ns, node.ID)
	return node.ID, nil
}

// Return leaf node of the ns.
func (t *Tree) LeafChildIDs(ns string) ([]string, error) {
	nodeId, err := t.getID(ns)
	if nodeId == "" || err != nil {
		return nil, err
	}

	// read the tree if not get from cache.
	node, err := t.GetNode(ns)
	if err != nil {
		return nil, err
	}
	if node.Type == Leaf {
		return []string{node.ID}, nil
	}
	return node.leafChildIDs()
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// First property argument is used as the machineReg.
// TODO: Permission Check
func (t *Tree) NewNode(name, parentNs string, nodeType int, property ...string) (string, error) {
	var nodeId, matchReg string
	var newNode Node
	if nodeType == Root {
		nodeId, name, nodeType, matchReg = rootID, rootNode, NonLeaf, NoMachineMatch
		parentNs = "-"
	} else {
		if len(property) > 0 && property[0] != "" {
			matchReg = property[0]
		} else {
			matchReg = NoMachineMatch
		}
		nodeId = common.GenUUID()
	}
	newNode = Node{NodeProperty{ID: nodeId, Name: name, Type: nodeType, MachineReg: matchReg}, []*Node{}}

	var nodes, parent *Node
	var err error
	// Create Pool node not lock the tree, because create leaf will lock the tree.
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if parentNs == "-" {
		// use the node as root node.
		nodes = &newNode
	} else {
		// append the node to the child node of its parent node.
		nodes, err = t.AllNodes()
		if err != nil {
			t.logger.Errorf("get all nodes error, parent ns: %s, error: %s", parentNs, err.Error())
			return "", err
		}
		parent, err = nodes.Get(parentNs)
		if err != nil {
			t.logger.Errorf("get parent id ns: %s, error: %v", parentNs, err)
			return "", ErrGetParent
		}

		// If the newnode alread exist, return err.
		if exist := nodes.Exist(newNode.Name + nodeDeli + parentNs); exist {
			return "", ErrNodeAlreadyExist
		}
		if parent.IsLeaf() {
			t.logger.Error("cannot create node under leaf, leaf node:", parentNs)
			return "", ErrCreateNodeUnderLeaf
		}
		parent.Children = append(parent.Children, &newNode)
		// if not create root/pool node, add node to cache.
		if newNode.Name != rootNode {
			t.Cache.Add(parent.ID, parentNs, &newNode)
		}
	}
	t.Nodes = nodes

	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return "", err
	}
	// Create a now bucket fot this node.
	if err := t.createBucketForNode(nodeId); err != nil {
		t.Cache.Delete(newNode.ID)
		t.logger.Errorf("NewNode createNodeBucket fail, nodeid:%s, error: %s\n", nodeId, err.Error())
		// Delete the new node in tree to rollback.
		parent.Children = parent.Children[:len(parent.Children)-1]
		if err := t.saveTree(); err != nil {
			t.logger.Error("Rollback tree node fail: %s", err.Error())
		}
		return "", err
	}

	// TODO: rollback if copy template fail
	if parentNs == "-" {
		for k, res := range model.RootTemplate {
			if err := t.SetResource(rootNode, k, res); err != nil {
				t.logger.Errorf("SetResourceByNs fail when create rootNode, error: %s", err.Error())
			}
		}
	} else {
		// Read the template of parent node.
		templateRes, err := t.templateOfNode(parent.ID)
		if err != nil {
			return "nil", err
		}
		for k, resStore := range templateRes {
			if nodeType == Leaf {
				k = k[len(template):]
			}
			if err = t.setByteToStore(nodeId, k, []byte(resStore)); err != nil {
				t.logger.Errorf("SetResourceByNs fail when newnode %s, error: %s", nodeId, err.Error())
			}
		}
	}
	return nodeId, nil
}

// Return  pool nodeID if create a pool node.
func (t *Tree) NewPool(parentId, offlineMatch string) (string, error) {
	poolId, ErrCreatePool := t.NewNode(poolNode, parentId, Leaf, offlineMatch)
	if ErrCreatePool == nil {
		return poolId, nil
	} else if ErrCreatePool == ErrNodeAlreadyExist {
		return "", nil
	}
	t.logger.Errorf("Create pool node fail:%s", ErrCreatePool.Error())
	return "", ErrCreatePool
}

func (t *Tree) UpdateNode(ns, name, machineReg string) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	allNodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}
	newNs := strings.Join(append([]string{name}, strings.Split(ns, nodeDeli)[1:]...), nodeDeli)
	if exist := allNodes.Exist(newNs); exist {
		return ErrNodeAlreadyExist
	}

	node, err := allNodes.Get(ns)
	if err != nil {
		t.logger.Errorf("GetByNs %s fail, error: %s", ns, err.Error())
		return err
	}
	node.update(name, machineReg)

	t.Nodes = allNodes
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return err
	}
	t.Cache.Delete(node.ID)
	return nil
}

// DelNode delete node from tree, remove bucket.
func (t *Tree) DelNode(ns string) error {
	nsSplit := strings.Split(ns, nodeDeli)
	if len(nsSplit) < 2 {
		return ErrInvalidParam
	}
	parentNs := strings.Join(nsSplit[1:], nodeDeli)
	delID, err := t.getID(ns)
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}

	// Not allow delete node which still has child node or machine.
	rl, err := t.GetResourceList(ns, "machine")
	if (err != nil && err != ErrNoLeafChild) ||
		(rl != nil && len(*rl) != 0) {
		t.logger.Errorf("not allow delete ns %s, error: %v", ns, err)
		return ErrNotAllowDel
	}

	t.Mu.Lock()
	defer t.Mu.Unlock()
	allNodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}
	parentNode, err := allNodes.Get(parentNs)
	if err != nil {
		t.logger.Errorf("GetByNs fail, error: %s", err.Error())
		return err
	}

	if err := parentNode.delChild(delID); err != nil {
		t.logger.Errorf("delete node fail, parent ns: %s, delete ID: %s, error: %s", parentNs, delID, err.Error())
		return err
	}

	if err := t.removeNodeFromStore(delID); err != nil {
		t.logger.Errorf("delete node from store fail, parent ns: %s, delete ID: %s, error: %s", parentNs, delID, err.Error())
		return err
	}
	t.logger.Infof("remove node (ID: %s) behind ns %s from store success: %s", delID, parentNs)
	t.Nodes = allNodes
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return err
	}
	t.Cache.Delete(delID)
	return nil
}
