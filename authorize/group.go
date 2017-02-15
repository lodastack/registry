package authorize

import (
	"encoding/json"
	"errors"
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

// update set the infomation to group.
func (g *Group) update(manager, items []string) error {
	if len(manager) == 0 && len(items) == 0 {
		return ErrInvalidParam
	}
	if len(manager) != 0 && manager[0] != "" {
		g.Manager = manager
	}
	if len(items) != 0 && items[0] != "" {
		g.Items = items
	}
	return nil
}

func (g *Group) getUpdatedByte(manager, items []string) ([]byte, error) {
	err := g.update(manager, items)
	if err != nil {
		return nil, err
	}
	return g.Byte()
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

// TODO: set batch to update group member
func (g *Group) CreateGroup(gName string, manager, items []string) error {
	if gName == "" || len(manager) == 0 || manager[0] == "" {
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
		GName:   gName,
		Manager: manager,
		Items:   items,
	}).Byte()
	if err != nil {
		return err
	}
	return g.cluster.Update([]byte(AuthBuck), getGKey(gName), gByte)
}

// UpdateGroup update, not update member infomation.
// TODO:
//    1. manager check
//    2. batch
func (g *Group) UpdateGroup(gName string, manager, items []string) error {
	group, err := g.GetGroup(gName)
	if err != nil {
		return err
	}

	gByte, err := group.getUpdatedByte(manager, items)
	if err != nil {
		return err
	}
	return g.cluster.Update([]byte(AuthBuck), getGKey(gName), gByte)
}

// TODO: batch to update user
func (g *Group) RemoveGroup(gName string) error {
	return g.cluster.RemoveKey([]byte(AuthBuck), getGKey(gName))
}

func (g *Group) CreateIfNotExist(group Group) (bool, error) {
	_, err := g.GetGroup(group.GName)
	if err == ErrGroupNotFound {
		err := g.CreateGroup(group.GName, group.Manager, group.Items)
		return err == nil, err
	}
	return false, err
}

func (g *Group) UpdateMember(GName string, addUsername []string, removeUsername []string) ([]byte, error) {
	return nil, nil
}
