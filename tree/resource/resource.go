package resource

import (
	"errors"
	"strings"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/limit"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/cluster"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrEmtpyResource      = errors.New("empty resource")
	defaultResourceWorker = 100
)

// return resource list by nodeId/resource type.
func (r *resourceMethod) getResourceList(nodeID, resourceType string) (*model.ResourceList, error) {
	resByte, err := r.c.View([]byte(nodeID), []byte(resourceType))
	if err != nil {
		return nil, err
	}
	rl := new(model.ResourceList)
	err = rl.Unmarshal(resByte)
	if err != nil && err != common.ErrEmptyResource {
		r.logger.Errorf("unmarshal resource fail, error: %s, data: %s:", err, string(resByte))
		return nil, err
	}
	return rl, nil
}

// return the resource list at form of []byte by nodeId/resource type.
// NOTE: return error ErrEmtpyResource if the resource list in store is emtpy.
func (r *resourceMethod) getResourceListByte(ns, resourceType string) (nodeID string, resByte []byte, err error) {
	nodeID, err = r.n.GetNodeIDByNS(ns)
	if err != nil {
		r.logger.Errorf("getNodeIDByNS fail: %s", err.Error())
		return "", nil, err
	}
	resOldByte, err := r.c.View([]byte(nodeID), []byte(resourceType))
	if err != nil {
		r.logger.Errorf("getByteFromStore fail or get none, nodeid: %s, ns : %s, error: %s", nodeID, resourceType, err.Error())
		return "", nil, errors.New("get resource fail")
	} else if len(resOldByte) == 0 {
		return nodeID, nil, ErrEmtpyResource
	}
	resByte = make([]byte, len(resOldByte))
	copy(resByte, resOldByte)
	return nodeID, resByte, nil
}

// GetResource return the Resource list by ns/resourceType.
// If the node is nonleaf node, return the resource list of all its leaf child node.
func (r *resourceMethod) GetResourceList(ns string, resourceType string) (*model.ResourceList, error) {
	node, err := r.n.GetNodeByNS(ns)
	if err != nil {
		return nil, err
	}

	if node.AllowResource(resourceType) {
		return r.getResourceList(node.ID, resourceType)
	}

	// Return all resource of the node's child leaf if get resource from NonLeaf.
	allResourceList := model.ResourceList{}

	leafIDs, err := node.LeafChildIDs()
	if err != nil {
		if err == common.ErrNoLeafChild {
			return nil, nil
		}
		return nil, err
	}
	for _, leafID := range leafIDs {
		oneNodeResourceList, err := r.getResourceList(leafID, resourceType)
		if err != nil {
			return nil, err
		}
		allResourceList.AppendResources(*oneNodeResourceList)
	}
	return &allResourceList, nil
}

// Get Resource by ns/resource type/resource ID.
func (r *resourceMethod) GetResource(ns, resType string, resID ...string) ([]model.Resource, error) {
	l, err := r.GetResourceList(ns, resType)
	if err != nil {
		r.logger.Errorf("GetResourceList fail, result: %v, error: %v", *l, err)
		return nil, err
	}
	if len(*l) == 0 {
		return nil, nil
	}
	return l.Get(model.IdKey, resID...)
}

// Set ResourceList to ns.
func (r *resourceMethod) SetResource(ns, resType string, rl model.ResourceList) error {
	node, err := r.n.GetNodeByNS(ns)
	if err != nil || node.ID == "" {
		r.logger.Errorf("Get node by ns(%s) fail", ns)
		return common.ErrGetNode
	}
	if !node.AllowResource(resType) {
		return common.ErrSetResourceToLeaf
	}

	var resStore []byte
	resStore, err = rl.Marshal()
	if err != nil {
		r.logger.Errorf("set resource to node fail, marshal resource to byte fail: %s\n", err)
		return err
	}

	return cluster.SetByte(r.c, node.ID, resType, resStore)
}

// UpdateResource One Resource by ns/resource type/resource ID/update map.
// NOTE: read and append at level of []byte, do not unmarshal.
func (r *resourceMethod) UpdateResource(ns, resType, resID string, updateMap map[string]string) error {
	nodeID, resOldByte, err := r.getResourceListByte(ns, resType)
	if err != nil {
		return err
	}

	resNewByte, err := model.UpdateResByID(resOldByte, resID, updateMap)
	if err != nil {
		r.logger.Errorf("UpdateResource fail becource update error: %s", err.Error())
		return err
	}
	return cluster.SetByte(r.c, nodeID, resType, resNewByte)
}

// AppendResource one resource to ns.
func (r *resourceMethod) AppendResource(ns, resType string, appendRes ...model.Resource) error {
	nodeID, resOldByte, err := r.getResourceListByte(ns, resType)
	if err != nil && err != ErrEmtpyResource {
		return err
	}

	resByte, err := model.AppendResources(resOldByte, appendRes...)
	if err != nil {
		r.logger.Errorf("AppendResources error, length of resOld: %d, appendRes: %+v, error: %s", len(resOldByte), appendRes, err.Error())
		return err
	}
	err = cluster.SetByte(r.c, nodeID, resType, resByte)
	return err
}

