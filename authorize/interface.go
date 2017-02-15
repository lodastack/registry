package authorize

import (
	"github.com/lodastack/registry/model"
)

var rootNode = "loda"

type GroupInf interface {
	// GetGroup return the group.
	GetGroup(gName string) (Group, error)

	// CreateGroup create a group.
	CreateGroup(gName string, manager, items []string) error

	// UpdateGroup update the group.
	UpdateGroup(gName string, manager, items []string) error

	// RemoveGroup remove the group.
	RemoveGroup(gName string) error
}

type UserInf interface {
	// get user.
	GetUser(username string) (User, error)

	// create/update group.
	SetUser(username string, groupIDs, dashboard []string) error

	// remove group.
	RemoveUser(username string) error

	// Check whether user exist or not.
	CheckUserExist(username string) (bool, error)
}

type Perm interface {
	// user interface
	UserInf

	// group interface
	GroupInf

	// check whether one query has the permission.
	Check(username, ns, resource, method string) (bool, error)

	// init default group.
	InitGroup(rootNode string) error
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
	ViewPrefix(bucket, keyPrefix []byte) (map[string]string, error)

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error
}

func NewPerm(cluster Cluster) (Perm, error) {
	if err := cluster.CreateBucketIfNotExist([]byte(AuthBuck)); err != nil {
		return nil, err
	}
	p := perm{
		Group{cluster: cluster},
		User{cluster: cluster},
	}
	// TODO: get rootNode by param.
	return &p, p.InitGroup(rootNode)
}
