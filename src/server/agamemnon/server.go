package agamemnon

import (
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"hash/crc32"
	"log"
	"net"
	"os"
	"time"

	pb "agamemnon/pb/protobuf"
	proto "github.com/golang/protobuf/proto"
)

// List of errors that can be returned to client
const (
	NO_ERR           = 0x00
	KEY_DNE_ERR      = 0x01
	NO_SPC_ERR       = 0x02
	SYS_OVERLOAD_ERR = 0x03
	KV_INTERNAL_ERR  = 0x04
	UNKNOWN_CMD_ERR  = 0x05
	INVALID_KEY_ERR  = 0x06
	INVALID_VAL_ERR  = 0x07
	NEED_REJOIN_ERR  = 0x20
)

// List of commands that can be sent to the server
const (
	PUT                = 0x01
	GET                = 0x02
	REMOVE             = 0x03
	SHUTDOWN           = 0x04
	WIPEOUT            = 0x05

	IS_ALIVE           = 0x06
	GET_PID            = 0x07
	GET_MEMBERSHIP_CNT = 0x08

	ADD_REPLICA        = 0x11   // current Node add replica
	SEND_REPLICA       = 0x12	// current Node to send replica
	RECOVER_PREV_NODE_KEYSPACE = 0x13
	NOTIFY_FAUILURE    = 0x22

	TEST_GOSSIP        = 0x40
	TEST_RECOVER_REPLICA = 0x41
	SIMULATE_FAILURE   = 0X20
)

const bufferSizeBytes = 11000
const readTimeoutMs = 100 * time.Millisecond
const clusterStartDelay = 6 * time.Second
const massPutDelay = 10 * time.Millisecond
const clusterSeparator = ","

// Constant to use for the server overload condition
var overloadWaitTimeMs = int32(5000)

// Global variable to store the UDP connection object so
// that it does not need to be passed to every function
var conn *net.UDPConn

// Channel to signal to the server to quit
var shutdown = make(chan bool)

// Channel to signal the end of mass transfer with incoming Node as a value
var transferEnd = make(chan Node)

// Get the CRC-32 IEEE checksum of the message ID and payload
//
// Arguments:
//     messageID: unique ID of the message
//     payload:   payload of the message
// Returns:
//     CRC-32 of the message ID concatenated with the payload
func getChecksum(messageID []byte, payload []byte) uint64 {
	return uint64(crc32.ChecksumIEEE(append(messageID, payload...)))
}

func buildReqMsgWithMsgId(reqPay pb.KVRequest, msgId []byte) *pb.Msg {
	rawReqPay, _ := proto.Marshal(&reqPay)
	return &pb.Msg{
		MessageID:  msgId,
		Payload:    rawReqPay,
		CheckSum:   getChecksum(msgId, rawReqPay),
		Type: 0, // request
	}
}

func getNetAddress(addr *net.UDPAddr) string {
	return addr.String()
}

// Create and send the response message to the client
//
// Arguments:
//		clientAddr: address to return to the response to
//		msgID: ID of the request
//		respPay: payload to return in the response
func sendResponse(clientAddr *net.UDPAddr, msgID []byte, respPay pb.KVResponse) {
	if self.Addr.String() == clientAddr.String() {
		fmt.Println("Stop. No need to send response to self. ", clientAddr.Port)
		fmt.Println(respPay)
		return
	}

	respPayBytes, err := proto.Marshal(&respPay)
	if err != nil {
		fmt.Println("Could not marshall the payload")
		return
	}

	checkSum := getChecksum(msgID, respPayBytes)

	respMsg := pb.Msg{
		MessageID:  msgID,
		Payload:    respPayBytes,
		CheckSum:   checkSum,
		Type: 1, // response
	}

	respMsgBytes, err := proto.Marshal(&respMsg)
	if err != nil {
		fmt.Println("Could not marshall the message")
		return
	}

	fmt.Println(self.Addr.String(), "🤖ABOUT TO SENT", respMsgBytes, clientAddr.String())
	//Cache the response if it's not a server overload
	if respPay.ErrCode != SYS_OVERLOAD_ERR {
		// If there is no space to add to the cache, send a server
		// overload response instead
		if !responseCache.Add(msgID, getNetAddress(clientAddr), respMsgBytes) {
			respPay.ErrCode = SYS_OVERLOAD_ERR
			respPay.OverloadWaitTime = &overloadWaitTimeMs
			sendResponse(clientAddr, msgID, respPay)
		}
	}

	fmt.Println("sending ", respMsgBytes, " to ", clientAddr.Port)
	// Send the message back to the client
	_, err = conn.WriteToUDP(respMsgBytes, clientAddr)
	if err != nil {
		fmt.Println("sendResponse WriteToUDP", err)
	}

	if _, ok := incomingCache.Delete(msgID); !ok {
		fmt.Println("Error: No Request Found in RequestsCache.", hex.EncodeToString(msgID))
	}
}

