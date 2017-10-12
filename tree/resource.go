package tree

import (
	"github.com/lodastack/registry/model"
)

func (t *Tree) SetResource(ns, resType string, l model.ResourceList) error {
	return t.r.SetResource(ns, resType, l)
}

func (t *Tree) GetResource(ns, resourceType string, stringresID ...string) ([]model.Resource, error) {
	return t.r.GetResource(ns, resourceType, stringresID...)
}

func (t *Tree) GetResourceList(ns, resourceType string) (*model.ResourceList, error) {
	return t.r.GetResourceList(ns, resourceType)
}

func (t *Tree) UpdateResource(ns, resType, resID string, updateMap map[string]string) error {
	return t.r.UpdateResource(ns, resType, resID, updateMap)
}

func (t *Tree) AppendResource(ns, resType string, appendRes ...model.Resource) error {
	return t.r.AppendResource(ns, resType, appendRes...)
}

func (t *Tree) MoveResource(oldNs, newNs, resType string, resourceIDs ...string) error {
	return t.r.MoveResource(oldNs, newNs, resType, resourceIDs...)
}

func (t *Tree) SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error) {
	return t.r.SearchResource(ns, resType, search)
}

func (t *Tree) CopyResource(fromNs, toNs, resType string, resourceIDs ...string) error {
	return t.r.CopyResource(fromNs, toNs, resType, resourceIDs...)
}

func (t *Tree) RemoveResource(ns, resourceType string, resId ...string) error {
	return t.r.RemoveResource(ns, resourceType, resId...)
}
