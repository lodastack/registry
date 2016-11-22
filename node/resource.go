package node

import (
	"github.com/lodastack/registry/model"
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

func (t *Tree) SearchResourceByNs(ns, resType string, search model.ResourceSearch) (map[string]*model.Resources, error) {
	leafIDs, err := t.Leaf(ns, IDFormat)
	if err != nil {
		return nil, err
	} else if len(leafIDs) == 0 {
		t.logger.Errorf("SearchResourceByNs fail: node has none leaf")
		return nil, ErrNilChildNode
	}

	result := map[string]*model.Resources{}

	for _, leafID := range leafIDs {
		resByte, err := t.getByteFromStore(leafID, resType)
		if err != nil {
			t.logger.Errorf("SearchResourceByNs fail,id: %s, type: %s ,getByteFromStore error: %s", leafID, resType, err.Error())
			return nil, err
		}

		search.Init()
		if resOfOneNs, err := search.Process(resByte); err != nil {
			return nil, err
		} else if len(resOfOneNs) != 0 {
			ns, err := t.getNsByID(leafID)
			if err != nil {
				t.logger.Errorf("SearchResourceByNs fail, getNsByID error: %s", err.Error())
				return nil, err
			}
			result[ns] = &model.Resources{}
			result[ns].AppendResources(resOfOneNs)
		}
	}

	return result, nil
}