// Forward message to the correct node
//
// Arguments:
//		msgID: ID of the request
//		rawMsg: marshaled message
//		toAddr: UPD Addr of node to send it to
func forwardRequest(originatorAddr *net.UDPAddr, msgID []byte, reqPay pb.KVRequest, rawMsg []byte, toNode Node) {
	if self.Addr.String() == toNode.Addr.String() {
		fmt.Println("Stop. Can't forward to self - ", toNode.Addr.String())
		return
	}

	sendRequestToNode(msgID, reqPay, &toNode)
}

func sendRequestToNodeUUID(reqPay pb.KVRequest, toNode *Node) []byte {
	msgId, _ := uuid.New().MarshalBinary()
	sendRequestToNode(msgId, reqPay, toNode)
	return msgId
}

func sendRequestToNode(msgID []byte, reqPay pb.KVRequest, toNode *Node) {
	if toNode == self {
		log.Println("sendRequestToNode Not sending message to self " + toNode.Addr.String())
		return
	}
	msg := buildReqMsgWithMsgId(reqPay, msgID)
	reqMsg, _ := proto.Marshal(msg)
	_, err := conn.WriteToUDP(reqMsg, toNode.Addr)
	if err != nil {
		fmt.Println("sendRequestMsg Err: ", err)
	}
}

// Given a response message from another server, parse the response,
// perform the necessary action (e.g. pass it back to the originator)
//
// Arguments:
//		msgID: message ID of the request
//		resPay unmarshalled payload
//		rawMsg: raw message with response
func handleResponse(srcAddr *net.UDPAddr, msgId []byte, resPay *pb.KVResponse, rawMsg []byte) {
	// check if it's the one off propagate request
	if clientAddr, ok := incomingCache.Delete(msgId); ok {
		if clientAddr == self.Addr {
			fmt.Println("Something is odd.")
		}

		responseCache.Add(msgId, getNetAddress(srcAddr), rawMsg)
		_, err := conn.WriteToUDP(rawMsg, clientAddr)

		fmt.Println("💡💡💡 forwarding response to ", clientAddr.Port, resPay)

		if err != nil {
			log.Println("handleResponse", err)
		}

	}else if resPay.ReceiveData {
		responseCache.Add(msgId, getNetAddress(srcAddr), resPay.Value)
	}else {
		// assume no memory restriction
		fmt.Println("When receive random response...")
		incomingCache.Add(msgId, srcAddr)
	}
}

// return msgId
func requestToReplicaNode(toNode *Node, reqPay pb.KVRequest, memo int) []byte {
	rep := uint32(memo)
	reqPay.ReplicaNum = &rep
	return sendRequestToNodeUUID(reqPay, toNode)
}

