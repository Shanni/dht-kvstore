package agamemnon

import (
	pb "agamemnon/pb/protobuf"
	"fmt"
	"sync"
	"time"
)

var heartbeatLogsEnabled = true
const heartbeatsNodeCount = 2 // two nodes to monitor

var sendToNodes []*Node
var TimestampLogs [][]HeartbeatTimestamp

var lock = &sync.Mutex{}

type HeartbeatTimestamp struct {
	msgId []byte
	time time.Time
}

//type Heartbeats struct {
//	sync.Mutex
//	sendToNode *Node
//	timestampLogs []HeartbeatTimestamp
//	failureCount int
//}

func InitHeartbeatsSendToNode(n int) {
	lock.Lock()
	defer lock.Unlock()

	sendToNodes = []*Node{}
	TimestampLogs = [][]HeartbeatTimestamp{}

	prev := self.prevNode()
	fmt.Println("Node ", self.Port, " prev ", prev.Port, prev.prevNode().Port)

	for i := 0; i < n; i++ {
		if prev == self {
			return
		}
		sendToNodes = append(sendToNodes, prev)
		prev = prev.prevNode()

		TimestampLogs = append(TimestampLogs, []HeartbeatTimestamp{})
	}

}

func sendHeartbeats() {
	for{
		for i, node := range sendToNodes{
			reqPay := pb.KVRequest{Command: IS_ALIVE}
			msg := sendRequestToNodeUUID(reqPay, node)

			fmt.Println(self.Port, "SENNNNNNNNNNNNND", node.Port)
			heartbeat := HeartbeatTimestamp{msgId: msg, time: time.Now()}

			lock.Lock()
			TimestampLogs[i] = append(TimestampLogs[i], heartbeat)
			lock.Unlock()

			fmt.Println(self.Port, "TOOO NODEE", len(TimestampLogs[i]), i, msg, "SEN TO", node.ipPort())
		}
		// heartbeat interval
		time.Sleep(400 * time.Millisecond)
	}
}

func handleHeartbeats(i int)  {
	maxWaitingTime := 1500 * time.Millisecond

	// failure count for each node, we allow up to 3+ failures
	failuresCount := 0
	for {
		for _, timestamp := range TimestampLogs[i] {
			fmt.Println("😅😅😅👽👽👽", self.Port,  "sent by", sendToNodes[i].ipPort(), "abouot to check ",timestamp.msgId)
			if waitingForResonse(timestamp.msgId, 50 * time.Millisecond) {
				failuresCount = 0
				lock.Lock()
				TimestampLogs[i] = TimestampLogs[i][1:]
				lock.Unlock()

				continue
			}

			if time.Now().After(timestamp.time.Add(maxWaitingTime)) {
				lock.Lock()
				TimestampLogs[i] = TimestampLogs[i][1:]
				lock.Unlock()

				failuresCount += 1
				break
			}

			fmt.Println("😱😱😱Waiting more rounds....", failuresCount, sendToNodes[i].Port, "reported by ", self.Port)

			if failuresCount >= 3 && time.Now().After(sendToNodes[i].LastTimeStamp.Add(maxWaitingTime)) {
				//claim the node failed
				fmt.Println("No good 😅😅😅😅😅😅 ", sendToNodes[i].Port, "failed")
				RemoveNode(sendToNodes[i])

				if i == 0 {
					RecoverDataStorage()
				} else {
					reqPay := pb.KVRequest{Command: RECOVER_PREV_NODE_KEYSPACE}
					sendRequestToNodeUUID(reqPay, self.prevNode())
				}
				startGossipFailure(sendToNodes[i])
				time.Sleep(time.Second)
				break
			}
		}
		time.Sleep(300 * time.Millisecond)
		//fmt.Println("👽👽👽👅👅", self.Port,  " check sent by", sendToNodes[i].ipPort())
	}
}

func HeartbeatManager() {
	InitHeartbeatsSendToNode(heartbeatsNodeCount)

	for i := 0; i < heartbeatsNodeCount; i++ {
		go handleHeartbeats(i)
	}
	go sendHeartbeats()

	go func() {
		PrintCluster()
		time.Sleep(5*time.Second)
	}()
}