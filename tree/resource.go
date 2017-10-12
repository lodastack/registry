package tree

import (
	"github.com/lodastack/registry/model"
)

// SetResource set the resource list to the ns.
func (t *Tree) SetResource(ns, resType string, l model.ResourceList) error {
	return t.r.SetResource(ns, resType, l)
}

// GetResource return the one resource of the ns.
func (t *Tree) GetResource(ns, resourceType string, stringresID ...string) ([]model.Resource, error) {
	return t.r.GetResource(ns, resourceType, stringresID...)
}

// GetResourceList return a type resource list of a node.
func (t *Tree) GetResourceList(ns, resourceType string) (*model.ResourceList, error) {
	return t.r.GetResourceList(ns, resourceType)
}

// UpdateResource update one resource by updateMap.
func (t *Tree) UpdateResource(ns, resType, resID string, updateMap map[string]string) error {
	return t.r.UpdateResource(ns, resType, resID, updateMap)
}

// AppendResource append resources to a ns.
func (t *Tree) AppendResource(ns, resType string, appendRes ...model.Resource) error {
	return t.r.AppendResource(ns, resType, appendRes...)
}

// MoveResource move one resource fo an other ns, the resouce will be removed from the old ns.
func (t *Tree) MoveResource(oldNs, newNs, resType string, resourceIDs ...string) error {
	return t.r.MoveResource(oldNs, newNs, resType, resourceIDs...)
}

// SearchResource search any preperty resource in the ns and its child ns.
func (t *Tree) SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error) {
	return t.r.SearchResource(ns, resType, search)
}

// CopyResource copy one resource from one ns to the other ns, the resource will still exist in the old ns.
func (t *Tree) CopyResource(fromNs, toNs, resType string, resourceIDs ...string) error {
	return t.r.CopyResource(fromNs, toNs, resType, resourceIDs...)
}

// RemoveResource remove one resource from a node.
func (t *Tree) RemoveResource(ns, resourceType string, resID ...string) error {
	return t.r.RemoveResource(ns, resourceType, resID...)
}
