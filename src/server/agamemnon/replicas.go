package agamemnon

import (
	pb "agamemnon/pb/protobuf"
	"fmt"
	"time"
)

type DataStorage struct {
	Replicas []KV
}

var dataStorage DataStorage

const replicaCopy = 2
const replicaLogEnabled = true

func InitReplicas() {
	for i := 0; i <= replicaCopy; i++ {
		copy := KV{}
		copy.KVStore = []StoreVal{}

		dataStorage.Replicas = append(dataStorage.Replicas, copy)
	}
}

func (store *DataStorage) addReplica(kv []StoreVal, repId int) {
	for _, item := range kv {
		store.Replicas[repId].Put(item.Key, item.Value, &item.Version)
	}
}

func (store DataStorage) compressReplica (repId int) []byte {
	return Compress(EncodeToBytes(store.Replicas[repId].KVStore))
}

func (store DataStorage) decompressReplica (data []byte) []StoreVal {
	return DecodeToKV(Decompress(data))
}
/**
	Failure recovery when Node3 is failed, current Node is Node4
	1 - A
	2 - B(+C)   A
	~3- C       B        A~ -> failed
  =>4 - D       C(+B)    B->A
	5   E       D        C(+B)
 */

func RecoverDataStorage() {
	// TODO: cache to responseCache?
	// 1. send replicaOne to prev node primary copy
	kvC := dataStorage.compressReplica(1)
	reqPay := pb.KVRequest{Command: ADD_REPLICA, Value: kvC}
	requestToReplicaNode(self.prevNode(), reqPay, 0)

	// 2. merge replicaTwo to replicaOne
	storageB := dataStorage.Replicas[2].KVStore
	dataStorage.addReplica(storageB, 1)

	// 3. send replicaTwo to next node replicaTwo
	kvB := dataStorage.compressReplica(2)
	reqPay3 := pb.KVRequest{Command: ADD_REPLICA, Value: kvB}
	requestToReplicaNode(self.nextNode(), reqPay3, 2)

	// 4. request from prev node replicaOne, replace replicaTwo
	reqPay4 := pb.KVRequest{Command: SEND_REPLICA}
	msgId := requestToReplicaNode(self.prevNode(), reqPay4, 1)
	dataStorage.Replicas[2].RemoveAll()

	respValue := waitingForResponseData(msgId, getNetAddress(self.prevNode().Addr), 2 * time.Second)
	if respValue != nil {
		storageValue := dataStorage.decompressReplica(respValue)
		dataStorage.addReplica(storageValue, 2)
	} else {
		fmt.Println("👩‍🎤👩‍🎤👩‍🎤👩‍🎤👩‍🎤👩‍🎤👩‍🎤👩‍🎤 failed to add A")
	}
}

func printReplicas(msg string)  {
	if replicaLogEnabled {
		fmt.Println("\n\n\n\n\n start ")
		for i, kv := range dataStorage.Replicas {
			kv.printDataStoreMsg(msg + "- # " + string(i))
		}
	}
}
