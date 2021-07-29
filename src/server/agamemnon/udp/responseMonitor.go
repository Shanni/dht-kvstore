package udp

import (
	pb "agamemnon/pb/protobuf"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type monitorValue struct {
	notify chan *pb.KVResponse
}

var monitorMap = map[string]*monitorValue{}
var monitorMutex = sync.Mutex{}

// WaitForResponse waits for response with msgID until Time.Now() + timeout.
// Blocks until response is notified as received OR until expiry (see NotifyResponseReceived).
// If response is not notified as received until expiry, stops waiting for response.
//
// Arguments:
//      msgId - id of the response message that is expected
//      timeout - timeout duration
// Returns:
//      []byte response if received
//      error otherwise (either write or timeout)
func WaitForResponse(msgID []byte, timeout time.Duration) (*pb.KVResponse, error) {
	monitorMutex.Lock()
	msgIdHex := hex.EncodeToString(msgID)
	expiry := time.Now().Add(timeout).UnixNano()
	var notify chan *pb.KVResponse
	if prevVal := monitorMap[msgIdHex]; prevVal != nil {
		notify = prevVal.notify
	} else {
		notify = make(chan *pb.KVResponse, 1)
		monitorMap[msgIdHex] = &monitorValue{notify: notify}
	}
	monitorMutex.Unlock()
	for {
		select {
		case res := <-notify:
			return res, nil
		default:
			if time.Now().UnixNano() > expiry {
				monitorMutex.Lock()
				delete(monitorMap, msgIdHex)
				monitorMutex.Unlock()
				return nil, fmt.Errorf("no response received until timeout")
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// NotifyResponseReceived notifies channel that response rawMsg with msgID is received.
// If response with such msgID was expected, removes it and notifies.
// Otherwise does nothing.
//
// Arguments:
//      msgId - id of the response message that is expected
//      resPay - unmarshaled response
func NotifyResponseReceived(msgID []byte, resPay *pb.KVResponse) {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	msgIdHex := hex.EncodeToString(msgID)
	if value := monitorMap[msgIdHex]; value != nil {
		delete(monitorMap, msgIdHex)
		value.notify <- resPay
	}
}
