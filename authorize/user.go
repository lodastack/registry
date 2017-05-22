package authorize

import (
	"encoding/json"
	"errors"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidParam = errors.New("invalid param")
)

type User struct {
	Username string   `json:"username"`
	Mobile   string   `json:"mobile"`
	Groups   []string `json:"groups"`

	cluster Cluster `json:"-"`
}

func getUKey(gid string) []byte { return []byte("u-" + gid) }

func (u *User) Byte() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) GetUser(username string) (User, error) {
	out := User{}
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

func (u *User) GetUserList(usernames []string) (map[string]User, error) {
	Users := make(map[string]User, len(usernames))
	for _, username := range usernames {
		user, err := u.GetUser(username)
		if err != nil {
			log.Errorf("GetUser %s error: %s", username, user)
			continue
		}
		Users[username] = user
	}

	return Users, nil
}

func (u *User) CheckUserExist(username string) (bool, error) {
	if username == "" {
		return false, ErrInvalidParam
	}
	if _, err := u.GetUser(username); err != nil {
		if err == ErrUserNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SetUser create/update user. But will not init/update groups.
func (u *User) SetUser(username, mobile string) error {
	if username == "" {
		return ErrInvalidParam
	}

	us, err := u.GetUser(username)
	if err != nil {
		// create a user.
		us.Username = username
		us.Mobile = mobile
		if _, ok := common.ContainString(config.C.Admins, username); ok {
			us.Groups = []string{lodaAdminGName}
		} else {
			us.Groups = []string{lodaDefaultGName}
		}
	} else {
		// update the user.
		if mobile != "" {
			us.Mobile = mobile
		}
	}

	uByte, err := us.Byte()
	if err != nil {
		return err
	}
	return u.cluster.Update([]byte(AuthBuck), getUKey(username), uByte)
}

// RemoveUser will remove user and remove the user from groups.
func (u *User) UserRemoveUser(username string) ([]string, error) {
	us, err := u.GetUser(username)
	if err != nil {
		return nil, err
	}

	return us.Groups, u.cluster.RemoveKey([]byte(AuthBuck), getUKey(username))
}

func (u *User) UpdateUser(username string, addGroup string, removeGroup string) (model.Row, error) {
	updateRow := model.Row{}
	user, err := u.GetUser(username)
	if err != nil {
		return updateRow, err
	}

	if addGroup != "" {
		user.Groups, _ = common.AddIfNotContain(user.Groups, addGroup)

	}
	if removeGroup != "" {
		user.Groups, _ = common.RemoveIfContain(user.Groups, removeGroup)
	}

	newUserByte, err := user.Byte()
	if err != nil {
		return updateRow, err
	}

	updateRow = model.Row{Bucket: []byte(AuthBuck), Key: getUKey(username), Value: newUserByte}
	return updateRow, nil
}
