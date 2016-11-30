package node

import (
	"errors"

	"github.com/lodastack/registry/model"
)

var (
	defaultResourceWorker = 100
)

// GetResource return the ResourceType resource belong to the node with NodeId.
// TODO: Permission Check
func (t *Tree) GetResourceByNodeID(NodeId string, ResourceType string) (*model.Resources, error) {
	ns, err := t.getNsByID(NodeId)
	if err != nil {
		t.logger.Errorf("get ns error when get resource: %s", err.Error())
		return nil, err
	}
	return t.GetResourceByNs(ns, ResourceType)
}

// GetResource return the ResourceType resource belong to the node with NodeName.
// TODO: Permission Check
func (t *Tree) GetResourceByNs(ns string, resourceType string) (*model.Resources, error) {
	node, err := t.GetNodeByNs(ns)
	if err != nil {
		t.logger.Errorf("get resource fail because get node by ns fail, ns: %s, resource: %s", ns, resourceType)
		return nil, err
	}

	// If get resource of NonLeaf, get resource at its leaf child node.
	if !node.AllowResource(resourceType) {
		return t.GetResourceFromNonLeaf(node, resourceType)
	}
	return t.getResFromStore(node.ID, resourceType)
}

func (t *Tree) getResFromStore(nodeId, resourceType string) (*model.Resources, error) {
	resByte, err := t.getByteFromStore(nodeId, resourceType)
	if err != nil {
		return nil, err
	}
	resources := new(model.Resources)
	err = resources.Unmarshal(resByte)
	if err != nil && err != ErrEmptyRes {
		t.logger.Error("GetResourceByNodeID fail:", err, string(resByte))
		return nil, err
	}
	return resources, nil
}

func (t *Tree) UpdateResourceByNs(ns, resType, resID string, updateMap map[string]string) error {
	nodeId, err := t.getIDByNs(ns)
	if err != nil {
		t.logger.Errorf("getIDByNs fail: %s", err.Error())
		return err
	}
	resOldByte, err := t.getByteFromStore(nodeId, resType)
	if err != nil || len(resOldByte) == 0 {
		t.logger.Errorf("getByteFromStore fail or get none, nodeid: %s, ns : %s, error: %v", nodeId, resType, err)
		return errors.New("get resource fail")
	}
	resNewByte, err := model.UpdateResByID(resOldByte, resID, updateMap)
	if err != nil {
		t.logger.Errorf("UpdateResourceByNs fail becource update error: %s", err.Error())
		return err
	}
	return t.setResourceByNodeID(nodeId, resType, resNewByte)
}

func (t *Tree) GetResourceFromNonLeaf(nonLeaf *Node, resourceType string) (*model.Resources, error) {
	allRes := model.Resources{}
	leafIDs, err := nonLeaf.leafID()
	if err != nil {
		return nil, err
	}

	for _, leafID := range leafIDs {
		if res, err := t.getResFromStore(leafID, resourceType); err != nil {
			return nil, err
		} else {
			allRes.AppendResources(*res)
		}
	}
	return &allRes, nil
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
	if !node.AllowResource(resType) {
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

// TODO: remove time and debug log
func (t *Tree) SearchResourceByNs(ns, resType string, search model.ResourceSearch) (map[string]*model.Resources, error) {
	result := map[string]*model.Resources{}
	leafIDs, err := t.LeafIDs(ns)
	if err != nil && len(leafIDs) == 0 {
		t.logger.Errorf("node has none leaf, ns: %s, error: %v", ns, err)
		return nil, ErrNilChildNode
	}

	var fail bool
	limit := NewFixed(defaultResourceWorker)
	resultChan := make(chan map[string]*model.Resources, defaultResourceWorker/2)
	// collect process result
	go func() {
		for {
			select {
			case nsResult := <-resultChan:
				for k, v := range nsResult {
					result[k] = v
				}
				limit.Release()
			case <-limit.Err:
				fail = true
				limit.Release()
			}
		}
	}()

	// search ns and report the result.
	if err := search.Init(); err != nil {
		return nil, err
	}
	for _, leafID := range leafIDs {
		limit.Take()
		go func(leafID string, search model.ResourceSearch) {
			nsResult := map[string]*model.Resources{}
			resByte, err := t.getByteFromStore(leafID, resType)
			// report error when getByteFromStore fail.
			if err != nil || len(resByte) == 0 {
				t.logger.Errorf("getByteFromStore fail or none input, id: %s, type: %s, input length:%d, error: %v",
					leafID, resType, len(resByte), err.Error())
				limit.Error(err)
				return
			}

			resOfOneNs, err := search.Process(resByte)
			// report error when search fail.
			if err != nil {
				t.logger.Errorf("Search fail, getNsByID error: %s", err.Error())
				limit.Error(err)
				return
			}
			if len(resOfOneNs) != 0 {
				ns, err := t.getNsByID(leafID)
				if err != nil {
					t.logger.Errorf("getNsByID favil, getNsByID error: %s", err.Error())
					limit.Error(err)
				} else {
					nsResult[ns] = &resOfOneNs
					resultChan <- nsResult
				}
			} else {
				// release the limit when search nothing.
				limit.Release()
			}

		}(leafID, search)
	}
	limit.Wait()
	if fail {
		return nil, errors.New("SearchResourceByNs fail")
	}

	return result, nil
}
