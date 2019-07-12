package authorize

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
	m "github.com/lodastack/store/model"
)

var (
	// AuthBuck is the bucket id to save group and authorize data.
	AuthBuck = "authorize"

	// DefaultUser is a user. no use
	DefaultUser = "loda-defaultuser"

	// DefaultGName is the group name has all read permission of all node..
	DefaultGName = "defaultgroup"

	// AdminGName is the group name has all permission of all node.
	AdminGName       = "admingroup"
	lodaDefaultGName = "loda-" + DefaultGName
	lodaAdminGName   = "loda-" + AdminGName
)

type perm struct {
	sync.RWMutex
	Group
	User
	cluster Cluster
}

// Check one query has the permission or not.
func (p *perm) Check(username, ns, resource, method, uri string) (bool, error) {
	u, err := p.GetUser(username)
	if err != nil {
		return false, errors.New("get user fail: " + err.Error())
	}
	if len(u.Groups) == 0 {
		// TODO: log
		return false, errors.New("have no group fail")
	}

	q := ns + "-" + resource + "-" + method
	if uri == "/api/v1/resource" && method == "GET" && resource == "deploy" {
		q = ns + "-" + resource + "-" + "PUT"
	}

	for _, gName := range u.Groups {
		// only support groups which has "-op" suffix
		// can update deploy resource.
		if uri == "/api/v1/resource" && method == "PUT" && resource == "deploy" && !strings.HasSuffix(gName, "-op") {
			continue
		}

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

// DefaultGroupItems return the item of default group.
// default user could get all resource,
// could get/post/put/delete the group which user is the group manager.
func (p *perm) DefaultGroupItems(ns string) []string {
	items := make([]string, len(model.Templates)+3)
	for index, res := range model.Templates {
		items[index] = fmt.Sprintf("%s-%s-%s", ns, res, "GET")
	}
	return items
}

// AdminGroupItems return the items of admin group.
func (p *perm) AdminGroupItems(ns string) []string {
	items := make([]string, len(model.Templates)*4)
	for index, res := range model.Templates {
		items[index*4] = fmt.Sprintf("%s-%s-%s", ns, res, "GET")
		items[index*4+1] = fmt.Sprintf("%s-%s-%s", ns, res, "PUT")
		items[index*4+2] = fmt.Sprintf("%s-%s-%s", ns, res, "POST")
		items[index*4+3] = fmt.Sprintf("%s-%s-%s", ns, res, "DELETE")
	}
	return items
}

// InitGroup createIfNotExist the default user and admin/default group.
func (p *perm) InitGroup(rootNode string) error {
	if err := p.SetUser(DefaultUser, "", "enable", ""); err != nil {
		return err
	}
	// create admin user if not exist.
	for _, admin := range config.C.CommonConf.Admins {
		if err := p.createUserIfNotExist(admin); err != nil {
			return err
		}
	}

	return p.checkDefaultGroup()
}

// checkDefaultGroup set default/admin group and set to default user.
func (p *perm) checkDefaultGroup() error {
	g := Group{
		GName:    lodaDefaultGName,
		Managers: config.C.CommonConf.Admins,
		Members:  config.C.CommonConf.Admins,
		Items:    p.DefaultGroupItems(rootNode)}
	g.Items = append(g.Items, fmt.Sprintf("%s-%s-%s", rootNode, "user", "PUT"))
	if err := p.createGroupIfNotExist(g); err != nil {
		fmt.Printf("init default group error: %s\n", err.Error())
		return err
	}

	g = Group{
		GName:    lodaAdminGName,
		Managers: config.C.CommonConf.Admins,
		Members:  config.C.CommonConf.Admins,
		Items:    p.AdminGroupItems(rootNode)}
	if err := p.createGroupIfNotExist(g); err != nil {
		fmt.Printf("init admin group error: %s\n", err.Error())
		return err
	}

	return nil
}

func (p *perm) createGroupIfNotExist(g Group) error {
	_, err := p.GetGroup(g.GName)
	if err == nil {
		return nil
	}
	if err != common.ErrGroupNotFound {
		return err
	}
	return p.CreateGroup(g.GName, g.Managers, g.Members, g.Items)

}

func (p *perm) createUserIfNotExist(username string) error {
	_, err := p.GetUser(username)
	if err == nil {
		return nil
	}
	if err != common.ErrUserNotFound {
		return err
	}
	return p.SetUser(username, "", "enable", "")
}

func (p *perm) CreateGroup(gName string, managers, members, items []string) error {
	updateRows := []m.Row{}
	p.Lock()
	defer p.Unlock()

	cgroupRow, err := p.createGroup(gName, managers, members, items)
	if err != nil {
		return err
	}
	updateRows = append(updateRows, cgroupRow)

	for _, username := range managers {
		if username == "" {
			continue
		}
		uRows, err := p.UpdateUser(username, gName, "")
		if err != nil {
			return err
		}
		updateRows = append(updateRows, uRows)
	}
	for _, username := range members {
		if username == "" {
			continue
		}
		uRows, err := p.UpdateUser(username, gName, "")
		if err != nil {
			return err
		}
		updateRows = append(updateRows, uRows)
	}
	return p.cluster.Batch(updateRows)
}

func (p *perm) UpdateMember(gName string, managers, members []string) error {
	updateRows := []m.Row{}
	p.Lock()
	defer p.Unlock()

	group, err := p.GetGroup(gName)
	if err != nil {
		return err
	}

	removeManagers := genRemoveUsers(group.Managers, managers)
	removeMembers := genRemoveUsers(group.Members, members)
	removeUsers := append(removeManagers, removeMembers...)

	addManagers := genAddUsers(group.Managers, managers)
	addMembers := genAddUsers(group.Members, members)
	addUsers := append(addManagers, addMembers...)

	for _, username := range removeUsers {
		if username == "" {
			continue
		}
		uRows, err := p.UpdateUser(username, "", gName)
		if err != nil {
			return err
		}
		updateRows = append(updateRows, uRows)
	}

	for _, username := range addUsers {
		if username == "" {
			continue
		}
		uRows, err := p.UpdateUser(username, gName, "")
		if err != nil {
			return err
		}
		updateRows = append(updateRows, uRows)
	}

	gRows, err := p.SetGroupMember(gName, managers, members)
	if err != nil {
		return err
	}
	updateRows = append(updateRows, gRows)

	return p.cluster.Batch(updateRows)
}

// RemoveUser remove group and update the groups of manger/member.
func (p *perm) RemoveUser(username string) error {
	groups, err := p.UserRemoveUser(username)
	if err != nil {
		return err
	}

	updateGroupRows := []m.Row{}
	for _, gName := range groups {
		udpateRow, err := p.UpdateGroupMember(gName, []string{}, []string{},
			[]string{username}, []string{username})
		if err != nil {
			return err
		}
		updateGroupRows = append(updateGroupRows, udpateRow)
	}
	return p.cluster.Batch(updateGroupRows)
}

// RemoveGroup remove the group.
func (p *perm) RemoveGroup(gName string) error {
	userList, err := p.removeGroup(gName)
	if err != nil {
		return err
	}

	updateGroupRows := []m.Row{}
	for _, username := range userList {
		udpateRow, err := p.UpdateUser(username, "", gName)
		if err != nil {
			return err
		}
		updateGroupRows = append(updateGroupRows, udpateRow)
	}

	if len(updateGroupRows) == 0 {
		return nil
	}
	return p.cluster.Batch(updateGroupRows)
}

// genAddUsers return the user list which exist in newUsers but not in oldUsers.
func genAddUsers(oldUsers, newUsers []string) []string {
	addUsers := make([]string, len(newUsers))
	var addCnt int
	for _, user := range newUsers {
		if _, contain := common.ContainString(oldUsers, user); !contain {
			addUsers[addCnt] = user
			addCnt++
		}
	}

	return addUsers[:addCnt]
}

// genRemoveUsers return the user list which exist in oldUsers but not in newUsers.
func genRemoveUsers(oldUsers, newUsers []string) []string {
	removeUsers := make([]string, len(oldUsers))
	var removeCnt int
	for _, user := range oldUsers {
		if _, contain := common.ContainString(newUsers, user); !contain {
			removeUsers[removeCnt] = user
			removeCnt++
		}
	}

	return removeUsers[:removeCnt]
}
