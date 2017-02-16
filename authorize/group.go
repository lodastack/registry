package authorize

import (
	"encoding/json"
	"errors"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
)

var (
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupAlreadyExist = errors.New("group already exist")
)

type Group struct {
	GName   string   `json:"name"`
	Manager []string `json:"manager"`
	Member  []string `json:"member"`
	Items   []string `json:"items"`

	cluster Cluster `json:"-"`
}

func getGKey(gName string) []byte { return []byte("g-" + gName) }

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

func (g *Group) CreateGroup(gName string, items []string) error {
	if gName == "" {
		return ErrInvalidParam
	}
	_, err := g.GetGroup(gName)
	if err != ErrGroupNotFound {
		if err == nil {
			return ErrGroupAlreadyExist
		}
		return err
	}

	gByte, err := (&Group{
		GName: gName,
		Items: items,
	}).Byte()
	if err != nil {
		return err
	}
	return g.cluster.Update([]byte(AuthBuck), getGKey(gName), gByte)
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
func (g *Group) GroupRemoveGroup(gName string) ([]string, error) {
	group, err := g.GetGroup(gName)
	if err != nil {
		return nil, err
	}

	managerAndMember := []string{}
	managerAndMember = append(managerAndMember, group.Manager...)
	managerAndMember = append(managerAndMember, group.Member...)
	return managerAndMember, g.cluster.RemoveKey([]byte(AuthBuck), getGKey(gName))
}

func (g *Group) CreateIfNotExist(group Group) (bool, error) {
	_, err := g.GetGroup(group.GName)
	if err == ErrGroupNotFound {
		err := g.CreateGroup(group.GName, group.Items)
		return err == nil, err
	}
	return false, err
}

func (g *Group) UpdateGroupMember(gName string, addManagers, addMembers, removeManagers, removeMembers []string) (model.Row, error) {
	updateRow := model.Row{}
	group, err := g.GetGroup(gName)
	if err != nil {
		return updateRow, err
	}

	for _, username := range addManagers {
		group.Manager, _ = common.AddIfNotContain(group.Manager, username)
	}
	for _, username := range removeManagers {
		group.Manager, _ = common.RemoveIfContain(group.Manager, username)
	}
	for _, username := range addMembers {
		group.Member, _ = common.AddIfNotContain(group.Member, username)
	}
	for _, username := range removeMembers {
		group.Member, _ = common.RemoveIfContain(group.Member, username)
	}

	newGroupByte, err := group.Byte()
	if err != nil {
		return updateRow, err
	}

	updateRow = model.Row{Bucket: []byte(AuthBuck), Key: getGKey(gName), Value: newGroupByte}
	return updateRow, nil
}
