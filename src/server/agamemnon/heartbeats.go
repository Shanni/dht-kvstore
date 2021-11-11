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

func InitHeartbeatsSendToNode(n int) {
	lock.Lock()
	defer lock.Unlock()

	sendToNodes = []*Node{}
	TimestampLogs = [][]HeartbeatTimestamp{}

	prev := self.prevNode()
	fmt.Println("Node ", self.Addr.String(), " prev ", prev.Addr.String(), prev.prevNode().Addr.String())

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

			fmt.Println(self.Addr.String(), "SENNNNNNNNNNNNND", node.Addr.String())
			heartbeat := HeartbeatTimestamp{msgId: msg, time: time.Now()}

			lock.Lock()
			TimestampLogs[i] = append(TimestampLogs[i], heartbeat)
			lock.Unlock()

			fmt.Println(self.Addr.String(), "TOOO NODEE", len(TimestampLogs[i]), i, msg, "SEN TO", node.Addr.String())
		}
		// heartbeat interval
		time.Sleep(400 * time.Millisecond)
	}
}

func handleHeartbeats(i int)  {
	maxWaitingTime := 300 * time.Millisecond

	// failure count for each node, we allow up to 3+ failures
	failuresCount := 0
	for {
		for _, timestamp := range TimestampLogs[i] {
			if waitingForResonse(timestamp.msgId, 120 * time.Millisecond) {
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

			if failuresCount >= 3 && time.Now().After(sendToNodes[i].LastTimeStamp.Add(maxWaitingTime)) {
				//claim the node failed
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