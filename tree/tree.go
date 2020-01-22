package tree

import (
	"strings"
	"sync"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/cluster"
	"github.com/lodastack/registry/tree/machine"
	"github.com/lodastack/registry/tree/node"
	"github.com/lodastack/registry/tree/resource"
)

var (
	template = model.TemplatePrefix
)

const (
	nodeBucket   = "loda"
	reportBucket = "report"

	rootNodeID = "0"
)

// Tree manage the node/resource/machine.
type Tree struct {
	Nodes    *node.Node
	cluster  cluster.Inf
	node     node.Inf
	resource resource.Inf
	machine  machine.Inf
	Mu       sync.RWMutex

	reports ReportInfo
	logger  *log.Logger
}

// NewTree return Tree obj.
func NewTree(cluster cluster.Inf) (*Tree, error) {
	nodeInf := node.NewNode(cluster)
	logger := log.New(config.C.LogConf.Level, "tree", model.LogBackend)
	r := resource.NewResource(cluster, nodeInf, logger)
	t := Tree{
		Nodes: &node.Node{
			node.NodeProperty{ID: rootNodeID, Name: node.RootNode, Type: node.NonLeaf, MachineReg: node.NotMatchMachine},
			[]*node.Node{}},
		cluster:  cluster,
		node:     nodeInf,
		resource: r,
		machine:  machine.NewMachine(nodeInf, r, logger),
		Mu:       sync.RWMutex{},
		logger:   logger,
		reports:  ReportInfo{sync.RWMutex{}, make(map[string]model.Report)},
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
	err := t.cluster.CreateBucketIfNotExist([]byte(nodeBucket))
	if err != nil {
		t.logger.Errorf("tree %s CreateBucketIfNotExist fail: %s", nodeBucket, err.Error())
		return err
	}
	if err := t.initNodeData(node.NodeDataKey); err != nil {
		t.logger.Error("init nodeDataKey fail:", err.Error())
		return err
	}

	return nil
}

// initialization tree node data and if empty.
func (t *Tree) initNodeData(key string) error {
	v, err := t.cluster.View([]byte(node.NodeDataBucketID), []byte(key))
	if err != nil {
		return err
	}
	if len(v) != 0 {
		return nil
	}

	t.logger.Info(key, "is not inited, begin to init")

	// Create rootNode map/bucket and init template.
	if _, err := t.NewNode("", "", "", node.Root); err != nil {
		panic("create root node fail: " + err.Error())
	}
	// Create root pool node.
	if _, err := t.NewNode(node.PoolNode, "pool", node.RootNode, node.Leaf, node.NotMatchMachine); err != nil {
		panic("create root pool node fail: " + err.Error())
	}

	t.NewNode("monitor", "monitor system", node.RootNode, node.NonLeaf, node.NotMatchMachine)
	t.NewNode("db", "monitor system", "monitor.loda", node.NonLeaf, node.NotMatchMachine)
	t.NewNode("common", "monitor system", "db.monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("alarm", "monitor system", "monitor.loda", node.NonLeaf, node.NotMatchMachine)
	t.NewNode("kapacitor", "monitor system", "alarm.monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("adapter", "monitor system", "alarm.monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("nodata", "monitor system", "alarm.monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("event", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("etcd", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("registry", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("mq", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("router", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)
	t.NewNode("ui", "monitor system", "monitor.loda", node.Leaf, node.NotMatchMachine)

	return nil
}

func (t *Tree) initReportBucket() error {
	err := t.cluster.CreateBucketIfNotExist([]byte(reportBucket))
	if err != nil {
		t.logger.Errorf("tree init %s CreateBucketIfNotExist fail: %s", reportBucket, err.Error())
		return err
	}
	t.reports.ReportInfo, err = t.readReport()
	if err != nil {
		t.logger.Error("tree init report fail, set empty")
		t.reports.ReportInfo = make(map[string]model.Report)
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
		c := time.Tick(time.Minute * 10)
		for {
			select {
			case <-c:
				reports := t.GetReportInfo()
				if err := t.CheckMachineStatusByReport(reports); err != nil {
					t.logger.Error("UpdateMachineStatusByReport fail:", err.Error())
				}
			}
		}
	}()
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
	return t.node.Save(treeByte)
}

// Create bucket for node.
func (t *Tree) createBucketForNode(nodeID string) error {
	return t.cluster.CreateBucket([]byte(nodeID))
}

func (t *Tree) removeNodeResourceFromStore(nodeID string) error {
	return t.cluster.RemoveBucket([]byte(nodeID))
}

// Get type resType resource of node with ID bucketId.
func (t *Tree) getByteFromStore(bucket, resType string) ([]byte, error) {
	return t.cluster.View([]byte(bucket), []byte(resType))
}

// Set resource to node bucket.
func (t *Tree) setByteToStore(bucket, resType string, resByte []byte) error {
	return t.cluster.Update([]byte(bucket), []byte(resType), resByte)
}

func (t *Tree) templateOfNode(nodeID string) (map[string][]byte, error) {
	return t.cluster.ViewPrefix([]byte(nodeID), []byte(template))
}

// UpdateNode update the node name or machineMatchStrategy.
func (t *Tree) UpdateNode(ns, name, comment, machineMatchStrategy string) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	allNodes, err := t.AllNodes()
	if err != nil {
		t.logger.Error("get all nodes error when GetNodesById")
		return err
	}

	// check the new ns exist or not if update the node name.
	if oldNsSplit := strings.Split(ns, node.NodeDeli); name != oldNsSplit[0] {
		newNs := strings.Join(append([]string{name}, oldNsSplit[1:]...), node.NodeDeli)
		if exist := allNodes.Exist(newNs); exist {
			return common.ErrNodeAlreadyExist
		}
	}

	node, err := allNodes.GetByNS(ns)
	if err != nil {
		t.logger.Errorf("GetByNs %s fail, error: %s", ns, err.Error())
		return err
	}
	node.Update(name, comment, machineMatchStrategy)

	t.Nodes = allNodes
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return err
	}
	return nil
}

func getParentNS(ns string) (string, error) {
	nsSplit := strings.Split(ns, node.NodeDeli)
	if len(nsSplit) < 2 {
		return "", common.ErrInvalidParam
	}
	return strings.Join(nsSplit[1:], node.NodeDeli), nil
}

func (t *Tree) allowRemoveNS(ns string) error {
	rl, err := t.GetResourceList(ns, "machine")
	// not allow remove the nonleaf node still has child node.
	if err != nil && err != common.ErrNoLeafChild {

		return err
	}
	// not allow remove the leaf node still has machine resource.
	if rl != nil && len(*rl) != 0 {
		t.logger.Errorf("not allow delete ns %s, error: %v", ns, err)
		return common.ErrNotAllowDel
	}
	return nil
}

// RemoveNode remove node from tree, remove bucket which save the resource.
func (t *Tree) RemoveNode(ns string) error {
	parentNs, err := getParentNS(ns)
	if err != nil {
		t.logger.Errorf("remove ns fail because the ns is root node or invalid, ns: %s", ns)
		return err
	}

	removeNodeID, err := t.getNodeIDByNS(ns)
	if err != nil {
		t.logger.Errorf("getNodeIDByNS error: %s", err.Error())
		return err
	}

	if err := t.allowRemoveNS(ns); err != nil {
		return err
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
	if err := parentNode.RemoveChildNode(removeNodeID); err != nil {
		t.logger.Errorf("delete node fail, parent ns: %s, delete ID: %s, error: %s", parentNs, removeNodeID, err.Error())
		return err
	}
	t.logger.Infof("remove node (ID: %s) behind ns %s from store success", removeNodeID, parentNs)
	t.Nodes = allNodes
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return err
	}

	if err := t.removeNodeResourceFromStore(removeNodeID); err != nil {
		t.logger.Errorf("remove node from store fail, parent ns: %s, delete ID: %s, error: %s", parentNs, removeNodeID, err.Error())
		return err
	}
	return nil
}

// NewNode create a node, return a pointer which point to node, and it bucketId. Property is preserved.
// First property argument is used as the machineReg.
func (t *Tree) NewNode(name, comment, parentNs string, nodeType int, machineRegistRule ...string) (string, error) {
	newNode := node.Node{
		node.NodeProperty{},
		[]*node.Node{},
	}
	newNode.Comment = comment
	if nodeType == node.Root {
		newNode.ID = rootNodeID
		newNode.Name, newNode.Type = node.RootNode, node.NonLeaf
		newNode.MachineReg = node.NotMatchMachine
		parentNs = "-"
	} else {
		newNode.ID = common.GenUUID()
		newNode.Name, newNode.Type = name, nodeType
		if len(machineRegistRule) > 0 && machineRegistRule[0] != "" {
			newNode.MachineReg = machineRegistRule[0]
		} else {
			newNode.MachineReg = node.NotMatchMachine
		}
	}

	// Create Pool node not lock the tree, because create leaf will lock the tree.
	parentNodeID, err := t.addNewNodeToTree(newNode, parentNs, nodeType)
	if err != nil {
		return "", err
	}
	if err := t.saveTree(); err != nil {
		t.logger.Error("NewNode save tree node fail,", err.Error())
		return "", err
	}

	// Create a new bucket fot this node resource.
	if err := t.createBucketForNode(newNode.ID); err != nil {
		t.logger.Errorf("NewNode createNodeBucket fail, nodeid:%s, error: %s\n", newNode.ID, err.Error())
		// rollback by remove the new node.
		if nodeType != node.Root {
			parent, err := t.Nodes.GetByNS(parentNs)
			if err != nil {
				t.logger.Errorf("get parent id ns: %s, error: %v", parentNs, err)
				return "", common.ErrGetParent
			}
			parent.Children = parent.Children[:len(parent.Children)-1]
		}
		if err := t.saveTree(); err != nil {
			t.logger.Errorf("Rollback tree node fail: %s", err.Error())
		}
		return "", err
	}

	return newNode.ID, t.initResourceOrTemplate(newNode, nodeType, parentNs, parentNodeID)
}

func (t *Tree) addNewNodeToTree(newNode node.Node, parentNs string, nodeType int) (string, error) {
	var nodes, parent *node.Node
	var err error
	t.Mu.Lock()
	defer t.Mu.Unlock()

	// use the node as root node.
	if nodeType == node.Root {
		nodes = &newNode
		t.Nodes = nodes
		return "", nil
	}

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

	newNS := newNode.Name + node.NodeDeli + parentNs
	if exist := nodes.Exist(newNS); exist {
		return "", common.ErrNodeAlreadyExist
	}
	if parent.IsLeaf() {
		t.logger.Error("cannot create node under leaf, leaf node:", parentNs)
		return "", common.ErrCreateNodeUnderLeaf
	}
	parent.Children = append(parent.Children, &newNode)
	t.Nodes = nodes
	return parent.ID, nil
}

func (t *Tree) initResourceOrTemplate(newNode node.Node, nodeType int, parentNs, parentNodeID string) error {
	if nodeType == node.Root {
		for k, res := range model.RootTemplate {
			if err := t.SetResource(node.RootNode, k, res); err != nil {
				t.logger.Errorf("SetResourceByNs fail when create rootNode, error: %s", err.Error())
			}
		}
		return nil
	}

	// Set the template of parent node to this new node.
	templateRes, err := t.templateOfNode(parentNodeID)
	if err != nil {
		return err
	}
	for templateName, templateValue := range templateRes {
		var resourceName string
		if nodeType == node.Leaf {
			resourceName = templateName[len(template):]
		} else {
			resourceName = templateName
		}
		if len(templateValue) == 0 {
			continue
		}

		// generate alarm resource new Ns.
		// NOTE: no rollback if make alarm resouce error.
		if resourceName == model.Alarm {
			rl := new(model.ResourceList)
			err = rl.Unmarshal([]byte(templateValue))
			if err != nil && err != common.ErrEmptyResource {
				t.logger.Errorf("unmarshal alarm resource fail, parent ns: %s, error: %s, data: %s:",
					parentNs, err, string(templateValue))
				return err
			}
			for index := range *rl {
				if (*rl)[index], err = GenAlarmFromTemplate(newNode.Name+node.NodeDeli+parentNs, (*rl)[index], ""); err != nil {
					t.logger.Errorf("make alarm template fail, parent ns: %s, error: %s",
						parentNs, err.Error())
					return err
				}
			}
			templateValue, err = rl.Marshal()
			if err != nil {
				t.logger.Errorf("marshal alarm template fail, error: %s", err.Error())
				return err
			}
		}
		if err = t.setByteToStore(newNode.ID, resourceName, templateValue); err != nil {
			t.logger.Errorf("SetResourceByNs fail when newnode %s, error: %s", newNode.ID, err.Error())
			return err
		}
	}
	return err
}