// check before sending response
// fault tolerant - wait till timeout (1s), for responds
func waitingForResonse(msgId []byte, duration time.Duration) bool {
	now := time.Now()
	for {
		if val := incomingCache.Get(msgId); val != nil {
			fmt.Println("........got a response.............", hex.EncodeToString(msgId))
			return true
		}
		if time.Now().After(now.Add(duration)) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// send request, and wait for the response to that request
func waitingForResponseData(msgId []byte, memo string, duration time.Duration) []byte {
	now := time.Now()
	for {
		if val := responseCache.Get(msgId, memo); val != nil {
			fmt.Println("REEEEEECEIVE DATA", hex.EncodeToString(msgId))
			return val
		}
		if time.Now().After(now.Add(duration)) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil
}

// Given a request message from a client, parse the request,
// perform the necessary action for the command of the request
// and return a response.
//
// Arguments:
// 		clientAddr: address to return to the response to
//		msgID: message ID of the request
//		reqPay unmarshalled payload
//		rawMsg original message for further processing
func handleRequest(clientAddr *net.UDPAddr, msgID []byte, reqPay pb.KVRequest, rawMsg []byte) {
	if respMsgBytes := responseCache.Get(msgID, getNetAddress(clientAddr)); respMsgBytes != nil {
		fmt.Println("Handle repeated request - 😡", respMsgBytes, "sending to ", clientAddr.Port)

		_, err := conn.WriteToUDP(respMsgBytes, clientAddr)
		if err != nil {
			fmt.Println("handleRequest WriteToUDP", err)
		}
	} else {
		incomingCache.Add(msgID, clientAddr)

		respPay := pb.KVResponse{}
		switch reqPay.Command {
		case PUT:
			fmt.Println("+PUT request come in from", clientAddr.Port)
			node := NodeForKey(reqPay.Key)
			if node.IsSelf && reqPay.ReplicaNum == nil {
				respPay.ErrCode = dataStorage.Replicas[0].Put(reqPay.Key, reqPay.Value, reqPay.Version)

				msgId := requestToReplicaNode(self.nextNode(), reqPay, 1)
				msgId2 := requestToReplicaNode(self.nextNode().nextNode(), reqPay, 2)

				fmt.Println("who's sending responsee 🤡 ", self.Addr.String(), " to ", clientAddr.Port)
				if waitingForResonse(msgId, time.Second) && waitingForResonse(msgId2, time.Second) {
					sendResponse(clientAddr, msgID, respPay)
				} else {
					// TODO: revert primary, send error
				}
			} else if reqPay.ReplicaNum != nil {
				respPay.ErrCode = dataStorage.Replicas[*reqPay.ReplicaNum].Put(reqPay.Key, reqPay.Value, reqPay.Version)
				sendResponse(clientAddr, msgID, respPay)
			} else {
				forwardRequest(clientAddr, msgID, reqPay, rawMsg, node)
			}
		case GET:
			node := NodeForKey(reqPay.Key)
			var version int32
			if node.IsSelf && reqPay.ReplicaNum == nil {
				respPay.Value, version, respPay.ErrCode = dataStorage.Replicas[0].Get(reqPay.Key)
				respPay.Version = &version
				// TODO: check failure, then send request to other two nodes.
				sendResponse(clientAddr, msgID, respPay)
			} else if reqPay.ReplicaNum != nil {

				respPay.Value, version, respPay.ErrCode =  dataStorage.Replicas[*reqPay.ReplicaNum].Get(reqPay.Key)
				sendResponse(clientAddr, msgID, respPay)
			} else {
				forwardRequest(clientAddr, msgID, reqPay, rawMsg, node)
			}
		case REMOVE:
			node := NodeForKey(reqPay.Key)
			if node.IsSelf && reqPay.ReplicaNum == nil {
				respPay.ErrCode = dataStorage.Replicas[0].Remove(reqPay.Key)

				msgId := requestToReplicaNode(self.nextNode(), reqPay, 1)
				msgId2 := requestToReplicaNode(self.nextNode().nextNode(), reqPay, 2)
				if waitingForResonse(msgId, time.Second) && waitingForResonse(msgId2, time.Second){
					sendResponse(clientAddr, msgID, respPay)
				} else {
					// TODO: revert primary, send error (can't revert primary lol)
					fmt.Println("????? can't remove fully??")
				}
			} else if reqPay.ReplicaNum != nil {
				respPay.ErrCode =  dataStorage.Replicas[*reqPay.ReplicaNum].Remove(reqPay.Key)
				sendResponse(clientAddr, msgID, respPay)
			} else {
				forwardRequest(clientAddr, msgID, reqPay, rawMsg, node)
			}
		case SHUTDOWN:
			shutdown <- true
		case WIPEOUT:
			if reqPay.ReplicaNum != nil {
				dataStorage.Replicas[*reqPay.ReplicaNum].RemoveAll()
			} else {
				respPay.ErrCode = dataStorage.Replicas[0].RemoveAll()
				dataStorage.Replicas[1].RemoveAll()
				dataStorage.Replicas[2].RemoveAll()
			}
			sendResponse(clientAddr, msgID, respPay)
		case IS_ALIVE:
			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case GET_PID:
			pid := int32(os.Getpid())
			respPay.Pid = &pid
			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case GET_MEMBERSHIP_CNT:
			members := GetMembershipCount()
			respPay.MembershipCount = &members

			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case NOTIFY_FAUILURE:
			failedNode := GetNodeByIpPort(*reqPay.NodeIpPort)
			if failedNode != nil {
				fmt.Println(self.Addr.String(), " STARTT CONTIUE GOSSSSSSIP 👻💩💩💩💩💩🤢🤢🤢🤢", *reqPay.NodeIpPort, "failed")
				RemoveNode(failedNode)
				startGossipFailure(failedNode)
			}
			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case ADD_REPLICA:
			kv := dataStorage.decompressReplica(reqPay.Value)
			dataStorage.addReplica(kv, int(*reqPay.ReplicaNum))

			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case SEND_REPLICA:
			respPay.Value = dataStorage.compressReplica(int(*reqPay.ReplicaNum))
			respPay.ReceiveData = true

			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case RECOVER_PREV_NODE_KEYSPACE:
			// TODO: error handling on and internal failure
			RecoverDataStorage()

			respPay.ErrCode = NO_ERR
			sendResponse(clientAddr, msgID, respPay)
		case TEST_GOSSIP:
			fmt.Println(self.Addr.String(), " TESTING GOSSIP 😡", *reqPay.NodeIpPort, "failed")
			RemoveNode(GetNodeByIpPort("127.0.0.1:3331"))
			startGossipFailure(GetNodeByIpPort("127.0.0.1:3331"))
		case TEST_RECOVER_REPLICA:
			reqPay := pb.KVRequest{Command: SHUTDOWN}
			sendRequestToNodeUUID(reqPay, self.prevNode())
			RemoveNode(self.prevNode())

			RecoverDataStorage()
		default:
			//respPay.ErrCode = UNKNOWN_CMD_ERR
			//sendResponse(clientAddr, msgID, respPay)
		}
	}
	printReplicas(self.Addr.String())
}

// Unmarshals a request from a clients
//
// Arguments:
//		buf: buffer containing the request bytes
// Returns:
//		Unmarshaled KVRequest object
//		Unmarshaled KVResponse object
//		Unmarshaled Msg object
//      error if any errors encountered
func unmarshalMessage(buf []byte) (*pb.KVRequest, *pb.KVResponse, *pb.Msg, error) {
	reqMsg := pb.Msg{}
	reqPay := pb.KVRequest{}
	resPay := pb.KVResponse{}

	err := proto.Unmarshal(buf, &reqMsg)
	if err != nil {
		fmt.Println(self.Addr.String(), "unmarshalMessage Failed to unmarshal ", len(buf))
		return nil, nil, nil, err
	}

	checkSum := getChecksum(reqMsg.MessageID, reqMsg.Payload)
	if checkSum != reqMsg.CheckSum {
		return nil, nil, nil, fmt.Errorf("wrong checksum")
	}

	if reqMsg.Type == 1 {
		// response
		err = proto.Unmarshal(reqMsg.Payload, &resPay)
		if err == nil {
			return nil, &resPay, &reqMsg, nil
		} else {
			fmt.Println("unmarshalMessage, failed to unmarshal response ", err)
		}
	} else if reqMsg.Type == 0 {
		// request
		err = proto.Unmarshal(reqMsg.Payload, &reqPay)
		if err == nil {
			return &reqPay, nil, &reqMsg, nil
		}else {
			fmt.Println("unmarshalMessage, failed to unmarshal request ", err)
		}
	} else if reqMsg.Type == 2 {
		// ack
	}
	return nil, nil, nil, err
}

// Starts the UDP server. Waits for a message to be received
// and spawns a goroutine to handle the client's request
//
// Arguments:
//		cluster: description of cluster topology
//              port: servers own port to listen on
func StartServer(port int, cluster []*Node) {
	InitCluster(cluster)

	go responseCache.TTLManager()
	go incomingCache.TTLManager()
	InitReplicas()

	//log.InitLogs(fmt.Sprintf("%v", Self().Port))

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Error resolving UDP server address", err)
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("Error starting listening to UDP", err)
	}
	defer conn.Close()

	fmt.Println("Server started UDP")

	fmt.Println(self.Addr.String(), " has", len(cluster))
	HeartbeatManager()

	rcvBuffer := make([]byte, bufferSizeBytes)
	for {
		select {
		case <-shutdown:
			fmt.Println("Server shutdown")
			return
		default:
			deadline := time.Now().Add(readTimeoutMs)
			_ = conn.SetReadDeadline(deadline)
			numBytes, clientAddr, err := conn.ReadFromUDP(rcvBuffer)

			updateTimeStamp(clientAddr) // could pass down node.

			if err == nil {
				rawMsg := rcvBuffer[:numBytes]
				req, resp, msg, err := unmarshalMessage(rawMsg)
				fmt.Println("BIG NEWS!", self.Addr.String(), " recieved ", msg.MessageID, numBytes, " from ", clientAddr.Port)
				if err == nil && msg.Type == 1 {
					//Hangle response
					//fmt.Println("Ever handle response?", clientAddr.Port, clientAddr.IP)
					//fmt.Println(resp)
					go handleResponse(clientAddr, msg.MessageID, resp, rawMsg)
				} else if err == nil && msg.Type == 2 {
					fmt.Println("It's an ACK ~")

				} else if err == nil && msg.Type == 0 {
					go handleRequest(clientAddr, msg.MessageID, *req, rawMsg)
				} else {
					fmt.Println("StartServer Failed unmarshall message: ", err, rawMsg)
				}
			}
		}
	}
}
