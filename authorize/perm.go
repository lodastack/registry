package authorize

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lodastack/registry/model"
)

var (
	AuthBuck = "authorize"

	defaultUser  = "default"
	defaultGroup = "default"

	defaultGid     = ""
	defaultManager = []string{"loda-manager"}
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
			// if has the perm of the ns or its parent, pass.
			if strings.HasSuffix(q, item) {
				return true, nil
			}
		}
	}
	return false, nil
}

// default group has get permission of all resource.
func defaultGroupItems() []string {
	items := make([]string, len(model.Templates))
	for index, res := range model.Templates {
		items[index] = "ns-" + res + "-GET"
	}
	return items
}

// init default group.
func (p perm) InitDefault() error {
	u, err := p.GetUser(defaultUser)
	if err != nil && err != ErrUserNotFound {
		return errors.New("get default user fail: " + err.Error())
	}

	switch len(u.GroupIDs) {
	case 0:
		defaultGid, err = p.SetGroup("", defaultManager, defaultGroupItems())
		if err != nil {
			err = errors.New("get default group fail: " + err.Error())
		}
		if err := p.SetUser(defaultUser, []string{defaultGid}, []string{}); err != nil {
			return errors.New("set default user fail: " + err.Error())
		}
	case 1:
		if _, err = p.GetGroup(u.GroupIDs[0]); err != nil {
			err = errors.New("get default group fail: " + err.Error())
		}
		defaultGid = u.GroupIDs[0]
	default:
		err = errors.New(fmt.Sprintf("group of default user is invalid: %v", u))
	}

	return err
}
