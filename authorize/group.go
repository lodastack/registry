package authorize

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/lodastack/registry/common"
)

var (
	ErrGroupNotFound = errors.New("group not found")
)

type Group struct {
	sync.RWMutex `json:"-"`
	Id           string   `json:"id"`
	Manager      []string `json:"manager"`
	Items        []string `json:"items"`

	cluster Cluster `json:"-"`
}

func getGKey(gid string) []byte { return []byte("g-" + gid) }

func (g Group) GetGroup(gid string) (Group, error) {
	out := Group{}
	g.RLock()
	defer g.RUnlock()
	if gid == "" {
		return out, ErrInvalidParam
	}
	gByte, err := g.cluster.View([]byte(AuthBuck), getGKey(gid))
	if err != nil {
		return out, err
	}

	if len(gByte) == 0 {
		return out, ErrGroupNotFound
	}
	err = json.Unmarshal(gByte, &out)
	return out, err
}

func (g Group) SetGroup(gid string, manager, items []string) (string, error) {
	var err error
	if len(manager) == 0 && len(items) == 0 {
		return "", ErrInvalidParam
	}

	gs := Group{Id: gid}
	if gs.Id == "" {
		// new group
		gs = Group{
			Id:      common.GenUUID(),
			Manager: manager,
			Items:   items,
		}
	} else {
		// update group
		if gs, err = g.GetGroup(gs.Id); err != nil {
			return "", ErrGroupNotFound
		}
		if len(manager) != 0 {
			gs.Manager = manager
		}
		if len(items) != 0 {
			gs.Items = items
		}
	}
	g.Lock()
	defer g.Unlock()
	gByte, err := json.Marshal(gs)
	if err != nil {
		return "", err
	}
	return gs.Id, g.cluster.Update([]byte(AuthBuck), getGKey(gs.Id), gByte)
}

func (g Group) RemoveGroup(gid string) error {
	g.Lock()
	defer g.Unlock()
	return g.cluster.RemoveKey([]byte(AuthBuck), getGKey(gid))
}
