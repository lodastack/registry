package machine

import (
	"github.com/lodastack/log"
	m "github.com/lodastack/models"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/node/cluster"
	"github.com/lodastack/registry/node/node"
	"github.com/lodastack/registry/node/resource"
)

type MachineInf interface {
	SearchMachine(hostname string) (map[string]string, error)
	MachineUpdate(oldName string, updateMap map[string]string) error
	MatchNs(hostname string) ([]string, error)
	RegisterMachine(newMachine model.Resource) (map[string]string, error)
	CheckMachineStatusByReport(reports map[string]m.Report) error
}

type Machine struct {
	c      cluster.ClusterInf
	n      node.NodeInf
	r      resource.ResourceInf
	logger *log.Logger
}

func NewMachine(c cluster.ClusterInf, n node.NodeInf, r resource.ResourceInf, logger *log.Logger) MachineInf {
	return &Machine{c: c, n: n, r: r, logger: logger}
}
