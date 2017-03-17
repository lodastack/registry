package node

import (
	"github.com/lodastack/registry/authorize"
	"github.com/lodastack/registry/model"
)

func GenAlarmFromTemplate(ns string, data map[string]string, ID string) (model.Resource, error) {
	data["groups"] = authorize.GetNsOpGName(ns)
	return model.NewAlarmResourceByMap(ns, data, ID)
}
