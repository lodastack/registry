package node

import (
	"encoding/json"
	"sync"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
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

	nsIDCache nodeCache
	nsNSCache nodeCache
	RWsync    *sync.RWMutex

	logger *log.Logger
}

func NewTree(cluster Cluster) (*Tree, error) {
	t := Tree{
		Nodes:     &Node{NodeProperty{ID: rootID, Name: rootNode, Type: NonLeaf, MachineReg: NoMachineMatch}, []*Node{}},
		Cluster:   cluster,
		RWsync:    &sync.RWMutex{},
		nsIDCache: nodeCache{&map[string]string{}, &sync.RWMutex{}},
		nsNSCache: nodeCache{&map[string]string{}, &sync.RWMutex{}},
		logger:    log.New("INFO", "tree", model.LogBackend),
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
		if _, err := t.NewNode(poolNode, rootID, Leaf, NoMachineMatch); err != nil {
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
	t.nsIDCache.Purge()
	t.nsNSCache.Purge()
	return t.Cluster.Update([]byte(rootNode), []byte(nodeDataKey), treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeId string) error {
	return t.Cluster.CreateBucket([]byte(nodeId))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) resByteByNodeID(bucketId, resType string) ([]byte, error) {
	return t.Cluster.View([]byte(bucketId), []byte(resType))
}

// Get []byte of allnodes.
func (t *Tree) allNodeByte() ([]byte, error) {
	return t.resByteByNodeID(rootNode, nodeDataKey)
}

// Set resource to node bucket.
func (t *Tree) setResourceByNodeID(nodeId, resType string, resByte []byte) error {
	return t.Cluster.Update([]byte(nodeId), []byte(resType), resByte)
}

// Append one resource to ns.
func (t *Tree) appendResourceByNodeID(nodeId, resType string, appendRes model.Resource) (string, error) {
	resOld, err := t.resByteByNodeID(nodeId, resType)
	if err != nil {
		t.logger.Error("resByteOfNode error, resOld: ", resOld, resOld, ", error:", err.Error())
		return "", err
	}
	resByte, UUID, err := model.AppendResources(resOld, appendRes)
	if err != nil {
		t.logger.Errorf("AppendResources error, resOld: %s, appendRes: %+v, error: %s", resOld, appendRes, err.Error())
		return "", err
	}
	err = t.setResourceByNodeID(nodeId, resType, resByte)
	return UUID, err
}

func (t *Tree) templateOfNode(nodeId string) (map[string][]byte, error) {
	return t.Cluster.ViewPrefix([]byte(nodeId), []byte(template))
}

// GetAllNodes return the whole nodes.
func (t *Tree) AllNodes() (*Node, error) {
	v, err := t.allNodeByte()
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
	nodes, err := t.AllNodes()
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
	nodes, err := t.AllNodes()
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

// Return leaf node of the ns.
func (t *Tree) Leaf(ns string, format string) ([]string, error) {
	var childNsList []string
	nodes, err := t.GetNodeByNs(ns)
	if err != nil {
		return nil, err
	}
	switch format {
	case IDFormat:
		childNsList, err = nodes.leafID()
	case NsFormat:
		childNsList, err = nodes.leafNs()
		if err != nil {
			return nil, err
		}
	default:
		err = ErrInvalidParam
	}
	return childNsList, err
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// First property argument is used as the machineReg.
// TODO: Permission Check
func (t *Tree) NewNode(name, parentId string, nodeType int, property ...string) (string, error) {
	var nodeId, matchReg string
	var newNode Node
	if nodeType == Root {
		nodeId, name, nodeType, matchReg = rootID, rootNode, NonLeaf, NoMachineMatch
		parentId = "-"
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
	if name != poolNode {
		t.RWsync.Lock()
		defer t.RWsync.Unlock()
	}
	if parentId == "-" {
		// use the node as root node.
		nodes = &newNode
	} else {
		var parentNs string
		// append the node to the child node of its parent node.
		nodes, err = t.AllNodes()
		if err != nil {
			t.logger.Errorf("get all nodes error, parent id: %s, error: %s", parentId, err.Error())
			return "", err
		}
		parent, parentNs, err = nodes.GetByID(parentId)
		if err != nil {
			t.logger.Error("get parent id error:", parentId)
			return "", ErrGetParent
		}

		// If the newnode alread exist, return err.
		if exist := nodes.Exist(newNode.Name + nodeDeli + parentNs); exist {
			return "", ErrNodeAlreadyExist
		}
		if parent.IsLeaf() {
			t.logger.Error("cannot create node under leaf, leaf nodeid:", parentId)
			return "", ErrCreateNodeUnderLeaf
		}

		parent.Children = append(parent.Children, &newNode)
	}
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

	// TODO: rollback if copy template fail
	if parentId == "-" {
		for k, res := range model.RootTemplate {
			resByte := []byte{}
			if res != nil {
				if resByte, err = json.Marshal(res); err != nil {
					panic("create root template fail: " + err.Error())
				}
			}
			if err := t.SetResourceByNs(rootNode, k, resByte); err != nil {
				t.logger.Errorf("SetResourceByNs fail when create rootNode, error: %s", err.Error())
			}
		}
	} else {
		// Read the template of parent node.
		templateRes, err := t.templateOfNode(parentId)
		if err != nil {
			return "nil", err
		}
		for k, resStore := range templateRes {
			if nodeType == Leaf {
				k = k[len(template):]
			}
			if err = t.setResourceByNodeID(nodeId, k, resStore); err != nil {
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

// TODO
func (t *Tree) RmOneResource(ns, resType, resID string) {}

// TODO
func (t *Tree) RmResByMap(nsResIDMap map[string]string, resType string) {}
