package authorize

import (
	"encoding/json"
	"errors"
	"sync"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidParam = errors.New("invalid param")
)

type User struct {
	sync.RWMutex `json:"-"`
	Username     string   `json:"username"`
	GroupIDs     []string `json:"gids"`
	Dashboard    []string `json:"dashboard"`

	cluster Cluster `json:"-"`
}

func getUKey(gid string) []byte { return []byte("u-" + gid) }

func (u User) GetUser(username string) (User, error) {
	out := User{}
	u.RLock()
	defer u.RUnlock()
	uByte, err := u.cluster.View([]byte(AuthBuck), getUKey(username))
	if err != nil {
		return out, err
	}
	if len(uByte) == 0 {
		return out, ErrUserNotFound
	}
	err = json.Unmarshal(uByte, &out)
	return out, err
}

func (u User) SetUser(username string, groupIDs, dashboard []string) error {
	if username == "" {
		return ErrInvalidParam
	}

	us, err := u.GetUser(username)
	if err != nil {
		us.Username = username
		us.GroupIDs = groupIDs
		us.Dashboard = dashboard
	} else {
		if len(groupIDs) != 0 {
			us.GroupIDs = groupIDs
		}
		if len(dashboard) != 0 {
			us.Dashboard = dashboard
		}
	}

	u.Lock()
	defer u.Unlock()
	uByte, err := json.Marshal(us)
	if err != nil {
		return err
	}
	return u.cluster.Update([]byte(AuthBuck), getUKey(username), uByte)
}

func (u User) RemoveUser(username string) error {
	u.Lock()
	defer u.Unlock()
	return u.cluster.RemoveKey([]byte(AuthBuck), getUKey(username))
}
