package node

const (
	Leaf = iota // leaf type of node
	Root        // non-leaf type of node
)

type NodeProperty struct {
	Id   string
	Name string
	Type int

	// regexp of machine in one node,
	// used to auto put a machine in one node
	MachineReg string
}

type Node struct {
	NodeProperty
	Clildren []NodeProperty
}
type Tree struct {
	Node    Node
	Cluster Cluster
}

func NewTree(cluster Cluster) *Tree {
	// TODO: read data from boltdb
	return &Tree{
		Node:    Node{NodeProperty{Id: "0", Name: "root", Type: Leaf, MachineReg: "*"}, []NodeProperty{}},
		Cluster: cluster,
	}
}
