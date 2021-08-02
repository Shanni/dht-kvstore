package main

import (
	pb "agamemnon/pb/protobuf"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"hash/crc32"
	"log"
	"net"
	"time"
)

type input struct {
	code    uint32
	key     string
	val     []byte
	version int32
}

func sendMessage(payload []byte, conn *net.UDPConn, messageId []byte) error {

	checksum := crc32.ChecksumIEEE(append(messageId, payload...))

	msg := &pb.Msg{
		MessageID: messageId,
		Payload:   payload,
		CheckSum:  uint64(checksum),
	}
	//fmt.Println("sending message: ", *msg)
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = conn.Write(body)
	if err != nil {
		return err
	}
	return nil
}

func SendRequest(conn *net.UDPConn, messageId []byte, code uint32, key string, value []byte, version int32) (uint32, error) {

	keyByte, _ := hex.DecodeString(key)
	reqPayload := &pb.KVRequest{
		Command: code,
		Key: keyByte,
		Value: value,
		Version: &version,
	}

	payload, _ := proto.Marshal(reqPayload)
	sendMessage(payload, conn, messageId)

	msg := &pb.Msg{}
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Duration( 200 * time.Millisecond)))
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return 0xff, err
	}
	//fmt.Println("debug ", buffer[:n])
	if err := proto.Unmarshal(buffer[:n], msg); err != nil {
		//log.Println("Failed to parse response message :", err)
		return 0xff, err
	}

	resMessageId := msg.MessageID
	resPayLoad := msg.Payload
	resChecksum := crc32.ChecksumIEEE(append(resMessageId, resPayLoad...))

	// responding checksum isn't match
	if uint64(resChecksum) != msg.CheckSum {
		log.Println("Checksum from server isn't correct: ", resChecksum, msg.CheckSum)
		return 0xff, nil
	}

	if !bytes.Equal(resMessageId, messageId) {
		log.Printf("MessageId from server isn't correct: actual - %x..., expect - %x... ", resMessageId[:4], messageId[:4])
		return 0xff, nil
	}

	resPayloadUndecode := &pb.KVResponse{}
	if err := proto.Unmarshal(msg.Payload, resPayloadUndecode); err != nil {
		return 0xff, err
	}

	errCode := resPayloadUndecode.GetErrCode()
	fmt.Println(resPayloadUndecode.GetValue())
	return errCode, nil
}


func gcpClient1(name string, ip string, port int) {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		panic("Error resolving server address")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln("UDP connection failed: ", err)
	}
	defer conn.Close()
	log.Println("Starting Client...")

	inputs := []input{
		{
			code:    0x01,
			key:     "ff00",
			val:     []byte{22, 23, 34, 45},
			version: 1,
		},
		input{
			code:    0x01,
			key:     "ff01",
			val:     nil,
			version: 0,
		},
		input{
			code:    0x01,
			key:     "ff02",
			val:     []byte{22, 22},
			version: 0,
		},
		input{
			code:    0x01,
			key:     "ff03",
			val:     nil,
			version: 0,
		},
		input{
			code:    0x01,
			key:     "ff04",
			val:     []byte{4},
			version: 0,
		},
		//input{
		//	code:    0x05,
		//	key:     "ff01",
		//	val:     nil,
		//	version: 0,
		//},
		//input{
		//	code:    0x01,
		//	key:     "ff02",
		//	val:     []byte{22,33},
		//	version: 0,
		//},
	}

	for i := 0; i < len(inputs); i++ {
		uuid := uuid.New()
		messageId, err := uuid.MarshalBinary()

		log.Printf("%s #%d requests \n", name, i+1)
		errCode, err := SendRequest(conn, messageId, inputs[i].code, inputs[i].key, inputs[i].val, inputs[i].version)
		fmt.Println(name, "Respounse from server: ", errCode)
		if err != nil {
			fmt.Println(name, "Failed: ", err)
		}
	}
}

func gcpClient2(name string, ip string, port int) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		panic("Error resolving server address")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln("UDP connection failed: ", err)
	}
	defer conn.Close()
	log.Println("Starting Client...")

	inputs := []input{
		{
			code:    0x04,
			key:     "ff00",
			val:     []byte{24},
			version: 1,
		},
		{
			code:    0x02,
			key:     "ff00",
			val:     nil,
			version: 0,
		},
		//input{
		//	code:    0x03,
		//	key:     "ff00",
		//	val:     nil,
		//	version: 0,
		//},
		//input{
		//	code:    0x01,
		//	key:     "ff01",
		//	val:     nil,
		//	version: 0,
		//},
		//input{
		//	code:    0x01,
		//	key:     "ff02",
		//	val:     []byte{22,33},
		//	version: 0,
		//},
	}

	for i := 0; i < len(inputs); i++ {
		uuid := uuid.New()
		messageId, err := uuid.MarshalBinary()

		log.Printf("%s #%d requests \n", name, i+1)

		errCode, err := SendRequest(conn, messageId, inputs[i].code, inputs[i].key, inputs[i].val, inputs[i].version)
		fmt.Println(name, "Respounse from server: ", errCode)
		if err != nil {
			fmt.Println(name, "Failed: ", err)
		}
	}
}

func gcpClient3(name string, ip string, port int) {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		panic("Error resolving server address")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln("UDP connection failed: ", err)
	}
	defer conn.Close()
	log.Println("Starting Client...")


	inputs := []input{
		{
			code:    0x02,
			key:     "ff00",
			val:     []byte{24},
			version: 1,
		},
		//input{
		//	code:    0x02,
		//	key:     "ff00",
		//	val:     nil,
		//	version: 0,
		//},
	}

	for i := 0; i < len(inputs); i++ {
		uuid := uuid.New()
		messageId, err := uuid.MarshalBinary()

		log.Printf("%s #%d requests \n", name, i+1)
		errCode, err := SendRequest(conn, messageId, inputs[i].code, inputs[i].key, inputs[i].val, inputs[i].version)
		fmt.Println(name, "Respounse from server: ", errCode)
		if err != nil {
			fmt.Println(name, "Failed: ", err)
		}
	}
}

func main() {
	fmt.Println("Start..")

	gcpClient1("CClientA", "10.138.0.3", 3332)

	time.Sleep(time.Second)
	//client1("CClientA", 3334)

	gcpClient2("B", "10.138.0.4",3333)
	time.Sleep(10 * time.Second)
	gcpClient2("B", "10.138.0.5",3334)
	//time.Sleep(10 * time.Second)
	//client2("B", 3335)
	//time.Sleep(10 * time.Second)
	//client2("B", 3331)
	//time.Sleep(10 * time.Second)
	//client2("B", 3334)
	//time.Sleep(10 * time.Second)

	gcpClient3("C", "10.138.0.2", 3331)
	time.Sleep(time.Second)

}
