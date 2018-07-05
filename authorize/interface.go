package authorize

import (
	"sync"

	m "github.com/lodastack/store/model"
)

var rootNode = "loda"

// GroupInf is interface to manager group.
type GroupInf interface {
	// ListGroup return the group list of under one ns.
	ListNsGroup(ns string) ([]Group, error)

	// GetGroup return the group by group name.
	GetGroup(gName string) (Group, error)

	// UpdateItems update the group permissions.
	UpdateItems(gName string, items []string) error

	// ReadGName return the ns and name of the group.
	ReadGName(gname string) (ns, name string)
}

// UserInf is interface to manager user.
type UserInf interface {
	// GetUser return user by username.
	GetUser(username string) (User, error)

	// GetUserList return a map[string]User,
	// key is username and value is User.
	GetUserList(usernames []string) (map[string]User, error)

	// SetUser create a user with username/mobile.
	SetUser(username, mobile, alert string) error

	// CheckUserExist check the username exist or not.
	CheckUserExist(username string) (bool, error)
}

// Perm is interface to manager authorize.
type Perm interface {
	// user interface
	UserInf

	// group interface
	GroupInf

	// DefaultGroupItems return the default permission of the ns.
	DefaultGroupItems(ns string) []string

	// DefaultGroupItems return the admin permission of the ns.
	AdminGroupItems(ns string) []string

	// Check return the query has the permission or not by ns/resource type/username/method.
	Check(username, ns, resourceType, method, uri string) (bool, error)

	// InitGroup init default/admin group and default user.
	InitGroup(rootNode string) error

	// CreateGroup create a group.
	CreateGroup(gName string, managers, members, items []string) error

	// UpdateMember update group member and the user groups.
	UpdateMember(group string, managers, members []string) error

	// RemoveUser remove user from his all group.
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
	Batch(rows []m.Row) error

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error
}

// NewPerm return interface Perm to manager authorize.
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
