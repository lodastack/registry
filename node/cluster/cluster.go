package cluster

import (
	m "github.com/lodastack/store/model"
)

// Cluster is the interface op must implement.
type ClusterInf interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// Remove removes a node from the store, specified by addr.
	Remove(addr string) error

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
	Batch(rows []m.Row) error

	// Backup database.
	Backup() ([]byte, error)

	// ViewPrefix returns the value for the keys has the keyPrefix.
	ViewPrefix(bucket, keyPrefix []byte) (map[string][]byte, error)
}

// Get type resType resource of node with ID bucketId.
func GetByte(c ClusterInf, bucket, resType string) ([]byte, error) {
	return c.View([]byte(bucket), []byte(resType))
}

// Set resource to node bucket.
func SetByte(c ClusterInf, bucket, resType string, resByte []byte) error {
	return c.Update([]byte(bucket), []byte(resType), resByte)
}
