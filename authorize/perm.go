package authorize

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/model"
)

var (
	AuthBuck = "authorize"

	DefaultUser  = "loda-defaultuser"
	defaultGName = "loda-defaultgroup"
	adminGName   = "loda-admingroup"

	Add    = "add"
	Remove = "remove"
)

type perm struct {
	sync.RWMutex `json:"-"`
	Group
	User
	cluster Cluster `json:"-"`
}

// check whether one query has the permission.
func (p *perm) Check(username, ns, resource, method string) (bool, error) {
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
func (p *perm) DefaultGroupItems(rootNode string) []string {
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
func (p *perm) AdminGroupItems(rootNode string) []string {
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
func (p *perm) InitGroup(rootNode string) error {
	if err := p.SetUser(DefaultUser, []string{}); err != nil {
		return err
	}
	if err := p.checkDefaultGroup(); err != nil {
		return err
	}

	if err := p.UpdateMember(defaultGName, []string{DefaultUser}, []string{DefaultUser}, Add); err != nil {
		return err
	}
	if err := p.UpdateMember(adminGName, []string{DefaultUser}, []string{DefaultUser}, Add); err != nil {
		return err
	}

	for _, admin := range config.C.Admins {
		if admin == "" {
			continue
		}
		p.UpdateMember(adminGName, []string{admin}, []string{admin}, Add)
	}
	return nil
}

// checkDefaultGroup set default/admin group and set to default user.
func (p *perm) checkDefaultGroup() error {
	g := Group{
		GName:   defaultGName,
		Manager: config.C.Admins,
		Member:  config.C.Admins,
		Items:   p.DefaultGroupItems(rootNode)}
	if _, err := p.CreateIfNotExist(g); err != nil {
		fmt.Printf("init default group error: %s\n", err.Error())
		return err
	}

	g = Group{
		GName:   adminGName,
		Manager: config.C.Admins,
		Member:  config.C.Admins,
		Items:   p.AdminGroupItems(rootNode)}
	if _, err := p.CreateIfNotExist(g); err != nil {
		fmt.Printf("init admin group error: %s\n", err.Error())
		return err
	}

	return nil
}

func (p *perm) UpdateMember(gName string, manager []string, members []string, action string) error {
	updateRows := []model.Row{}
	p.Lock()
	defer p.Unlock()

	switch action {
	case Add:
		for _, username := range manager {
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
		if len(updateRows) == 0 {
			return nil
		}
		gRows, err := p.UpdateGroupMember(gName, manager, members, []string{}, []string{})
		if err != nil {
			return err
		}
		updateRows = append(updateRows, gRows)
	case Remove:
		for _, username := range manager {
			if username == "" {
				continue
			}
			uRows, err := p.UpdateUser(username, "", gName)
			if err != nil {
				return err
			}
			updateRows = append(updateRows, uRows)
		}
		for _, username := range members {
			if username == "" {
				continue
			}
			uRows, err := p.UpdateUser(username, "", gName)
			if err != nil {
				return err
			}
			updateRows = append(updateRows, uRows)
		}
		if len(updateRows) == 0 {
			return nil
		}
		gRows, err := p.UpdateGroupMember(gName, []string{}, []string{}, manager, members)
		if err != nil {
			return err
		}
		updateRows = append(updateRows, gRows)
	default:
		return ErrInvalidParam
	}

	return p.cluster.Batch(updateRows)
}

// RemoveUser remove group and update the groups of manger/member.
func (p *perm) RemoveUser(username string) error {
	groups, err := p.UserRemoveUser(username)
	if err != nil {
		return err
	}

	updateGroupRows := []model.Row{}
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
	userList, err := p.GroupRemoveGroup(gName)
	if err != nil {
		return err
	}

	updateGroupRows := []model.Row{}
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
