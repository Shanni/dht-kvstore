package agamemnon

import (
	pb "agamemnon/pb/protobuf"
	"math"
	"math/rand"
	"time"
)

const gossipCount = 3

func startGossipFailure(node *Node) {
	randIndex := generateUniqueRandomIndexes()
	for _, index := range randIndex {
		if index == node.Index || index >= len(cluster) {
			continue
		}

		str := node.Addr.String()
		reqPay := pb.KVRequest{Command: NOTIFY_FAUILURE, NodeIpPort: &str}

		// Goroutine, concurrent process might modify the cluster at the same time.
		if GetNodeByIndex(index) == nil {
			continue
		}
		sendRequestToNodeUUID(reqPay, cluster[index])
	}

	// listen to new node heartbeats
	InitHeartbeatsSendToNode(heartbeatsNodeCount)
	//PrintCluster()
}

func generateUniqueRandomIndexes() []int{
	count := int(math.Min(float64(gossipCount), float64(len(cluster))))
	randArr := []int{}

	rand.Seed(time.Now().UnixNano())
	for i, randomNum := range rand.Perm(len(cluster)) {

		if i >= count {
			return randArr
		}

		if randomNum == self.Index {
			continue
		}

		randArr = append(randArr, randomNum)
	}
	return randArr
}