// DeleteResource remove a resource by ns/resTYpe/resID.
func (r *resourceMethod) RemoveResource(ns, resType string, resID ...string) error {
	nodeID, err := r.n.GetNodeIDByNS(ns)
	if err != nil {
		r.logger.Errorf("getIDByNs fail: %s", err.Error())
		return err
	}
	resOldByte, err := cluster.GetByte(r.c, nodeID, resType)
	if err != nil || len(resOldByte) == 0 {
		r.logger.Errorf("getByteFromStore fail or get none, nodeid: %s, ns : %s, error: %v", nodeID, resType, err)
		return errors.New("get resource fail")
	}

	resNewByte, err := model.DeleteResource(resOldByte, resID...)
	if err != nil {
		return err
	}
	return cluster.SetByte(r.c, nodeID, resType, resNewByte)
}

func (r *resourceMethod) CopyResource(fromNs, toNs, resType string, resourceIDs ...string) error {
	rs, err := r.GetResource(fromNs, resType, resourceIDs...)
	if err != nil || rs == nil {
		r.logger.Errorf("GetResource fail, ns: %s, error: %s", toNs, err)
		return err
	}

	// Check pk value of resource.
	pkValueList := []string{}
	for _, r := range rs {
		pkValue, _ := r.ReadProperty(model.PkProperty[resType])
		if pkValue == "" {
			return errors.New("resource has invalid pk property")
		}
		pkValueList = append(pkValueList, pkValue)
		continue
	}
	// Check pk in new ns.
	searchPk, err := model.NewSearch(false, model.PkProperty[resType], pkValueList...)
	if err != nil {
		r.logger.Errorf("search resource in new ns before move to ns %s fail: %s", toNs, err.Error())
		return err
	}
	searchInNewNs, err := r.SearchResource(toNs, resType, searchPk)
	if err != nil {
		r.logger.Errorf("check the addend resource fail: %s", err.Error())
		return err
	}
	if l, ok := searchInNewNs[toNs]; ok {
		alreadyExist := []string{}
		for _, r := range *l {
			pkV, _ := r.ReadProperty(model.PkProperty[resType])
			alreadyExist = append(alreadyExist, pkV)
		}
		r.logger.Errorf("resource pk %v already exist in the ns, data: %+v", alreadyExist, l)
		return errors.New("resource pk " + strings.Join(alreadyExist, ",") + " already in new ns")
	}

	// set new resID to rs.
	for i := range rs {
		rs[i].NewID()
	}

	if err := r.AppendResource(toNs, resType, rs...); err != nil {
		r.logger.Errorf("AppendResource resource fail, ns %s, resource type: %s, resourceID: %v, error: %s",
			toNs, resType, rs[0], err.Error())
		return err
	}
	return nil
}

// MoveResource move the resource to a new ns.
func (r *resourceMethod) MoveResource(oldNs, newNs, resType string, resourceIDs ...string) error {
	if err := r.CopyResource(oldNs, newNs, resType, resourceIDs...); err != nil {
		return err
	}
	if err := r.RemoveResource(oldNs, resType, resourceIDs...); err != nil {
		r.logger.Errorf("DeleteResource resource fail, ns %s, resource type: %s, resourceID: %v, error: %s",
			newNs, resType, resourceIDs, err.Error())
		return err
	}
	return nil
}

// SearchResource search the resource.
func (r *resourceMethod) SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error) {
	result := map[string]*model.ResourceList{}
	leafIDs, err := r.n.LeafChildIDs(ns)
	if err != nil && len(leafIDs) == 0 {
		r.logger.Errorf("node has none leaf, ns: %s, error: %v", ns, err)
		return nil, common.ErrNilChildNode
	}

	var fail bool
	limit := limit.NewLimit(defaultResourceWorker)
	resultChan := make(chan map[string]*model.ResourceList, defaultResourceWorker/2)
	defer close(resultChan)
	// Collect process result.
	go func() {
		for {
			select {
			case nsResult, live := <-resultChan:
				if !live {
					limit.Close()
					return
				}
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

	if err := search.Init(); err != nil {
		return nil, err
	}
	// Search in ns and report the result.
	for _, leafID := range leafIDs {
		limit.Take()
		go func(leafID string, search model.ResourceSearch) {
			nsResult := map[string]*model.ResourceList{}
			resByte, err := cluster.GetByte(r.c, leafID, resType)
			// report error when getByteFromStore fail.
			if len(resByte) == 0 {
				limit.Release()
				return
			}
			if err != nil {
				r.logger.Errorf("getByteFromStore fail or none input, id: %s, type: %s, input length:%d, error: %s",
					leafID, resType, len(resByte), err.Error())
				limit.Error(err)
				return
			}

			resOfOneNs, err := search.Process(resByte)
			// report error when search fail.
			if err != nil {
				r.logger.Errorf("Search fail, getNsByID error: %s", err.Error())
				limit.Error(err)
				return
			}
			if len(resOfOneNs) != 0 {
				ns, err := r.n.GetNodeNSByID(leafID)
				if err != nil {
					r.logger.Errorf("getNsByID favil, getNsByID error: %s", err.Error())
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

	// Wait the process is complete done.
	limit.Wait()
	if fail {
		return nil, errors.New("SearchResourceByNs fail")
	}

	return result, nil
}
