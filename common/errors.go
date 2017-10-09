package common

import "errors"

var (
	ErrInitNodeBucket      = errors.New("init node bucket fail")
	ErrInitNodeKey         = errors.New("init node bucket k-v fail")
	ErrGetNode             = errors.New("get node fail")
	ErrNodeNotFound        = errors.New("node not found")
	ErrGetParent           = errors.New("get parent node error")
	ErrCreateNodeUnderLeaf = errors.New("can not create node under leaf node")
	ErrSetResourceToLeaf   = errors.New("can not set resource to leaf node")
	ErrGetNodeID           = errors.New("get nodeid fail")
	ErrInvalidParam        = errors.New("invalid param")
	ErrNilChildNode        = errors.New("get none child node")
	ErrNodeAlreadyExist    = errors.New("node already exist")
	ErrNoLeafChild         = errors.New("have no leaf child node")
	ErrNotAllowDel         = errors.New("not allow to be delete")

	ErrEmptyResource error = errors.New("empty resources")
)
