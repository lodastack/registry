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

	DefaultUser  = "loda-defaultuser"
	defaultGName = "loda-defaultgroup"
	adminGName   = "loda-admingroup"
)

type perm struct {
	Group
	User
}

// check whether one query has the permission.
func (p perm) Check(username, ns, resource, method string) (bool, error) {
	u, err := p.GetUser(username)
	if err != nil {
		return false, errors.New("get user fail: " + err.Error())
	}
	if len(u.Groups) == 0 {
		// TODO: log
		return false, errors.New("have no group fail")
	}

	q := ns + "-" + resource + "-" + method
	for _, gName := range u.Groups {
		g, err := p.GetGroup(gName)
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

// defaultGroupItems return the item of default group.
// default user could get all resource,
// could get/post/put/delete the group which user is the group manager.
func defaultGroupItems(rootNode string) []string {
	items := make([]string, len(model.Templates)+3)
	for index, res := range model.Templates {
		items[index] = fmt.Sprintf("%s-%s-%s", rootNode, res, "GET")
	}
	items[len(items)-3] = fmt.Sprintf("%s-%s-%s", rootNode, "group", "PUT")
	items[len(items)-2] = fmt.Sprintf("%s-%s-%s", rootNode, "group", "POST")
	items[len(items)-1] = fmt.Sprintf("%s-%s-%s", rootNode, "group", "DELETE")
	return items
}

// adminGroupItems return the items of admin group.
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

// InitGroup createIfNotExist the default user and admin/default group.
func (p perm) InitGroup(rootNode string) error {
	if err := p.checkDefaultGroup(); err != nil {
		return err
	}
	return p.SetUser(DefaultUser, []string{adminGName, defaultGName}, []string{})
}

// Set default/admin group and set to default user.
func (p perm) checkDefaultGroup() error {
	g := Group{
		GName:   defaultGName,
		Manager: config.C.Admins,
		Member:  config.C.Admins,
		Items:   defaultGroupItems(rootNode)}
	if _, err := p.CreateIfNotExist(g); err != nil {
		fmt.Printf("init default group error: %s\n", err.Error())
		return err
	}

	g = Group{
		GName:   adminGName,
		Manager: config.C.Admins,
		Member:  config.C.Admins,
		Items:   adminGroupItems(rootNode)}
	if _, err := p.CreateIfNotExist(g); err != nil {
		fmt.Printf("init admin group error: %s\n", err.Error())
		return err
	}

	return nil
}
