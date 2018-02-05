package authorize

import (
	"encoding/json"
	"strings"

	"github.com/lodastack/registry/common"
	m "github.com/lodastack/store/model"
)

var (
	// OP group name
	OP = "op"
	// DEV group name
	DEV = "dev"

	nsSep             = "."
	groupNameSep byte = '-'
)

// Group is the infrmation one group has.
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

// ReadGName return the ns and name of group name.
func (g *Group) ReadGName(gname string) (ns, name string) {
	ns, name = readGName(gname)
	ns = reverceNs(ns)
	return
}

// GetNsDevGName return dev group name of ns.
func GetNsDevGName(ns string) string {
	return GetGNameByNs(ns, DEV)
}

// GetNsOpGName return op group name of ns.
func GetNsOpGName(ns string) string {
	return GetGNameByNs(ns, OP)
}

// Byte return group at from of []byte.
func (g *Group) Byte() ([]byte, error) {
	return json.Marshal(g)
}

// GetGroup return group by group name.
func (g *Group) GetGroup(gName string) (Group, error) {
	group := Group{}
	if gName == "" {
		return group, common.ErrInvalidParam
	}
	gByte, err := g.cluster.View([]byte(AuthBuck), getGKey(gName))
	if err != nil {
		return group, err
	}

	if len(gByte) == 0 {
		return group, common.ErrGroupNotFound
	}
	err = json.Unmarshal(gByte, &group)
	return group, err
}

// ListNsGroup return group list of one ns.
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

func (g *Group) createGroup(gName string, managers, members, items []string) (m.Row, error) {
	updateRow := m.Row{}
	if gName == "" {
		return updateRow, common.ErrInvalidParam
	}
	_, err := g.GetGroup(gName)
	if err != common.ErrGroupNotFound {
		if err == nil {
			return updateRow, common.ErrGroupAlreadyExist
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

	updateRow = m.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: gByte}
	return updateRow, nil
}

// UpdateItems group items which manage the group permissions.
func (g *Group) UpdateItems(gName string, items []string) error {
	group, err := g.GetGroup(gName)
	if err != nil {
		return err
	}

	if len(items) == 0 || items[0] == "" {
		return common.ErrInvalidParam
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

// UpdateGroupMember add/remove user from group.
func (g *Group) UpdateGroupMember(gName string, addManagers, addMembers, removeManagers, removeMembers []string) (m.Row, error) {
	updateRow := m.Row{}
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

	updateRow = m.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: newGroupByte}
	return updateRow, nil
}

// SetGroupMember sets user from group.
// keep user custom config order
func (g *Group) SetGroupMember(gName string, managers, members []string) (m.Row, error) {
	updateRow := m.Row{}
	group, err := g.GetGroup(gName)
	if err != nil {
		return updateRow, err
	}

	group.Managers = []string{}
	um := make(map[string]struct{})
	for _, username := range managers {
		if _, ok := um[username]; ok {
			continue
		}
		group.Managers = append(group.Managers, username)
		um[username] = struct{}{}
	}
	group.Members = []string{}
	um = make(map[string]struct{})
	for _, username := range members {
		if _, ok := um[username]; ok {
			continue
		}
		group.Members = append(group.Members, username)
		um[username] = struct{}{}
	}

	newGroupByte, err := group.Byte()
	if err != nil {
		return updateRow, err
	}

	updateRow = m.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: newGroupByte}
	return updateRow, nil
}
