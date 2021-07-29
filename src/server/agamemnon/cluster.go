package agamemnon

import (
	//"agamemnon/src/server/agamemnon/str"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

type Cluster struct {
	//nodes []*Node
}

// cluster is a slice of nodes *sorted* (by HashCode)
var cluster []*Node
var self *Node

var clusterMutex = &sync.RWMutex{}
var addrMap map[*net.UDPAddr]*Node

func SetSelf(node *Node) {
	self = node
}

func InitCluster(nodes []*Node) {
	if self == nil {
		panic(fmt.Errorf("there is no self"))
	}

	cluster = nodes
	updateIndex()

	addrMap = map[*net.UDPAddr]*Node{}
	for _, node := range cluster {
		addrMap[node.Addr] = node
	}
}

func updateTimeStamp(listenFrom *net.UDPAddr) *Node {
	if node, ok := addrMap[listenFrom]; ok {
		node.LastTimeStamp = time.Now()
		return node
	}
	return nil
}

func hash(key []byte) uint32 {
	hash := crc32.NewIEEE()
	_, _ = hash.Write(key)
	return hash.Sum32()
}

// indexOfHashCode returns the index of the node responsible for the hashcode in a consistent hashcode space
func indexOfHashCode(hashCode uint32) int {
	indexOfTheNode := sort.Search(len(cluster), func(i int) bool {
		return cluster[i].HashCode >= hashCode
	})
	if indexOfTheNode == len(cluster) {
		return 0
	} else {
		return indexOfTheNode
	}
}

func updateIndex() {
	for i, _ := range cluster {
		cluster[i].Index = i
	}
}

func GetNodeByIndex(i int) *Node {
	clusterMutex.RLock()
	defer clusterMutex.RUnlock()

	if i >= len(cluster) {
		return nil
	}
	return cluster[i]
}

// NodeForKey returns the Node that should handle the key
func NodeForKey(key []byte) Node {
	return *cluster[indexOfHashCode(hash(key))]
}

func nodeForHashCode(hashCode uint32) Node {
	return *cluster[indexOfHashCode(hashCode)]
}

// HashCodeForIpPort computes hashcode for a ip:port string in a consistent hashcode space
func HashCodeForIpPort(ipPort string) uint32 {
	return hash([]byte(ipPort))
}

func GetNodeByIpPort(ipPort string) *Node {
	for _, node := range cluster {
		if node.ipPort() == ipPort {
			return node
		}
	}
	return nil
}

func RemoveNode(node *Node) {
	clusterMutex.Lock()
	defer clusterMutex.Unlock()

	for i, n := range cluster {
		if n.ipPort() == node.ipPort(){
			cluster = append(cluster[:i], cluster[i+1:]...)
			break
		}
	}
	updateIndex()
}

//// JoinNode adds the node with ipPort to cluster
////
//// Returns:
////      true if node was added; false if node already exist or if encountered unlikely error
//func JoinNode(ipPort string) bool {
//	clusterMutex.Lock()
//	defer clusterMutex.Unlock()
//	nodeHashCode := HashCodeForIpPort(ipPort)
//	existingNode := nodeForHashCode(nodeHashCode)
//	if existingNode.HashCode == nodeHashCode {
//		return false
//	}
//	ip, port, err := str.ParseIpPort(ipPort)
//	if err != nil {
//		//log.Force(log.BuildErr("error parsing ip port", err))
//		return false
//	}
//	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
//	if err != nil {
//		//log.Force(fmt.Sprintf("Error resolving node address: %v", ipPort))
//		return false
//	}
//	newNode := Node{Ip: ip, Port: port, HashCode: nodeHashCode, Addr: addr, IsSelf: false}
//	cluster = append(cluster, &newNode)
//	sort.Slice(cluster, func(i, j int) bool {
//		return cluster[i].HashCode < cluster[j].HashCode
//	})
//	return true
//}

// GetMembershipCount returns the number of nodes in a cluster
func GetMembershipCount() int32 {
	clusterMutex.Lock()
	defer clusterMutex.Unlock()
	return int32(len(cluster))
}

func (n *Node) printNode() {
	log.Println(self.ipPort(), " 🥶        ",len(cluster),"      Node ", n.ipPort(), n.Index, n.HashCode)
}

func PrintCluster()  {
	log.Println(self.ipPort(), "MARK")
	for _,n := range cluster {
		n.printNode()
	}
}
