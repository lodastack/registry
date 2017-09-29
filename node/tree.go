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
	"github.com/lodastack/registry/node/cluster"
	n "github.com/lodastack/registry/node/node"

	m "github.com/lodastack/models"
)

var (
	template    string = model.TemplatePrefix
	ErrEmptyRes        = model.ErrEmptyRes
)

const (
	nodeDataKey = "node"
	nodeIdKey   = "nodeid"

	nodeBucket   = "loda"
	reportBucket = "report"

	rootNode = "loda"
	poolNode = "pool"
	rootID   = "0"
	nodeDeli = "."

	NsFormat = "ns"
	IDFormat = "id"

	NoMachineMatch = "^$"
)

type Tree struct {
	Nodes *n.Node
	c     cluster.ClusterInf
	n     n.NodeInf

	Mu sync.RWMutex

	reports ReportInfo
	logger  *log.Logger
}

func NewTree(cluster cluster.ClusterInf) (*Tree, error) {
	t := Tree{
		Nodes:   &n.Node{n.NodeProperty{ID: rootID, Name: rootNode, Type: n.NonLeaf, MachineReg: NoMachineMatch}, []*n.Node{}},
		c:       cluster,
		n:       n.NewNodeMethod(cluster),
		Mu:      sync.RWMutex{},
		logger:  log.New(config.C.LogConf.Level, "tree", model.LogBackend),
		reports: ReportInfo{sync.RWMutex{}, make(map[string]m.Report)},
	}
	err := t.init()
	return &t, err
}

func (t *Tree) init() error {
	if err := t.initNodeBucket(); err != nil {
		return err
	}
	return t.initReportBucket()
}

