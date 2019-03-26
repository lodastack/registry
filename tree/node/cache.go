package node

import (
	"fmt"
	"sync"
	"unsafe"
)

var (
	nodePool map[string]*Node
	mu       sync.RWMutex
)

func init() {
	nodePool = make(map[string]*Node, 0)
}

func getAllNodeByByte(b []byte) (*Node, error) {
	allNode := &Node{}
	// black magic: transform []byte to string
	allNodeString := *(*string)(unsafe.Pointer(&b))

	// get node by string from cache
	mu.RLock()
	cacheNode, exist := nodePool[allNodeString]
	mu.RUnlock()

	if !exist {
		// unmarshal the []byte to node, and save to cache
		if err := allNode.UnmarshalJSON(b); err != nil {
			return nil, fmt.Errorf("unmarshal node fail: %v", err)
		}
		mu.Lock()
		if cacheNode, exist = nodePool[allNodeString]; !exist {
			nodePool[allNodeString] = allNode
		} else {
			allNode = cacheNode
		}
		mu.Unlock()
	} else {
		allNode = cacheNode
	}
	// return the copy of node
	return (&Node{}).Copy(allNode), nil
}
