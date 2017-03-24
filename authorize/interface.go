package authorize

import (
	"sync"

	"github.com/lodastack/registry/model"
)

var rootNode = "loda"

type GroupInf interface {
	// ListGroup return the group which name have the prefix.
	ListNsGroup(ns string) ([]Group, error)

	// GetGroup return the group.
	GetGroup(gName string) (Group, error)

	// UpdateGroup update the group.
	UpdateItems(gName string, items []string) error
}

type UserInf interface {
	// get user.
	GetUser(username string) (User, error)

	// GetUserList return a map[string]User,
	// key is username and value is User.
	GetUserList(usernames []string) (map[string]User, error)

	// create/update group.
	SetUser(username, mobile string) error

	// Check whether user exist or not.
	CheckUserExist(username string) (bool, error)
}

type Perm interface {
	// user interface
	UserInf

	// group interface
	GroupInf

	AdminGroupItems(rootNode string) []string

	// check whether one query has the permission.
	Check(username, ns, resource, method string) (bool, error)

	// InitGroup init default/admin group and default user.
	InitGroup(rootNode string) error

	// CreateGroup create a group.
	CreateGroup(gName string, managers, members, items []string) error

	// UpdateGroupMember update group member and user groups.
	UpdateMember(group string, manager []string, members []string, action string) error

	// remove group.
	RemoveUser(username string) error

	// RemoveGroup remove the group.
	RemoveGroup(gName string) error
}

// Cluster is the interface op must implement.
type Cluster interface {
	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// RemoveKey removes the key from the bucket.
	RemoveKey(bucket, key []byte) error

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error
}

func NewPerm(cluster Cluster) (Perm, error) {
	if err := cluster.CreateBucketIfNotExist([]byte(AuthBuck)); err != nil {
		return nil, err
	}
	p := perm{
		sync.RWMutex{},
		Group{cluster: cluster},
		User{cluster: cluster},
		cluster,
	}
	// TODO: get rootNode by param.
	return &p, p.InitGroup(rootNode)
}
