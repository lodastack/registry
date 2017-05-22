package authorize

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
)

var (
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupAlreadyExist = errors.New("group already exist")

	OP                  = "op"
	DEV                 = "dev"
	nsSep        string = "."
	groupNameSep byte   = '-'
)

type Group struct {
	GName    string   `json:"gname"`
	Managers []string `json:"managers"`
	Members  []string `json:"members"`
	Items    []string `json:"items"`

	cluster Cluster `json:"-"`
}

func getGKey(gName string) []byte { return []byte("g-" + gName) }

// GetGNameByNs reverse the ns and join the name.
// The purpose of reverse the ns is list the group by prefix,
// so that can list the group by ns.
// e.g: server1.product1.loda, op -> loda.product1.server1-op
func GetGNameByNs(ns, name string) string {
	return joinGroupName(reverceNs(ns), name)
}

func reverceNs(ns string) string {
	nsSplit := strings.Split(ns, nsSep)
	return strings.Join(common.Reverse(nsSplit), nsSep)
}

func joinGroupName(ns, name string) string {
	return ns + string(groupNameSep) + name
}

func readGName(gname string) (ns, name string) {
	lastIndex := strings.LastIndexByte(gname, groupNameSep)
	ns, name = gname[:lastIndex], gname[lastIndex+1:]
	return
}

func (g *Group) ReadGName(gname string) (ns, name string) {
	ns, name = readGName(gname)
	ns = reverceNs(ns)
	return
}

func GetNsDevGName(ns string) string {
	return GetGNameByNs(ns, DEV)
}

func GetNsOpGName(ns string) string {
	return GetGNameByNs(ns, OP)
}

func (g *Group) Byte() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Group) GetGroup(gName string) (Group, error) {
	group := Group{}
	if gName == "" {
		return group, ErrInvalidParam
	}
	gByte, err := g.cluster.View([]byte(AuthBuck), getGKey(gName))
	if err != nil {
		return group, err
	}

	if len(gByte) == 0 {
		return group, ErrGroupNotFound
	}
	err = json.Unmarshal(gByte, &group)
	return group, err
}

func (g *Group) ListNsGroup(ns string) ([]Group, error) {
	gNamePrefix := GetGNameByNs(ns, "")
	reverseNs := reverceNs(ns)
	groupMap, err := g.cluster.ViewPrefix([]byte(AuthBuck), getGKey(gNamePrefix))
	if err != nil {
		return nil, err
	}

	GroupList, i := make([]Group, len(groupMap)), 0
	for _, gByte := range groupMap {
		if len(gByte) == 0 {
			continue
		}
		group := Group{}
		err = json.Unmarshal(gByte, &group)
		if err != nil {
			return GroupList, err
		}

		groupNs, _ := readGName(group.GName)
		if reverseNs == groupNs {
			GroupList[i] = group
			i++
		}
	}
	return GroupList[:i], nil
}

func (g *Group) createGroup(gName string, managers, members, items []string) (model.Row, error) {
	updateRow := model.Row{}
	if gName == "" {
		return updateRow, ErrInvalidParam
	}
	_, err := g.GetGroup(gName)
	if err != ErrGroupNotFound {
		if err == nil {
			return updateRow, ErrGroupAlreadyExist
		}
		return updateRow, err
	}

	gByte, err := (&Group{
		GName:    gName,
		Managers: managers,
		Members:  members,
		Items:    items,
	}).Byte()
	if err != nil {
		return updateRow, err
	}

	updateRow = model.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: gByte}
	return updateRow, nil
}

// UpdateGroup update, not update member infomation.
// TODO: manager check
func (g *Group) UpdateItems(gName string, items []string) error {
	group, err := g.GetGroup(gName)
	if err != nil {
		return err
	}

	if len(items) == 0 || items[0] == "" {
		return ErrInvalidParam
	}
	group.Items = items

	gByte, err := group.Byte()
	if err != nil {
		return err
	}
	return g.cluster.Update([]byte(AuthBuck), getGKey(gName), gByte)
}

// TODO: manager check
func (g *Group) removeGroup(gName string) ([]string, error) {
	group, err := g.GetGroup(gName)
	if err != nil {
		return nil, err
	}

	managerAndMember := []string{}
	managerAndMember = append(managerAndMember, group.Managers...)
	managerAndMember = append(managerAndMember, group.Members...)
	return managerAndMember, g.cluster.RemoveKey([]byte(AuthBuck), getGKey(gName))
}

func (g *Group) UpdateGroupMember(gName string, addManagers, addMembers, removeManagers, removeMembers []string) (model.Row, error) {
	updateRow := model.Row{}
	group, err := g.GetGroup(gName)
	if err != nil {
		return updateRow, err
	}

	for _, username := range addManagers {
		group.Managers, _ = common.AddIfNotContain(group.Managers, username)
	}
	for _, username := range removeManagers {
		group.Managers, _ = common.RemoveIfContain(group.Managers, username)
	}
	for _, username := range addMembers {
		group.Members, _ = common.AddIfNotContain(group.Members, username)
	}
	for _, username := range removeMembers {
		group.Members, _ = common.RemoveIfContain(group.Members, username)
	}

	newGroupByte, err := group.Byte()
	if err != nil {
		return updateRow, err
	}

	updateRow = model.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: newGroupByte}
	return updateRow, nil
}
