package node

import (
	"github.com/lodastack/registry/model"
)

// GetResource return the ResourceType resource belong to the node with NodeId.
// TODO: Permission Check
func (t *Tree) GetResourceByNodeID(NodeId string, ResourceType string) (*model.Resources, error) {
	resByte, err := t.getResByteOfNode(NodeId, ResourceType)
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

// GetResource return the ResourceType resource belong to the node with NodeName.
// TODO: Permission Check
func (t *Tree) GetResourceByNs(ns string, ResourceType string) (*model.Resources, error) {
	nodeId, err := t.getIDByNs(ns)
	if nodeId == "" || err != nil {
		t.logger.Error("GetResourceByNodeName GetNodeByNs fail, ns %s, error:%+v \n", ns, err)
		return nil, ErrGetNodeID
	}
	return t.GetResourceByNodeID(nodeId, ResourceType)
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
	leafIDs, err := t.GetLeaf(ns, IDFormat)
	if err != nil {
		return nil, err
	} else if len(leafIDs) == 0 {
		return nil, ErrNilChildNode
	}

	result := map[string]*model.Resources{}

	for _, leafID := range leafIDs {
		resByte, err := t.getResByteOfNode(leafID, resType)
		if err != nil {
			return nil, err
		}

		search.Init()
		if resOfOneNs, err := search.Process(resByte); err != nil {
			return nil, err
		} else if len(resOfOneNs) != 0 {
			ns, err := t.getNsByID(leafID)
			if err != nil {
				return nil, err
			}
			result[ns] = &model.Resources{}
			result[ns].AppendResources(resOfOneNs)
		}
	}

	return result, nil
}