func (t *Tree) initNodeBucket() error {
	err := t.c.CreateBucketIfNotExist([]byte(nodeBucket))
	if err != nil {
		t.logger.Errorf("tree %s CreateBucketIfNotExist fail: %s", nodeBucket, err.Error())
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

func (t *Tree) initReportBucket() error {
	err := t.c.CreateBucketIfNotExist([]byte(reportBucket))
	if err != nil {
		t.logger.Errorf("tree init %s CreateBucketIfNotExist fail: %s", reportBucket, err.Error())
		return err
	}
	t.reports.ReportInfo, err = t.readReport()
	if err != nil {
		t.logger.Error("tree init report fail, set empty")
		t.reports.ReportInfo = make(map[string]m.Report)
	}

	// Persistent report data.
	go func() {
		if config.C.CommonConf.PersistReport <= 0 {
			return
		}
		c := time.Tick(time.Duration(config.C.CommonConf.PersistReport) * time.Hour)
		for {
			select {
			case <-c:
				t.reports.Lock()
				t.setReport(t.reports.ReportInfo)
				t.reports.Unlock()
			}
		}
	}()

	// Update machine status based on the report。
	go func() {
		c := time.Tick(time.Hour)
		for {
			select {
			case <-c:
				reports := t.GetReportInfo()
				if err := t.UpdateMachineStatus(reports); err != nil {
					t.logger.Error("UpdateMachineStatus fail:", err.Error())
				}
			}
		}
	}()
	return nil
}

// initKey initialization tree data and idmap if they are nil.
func (t *Tree) initKey(key string) error {
	v, err := t.c.View([]byte(rootNode), []byte(key))
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
		if _, err := t.NewNode("", "", n.Root); err != nil {
			panic("create root node fail: " + err.Error())
		}
		// Create root pool node.
		if _, err := t.NewNode(poolNode, rootNode, n.Leaf, NoMachineMatch); err != nil {
			panic("create root pool node fail: " + err.Error())
		}
	case nodeIdKey:
		// Initialization NodeId Map to store.
		initByte, _ := json.Marshal(map[string]string{rootID: rootNode})
		if err != nil {
			return common.ErrInitNodeKey
		}
		if err = t.c.Update([]byte(rootNode), []byte(nodeIdKey), initByte); err != nil {
			return common.ErrInitNodeKey
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
	return t.c.Update([]byte(rootNode), []byte(nodeDataKey), treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeId string) error {
	return t.c.CreateBucket([]byte(nodeId))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) getByteFromStore(bucketId, resType string) ([]byte, error) {
	return t.c.View([]byte(bucketId), []byte(resType))
}

func (t *Tree) removeNodeFromStore(bucketId string) error {
	return t.c.RemoveBucket([]byte(bucketId))
}

// Set resource to node bucket.
func (t *Tree) setByteToStore(nodeId, resType string, resByte []byte) error {
	return t.c.Update([]byte(nodeId), []byte(resType), resByte)
}

func (t *Tree) templateOfNode(nodeId string) (map[string][]byte, error) {
	return t.c.ViewPrefix([]byte(nodeId), []byte(template))
}

// GetAllNodes return the whole nodes.
func (t *Tree) AllNodes() (n *n.Node, err error) {
	if n, err = t.n.AllNodes(); err != nil {
		t.logger.Errorf("AllNodes fail, node %v, error: %s", *n, err.Error())
	}
	return
}

// GetNode return exact node by node NS.
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

// getNsByID return id of node with name nodeName.
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

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// First property argument is used as the machineReg.
// TODO: Permission Check
func (t *Tree) NewNode(name, parentNs string, nodeType int, property ...string) (string, error) {
	var nodeId, matchReg string
	var newNode n.Node
	if nodeType == n.Root {
		nodeId, name, nodeType, matchReg = rootID, rootNode, n.NonLeaf, NoMachineMatch
		parentNs = "-"
	} else {
		if len(property) > 0 && property[0] != "" {
			matchReg = property[0]
		} else {
			matchReg = NoMachineMatch
		}
		nodeId = common.GenUUID()
	}
	newNode = n.Node{n.NodeProperty{ID: nodeId, Name: name, Type: nodeType, MachineReg: matchReg}, []*n.Node{}}

	var nodes, parent *n.Node
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
		parent, err = nodes.GetByNS(parentNs)
		if err != nil {
			t.logger.Errorf("get parent id ns: %s, error: %v", parentNs, err)
			return "", common.ErrGetParent
		}

		// If the newnode alread exist, return err.
		if exist := nodes.Exist(newNode.Name + nodeDeli + parentNs); exist {
			return "", common.ErrNodeAlreadyExist
		}
		if parent.IsLeaf() {
			t.logger.Error("cannot create node under leaf, leaf node:", parentNs)
			return "", common.ErrCreateNodeUnderLeaf
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
		// Set the template of parent node to this new node.
		templateRes, err := t.templateOfNode(parent.ID)
		if err != nil {
			return "", err
		}
		for k, resStore := range templateRes {
			if nodeType == n.Leaf {
				k = k[len(template):]
			}
			if len(resStore) == 0 {
				continue
			}

			// generate alarm resource new Ns.
			// NOTE: not rollback if make alarm resouce error
			if k == model.Alarm {
				rl := new(model.ResourceList)
				err = rl.Unmarshal([]byte(resStore))
				if err != nil && err != ErrEmptyRes {
					t.logger.Errorf("unmarshal alarm resource fail, parent ns: %s, error: %s, data: %s:",
						parentNs, err, string(resStore))
					return "", err
				}
				for index := range *rl {
					if (*rl)[index], err = GenAlarmFromTemplate(newNode.Name+nodeDeli+parentNs, (*rl)[index], ""); err != nil {
						t.logger.Errorf("make alarm template fail, parent ns: %s, error: %s",
							parentNs, err.Error())
						return "", err
					}
				}
				resStore, err = rl.Marshal()
				if err != nil {
					t.logger.Errorf("marshal alarm template fail, error: %s", err.Error())
					return "", err
				}
			}
			if err = t.setByteToStore(nodeId, k, resStore); err != nil {
				t.logger.Errorf("SetResourceByNs fail when newnode %s, error: %s", nodeId, err.Error())
				return "", err
			}
		}
	}
	return nodeId, nil
}

func (t *Tree) UpdateNode(ns, name, machineReg string) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	allNodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}

	if oldNsSplit := strings.Split(ns, nodeDeli); name != oldNsSplit[0] {
		newNs := strings.Join(append([]string{name}, oldNsSplit[1:]...), nodeDeli)
		if exist := allNodes.Exist(newNs); exist {
			return common.ErrNodeAlreadyExist
		}
	}

	node, err := allNodes.GetByNS(ns)
	if err != nil {
		t.logger.Errorf("GetByNs %s fail, error: %s", ns, err.Error())
		return err
	}
	node.Update(name, machineReg)

	t.Nodes = allNodes
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return err
	}
	return nil
}

// DelNode delete node from tree, remove bucket.
func (t *Tree) DelNode(ns string) error {
	nsSplit := strings.Split(ns, nodeDeli)
	if len(nsSplit) < 2 {
		return common.ErrInvalidParam
	}
	parentNs := strings.Join(nsSplit[1:], nodeDeli)
	delID, err := t.getNodeIDByNS(ns)
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}

	// Not allow delete node which still has child node or machine.
	rl, err := t.GetResourceList(ns, "machine")
	if (err != nil && err != common.ErrNoLeafChild) ||
		(rl != nil && len(*rl) != 0) {
		t.logger.Errorf("not allow delete ns %s, error: %v", ns, err)
		return common.ErrNotAllowDel
	}

	t.Mu.Lock()
	defer t.Mu.Unlock()
	allNodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}
	parentNode, err := allNodes.GetByNS(parentNs)
	if err != nil {
		t.logger.Errorf("GetByNs fail, error: %s", err.Error())
		return err
	}

	if err := parentNode.DelChild(delID); err != nil {
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
	return nil
}
