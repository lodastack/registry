package authorize

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
)

var (
	AuthBuck = "authorize"

	DefaultUser = "default"
	defaultGid  = ""
	adminGid    = ""
)

type perm struct {
	GroupInf
	UserInf
}

// check whether one query has the permission.
func (p perm) Check(username, ns, resource, method string) (bool, error) {
	u, err := p.GetUser(username)
	if err != nil {
		return false, errors.New("get user fail: " + err.Error())
	}
	if len(u.GroupIDs) == 0 {
		// TODO: log
		return false, errors.New("have no group fail")
	}

	q := ns + "-" + resource + "-" + method
	for _, gid := range u.GroupIDs {
		g, err := p.GetGroup(gid)
		if err != nil {
			// TODO: log
			continue
		}
		for _, item := range g.Items {
			if item == "" {
				continue
			}
			// if has the perm of the ns or its parent, pass.
			if strings.HasSuffix(q, item) {
				return true, nil
			}
		}
	}
	return false, nil
}

// default group has get permission of all resource.
func defaultGroupItems(rootNode string) []string {
	items := make([]string, len(model.Templates))
	for index, res := range model.Templates {
		items[index] = fmt.Sprintf("%s-%s-%s", rootNode, res, "GET")
	}
	return items
}

func adminGroupItems(rootNode string) []string {
	items := make([]string, len(model.Templates)*4)
	for index, res := range model.Templates {
		items[index*4] = fmt.Sprintf("%s-%s-%s", rootNode, res, "GET")
		items[index*4+1] = fmt.Sprintf("%s-%s-%s", rootNode, res, "PUT")
		items[index*4+2] = fmt.Sprintf("%s-%s-%s", rootNode, res, "POST")
		items[index*4+3] = fmt.Sprintf("%s-%s-%s", rootNode, res, "DELETE")
	}
	return items
}

// init default group.
func (p perm) InitGroup(rootNode string) error {
	u, err := p.GetUser(DefaultUser)
	if err != nil && err != ErrUserNotFound {
		return errors.New("get default user fail: " + err.Error())
	}

	if len(u.GroupIDs) == 2 {
		return p.readDefaultGroup(u.GroupIDs)
	}
	return p.setDefaultGroup()
}

// Read default/admin group from default user.
func (p perm) readDefaultGroup(defaulUserGroup []string) error {
	if _, err := p.GetGroup(defaulUserGroup[0]); err != nil {
		return err
	}
	if _, err := p.GetGroup(defaulUserGroup[1]); err != nil {
		return err
	}
	defaultGid = defaulUserGroup[0]
	adminGid = defaulUserGroup[1]
	return nil
}

// Set default/admin group and set to default user.
func (p perm) setDefaultGroup() error {
	var err error
	defaultGid, err = p.SetGroup("", config.C.Admins, defaultGroupItems(rootNode))
	if err != nil {
		return err
	}
	adminGid, err = p.SetGroup("", config.C.Admins, adminGroupItems(rootNode))
	if err != nil {
		return err
	}
	return p.SetUser(DefaultUser, []string{defaultGid, adminGid}, []string{})
}
