package resource

import (
	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree/cluster"
	"github.com/lodastack/registry/tree/node"
)

type ResourceInf interface {
	GetResourceList(ns string, resourceType string) (*model.ResourceList, error)
	GetResource(ns, resType string, resID ...string) ([]model.Resource, error)
	SetResource(ns, resType string, rl model.ResourceList) error
	RemoveResource(ns, resType string, resID ...string) error
	CopyResource(fromNs, toNs, resType string, resourceIDs ...string) error

	UpdateResource(ns, resType, resID string, updateMap map[string]string) error
	AppendResource(ns, resType string, appendRes ...model.Resource) error
	MoveResource(oldNs, newNs, resType string, resourceIDs ...string) error
	SearchResource(ns, resType string, search model.ResourceSearch) (map[string]*model.ResourceList, error)
}

type ResourceMethod struct {
	c      cluster.ClusterInf
	n      node.NodeInf
	logger *log.Logger
}

func NewResourceMethod(c cluster.ClusterInf, n node.NodeInf, logger *log.Logger) ResourceInf {
	return &ResourceMethod{c: c, n: n, logger: logger}
}
