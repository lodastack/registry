package authorize

import (
	"encoding/json"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"
	m "github.com/lodastack/store/model"
)

// User is the infomation one user has.
type User struct {
	Username string   `json:"username"`
	Mobile   string   `json:"mobile"`
	Alert    string   `json:"alert"`
	Groups   []string `json:"groups"`

	cluster Cluster `json:"-"`
}

func getUKey(gid string) []byte { return []byte("u-" + gid) }

// Byte return the user at []byte format.
func (u *User) Byte() ([]byte, error) {
	return json.Marshal(u)
}

// GetUser return user by username.
func (u *User) GetUser(username string) (User, error) {
	out := User{}
	uByte, err := u.cluster.View([]byte(AuthBuck), getUKey(username))
	if err != nil {
		return out, err
	}
	if len(uByte) == 0 {
		return out, common.ErrUserNotFound
	}
	err = json.Unmarshal(uByte, &out)
	return out, err
}

// GetUserList return user list by username list.
func (u *User) GetUserList(usernames []string) (map[string]User, error) {
	Users := make(map[string]User, len(usernames))
	for _, username := range usernames {
		if username == "" {
			continue
		}
		user, err := u.GetUser(username)
		if err != nil {
			log.Errorf("GetUser %s error: %s", username, user)
			continue
		}
		Users[username] = user
	}

	return Users, nil
}

// CheckUserExist return the username exist or not.
func (u *User) CheckUserExist(username string) (bool, error) {
	if username == "" {
		return false, common.ErrInvalidParam
	}
	if _, err := u.GetUser(username); err != nil {
		if err == common.ErrUserNotFound {
			return false, nil
		}
		log.Errorf("GetUser %s fail: %s", username, err.Error())
		return false, err
	}
	return true, nil
}

// SetUser create/update user. But will not init/update groups.
func (u *User) SetUser(username, mobile, alert string) error {
	if username == "" || (alert != "disable" && alert != "enable") {
		return common.ErrInvalidParam
	}

	us, err := u.GetUser(username)
	if err != nil {
		// create a user.
		us.Username = username
		us.Mobile = mobile
		us.Alert = alert
		if _, ok := common.ContainString(config.C.CommonConf.Admins, username); ok {
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

// UserRemoveUser remove the user and from the groups the user has.
func (u *User) UserRemoveUser(username string) ([]string, error) {
	us, err := u.GetUser(username)
	if err != nil {
		return nil, err
	}

	return us.Groups, u.cluster.RemoveKey([]byte(AuthBuck), getUKey(username))
}

// UpdateUser add or remove the user to or from group.
func (u *User) UpdateUser(username string, addGroup string, removeGroup string) (m.Row, error) {
	updateRow := m.Row{}
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

	updateRow = m.Row{Bucket: []byte(AuthBuck), Key: getUKey(username), Value: newUserByte}
	return updateRow, nil
}
