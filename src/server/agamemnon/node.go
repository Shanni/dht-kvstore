package agamemnon

import (
	"net"
	"time"
)

type Node struct {
	HashCode uint32
	Addr     *net.UDPAddr
	IsSelf   bool
	Index    int

	LastTimeStamp time.Time
}

// Self returns the Node that represents the current node
func Self() Node {
	return *self
}

func (node *Node) nextNode() *Node {
	nextIndex := (node.Index + 1) % len(cluster)
	return cluster[nextIndex]
}

func (node *Node) prevNode() *Node {
	prevIndex := (node.Index - 1 + len(cluster)) % len(cluster)
	return cluster[prevIndex]
}

// index returns the index of the node in cluster
func (n *Node) getIndex() int {
	return n.Index
}
