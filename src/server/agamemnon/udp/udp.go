package udp

import (
	pb "agamemnon/pb/protobuf"
	"net"
	"time"
)

const DefaultRetries = 3
const DefaultTimeout = 100 * time.Millisecond

// WriteToUDPWaitResponse writes to UDP and waits for response
//
// Arguments:
//      conn - connection
//      msgID - id of the message (to wait for response)
//      msg - message to write
//      toAddr - address of the node to write to
//      timeout - timeout
// Returns:
//      KVResponse response if received
//      error otherwise (either write or timeout)
func WriteToUDPWaitResponse(
	conn *net.UDPConn,
	msgID []byte,
	msg []byte,
	toAddr *net.UDPAddr,
	timeout time.Duration,
) (*pb.KVResponse, error) {
	var err error = nil
	for i := 0; i < DefaultRetries; i++ {
		_, err = conn.WriteToUDP(msg, toAddr)
		if err == nil {
			var res *pb.KVResponse
			res, err = WaitForResponse(msgID, timeout)
			if err == nil {
				return res, nil
			}
		}
	}
	return nil, err
}
