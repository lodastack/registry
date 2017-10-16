package cluster

// store cluster save the tree data include node/resource data.
// the node infomation save in a bucket which include the relationship of this node and their nodeID,
// nodeID used as bucketid to save the node's resource data.

import (
	"github.com/lodastack/store/model"
)

// Inf is the cluster interface the cluster should have.
type Inf interface {
	// Create a bucket, via distributed consensus.
	CreateBucket(name []byte) error

	// Create a bucket via distributed consensus if not exist.
	CreateBucketIfNotExist(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)
}

// GetByte return the resource byte of the nodeID/resourceType.
func GetByte(c Inf, nodeID, resourceType string) ([]byte, error) {
	return c.View([]byte(nodeID), []byte(resourceType))
}

// SetByte set the resource to a node.
func SetByte(c Inf, nodeID, resourceType string, resourceByte []byte) error {
	return c.Update([]byte(nodeID), []byte(resourceType), resourceByte)
}
