package node

import (
	"errors"

	"github.com/lodastack/registry/model"
)

var (
	ErrNotFound           = errors.New("not found")
	defaultResourceWorker = 100
)

func (t *Tree) getResFromStore(nodeId, resourceType string) (*model.ResourceList, error) {
	resByte, err := t.getByteFromStore(nodeId, resourceType)
	if err != nil {
		return nil, err
	}
	rl := new(model.ResourceList)
	err = rl.Unmarshal(resByte)
	if err != nil && err != ErrEmptyRes {
		t.logger.Errorf("unmarshal resource fail, error: %s, data: %s:", err, string(resByte))
		return nil, err
	}
	return rl, nil
}

// GetResource return the ResourceType resource belong to the node with NodeName.
// TODO: Permission Check
func (t *Tree) GetResourceList(ns string, resourceType string) (*model.ResourceList, error) {
	node, err := t.GetNode(ns)
	if err != nil {
		t.logger.Errorf("get resource fail because get node by ns fail, ns: %s, resource: %s", ns, resourceType)
		return nil, err
	}

	if node.AllowResource(resourceType) {
		return t.getResFromStore(node.ID, resourceType)
	}

	// Return all resource of the node's child leaf if get resource from NonLeaf.
	allRes := model.ResourceList{}
	leafIDs, err := node.leafID()
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

func (t *Tree) GetOneResource(ns, resType, resID string) (model.Resource, error) {
	rl, err := t.GetResourceList(ns, resType)
	if err != nil || len(*rl) == 0 {
		t.logger.Errorf("GetResourceList fail, result: %v, error: %v", *rl, err.Error())
		return nil, err
	}
	return rl.GetOneResource(resID)
}

func (t *Tree) UpdateResource(ns, resType, resID string, updateMap map[string]string) error {
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
		t.logger.Errorf("UpdateResource fail becource update error: %s", err.Error())
		return err
	}
	return t.setByteToStore(nodeId, resType, resNewByte)
}

// Append one resource to ns.
func (t *Tree) AppendResource(ns, resType string, appendRes model.Resource) (string, error) {
	nodeID, err := t.getIDByNs(ns)
	if err != nil {
		t.logger.Errorf("getID of ns %s fail when appendResource, error: %+v", ns, err)
		return "", err
	}
	resOldByte, err := t.getByteFromStore(nodeID, resType)
	if err != nil {
		t.logger.Errorf("resByteOfNode error, length of resOldByte: %d, error: %s", len(resOldByte), err.Error())
		return "", err
	}
	resByte, UUID, err := model.AppendResources(resOldByte, appendRes)
	if err != nil {
		t.logger.Errorf("AppendResources error, length of resOld: %d, appendRes: %+v, error: %s", len(resOldByte), appendRes, err.Error())
		return "", err
	}
	err = t.setByteToStore(nodeID, resType, resByte)
	return UUID, err
}

func (t *Tree) SetResource(ns, resType string, ResByte []byte) error {
	node, err := t.GetNode(ns)
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

	return t.setByteToStore(node.ID, resType, resStore)
}

func (t *Tree) DeleteResource(ns, resType, resId string) error {
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

	resNewByte, err := model.DeleteResource(resOldByte, resId)
	if err != nil {
		return err
	}
	return t.setByteToStore(nodeId, resType, resNewByte)
}

// TODO: check resource name
func (t *Tree) MoveResource(oldNs, newNs, resType, resourceID string) error {
	resource, err := t.GetOneResource(oldNs, resType, resourceID)
	if err != nil {
		t.logger.Errorf("GetOneResource fail: %s", newNs, err.Error())
		return err
	}
	_, err = t.GetOneResource(newNs, resType, resourceID)
	if err != ErrNotFound {
		t.logger.Errorf("connot move resource from ns %s to ns: %s, resource type: %s,resourceID: %s",
			oldNs, newNs, resType, resourceID)
		return err
	}
	if _, err := t.AppendResource(newNs, resType, resource); err != nil {
		t.logger.Errorf("AppendResource resource fail, ns %s, resource type: %s, resourceID: %v, error: %s",
			newNs, resType, resource, err.Error())
		return err
	}
	if err := t.DeleteResource(oldNs, resType, resourceID); err != nil {
		t.logger.Errorf("DeleteResource resource fail, ns %s, resource type: %s, resourceID: %s, error: %s",
			newNs, resType, resourceID, err.Error())
		return err
	}
	return nil
}

// TODO: remove time and debug log
func (t *Tree) SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error) {
	result := map[string]*model.ResourceList{}
	leafIDs, err := t.LeafIDs(ns)
	if err != nil && len(leafIDs) == 0 {
		t.logger.Errorf("node has none leaf, ns: %s, error: %v", ns, err)
		return nil, ErrNilChildNode
	}

	var fail bool
	limit := NewFixed(defaultResourceWorker)
	resultChan := make(chan map[string]*model.ResourceList, defaultResourceWorker/2)
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
			nsResult := map[string]*model.ResourceList{}
			resByte, err := t.getByteFromStore(leafID, resType)
			// report error when getByteFromStore fail.
			if len(resByte) == 0 {
				limit.Release()
				return
			}
			if err != nil {
				t.logger.Errorf("getByteFromStore fail or none input, id: %s, type: %s, input length:%d, error: %s",
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
