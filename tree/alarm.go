package tree

import (
	"github.com/lodastack/registry/authorize"
	"github.com/lodastack/registry/model"
)

// GenAlarmFromTemplate set the gourp infomation and return alarm.
func GenAlarmFromTemplate(ns string, data map[string]string, ID string) (model.Resource, error) {
	data["groups"] = authorize.GetNsOpGName(ns)
	return model.NewAlarmResourceByMap(ns, data, ID)
}
