package main

import (
	"agamemnon/src/server/agamemnon"
	"fmt"
	"strconv"
	"sync"
)

func test_compress()  {
	kv := agamemnon.DataStorage{
		Replicas: []agamemnon.KV{
			{
				sync.RWMutex{},
				[]agamemnon.StoreVal{},
			},
		},
	}

	for i := 0; i < 15000; i++{
		kv.Replicas[0].KVStore = append(kv.Replicas[0].KVStore, agamemnon.StoreVal{
			Key: []byte(strconv.Itoa(i)),
			Value: []byte(strconv.Itoa(i) + "CGUKVBYILUINDIO:WENDINFENFJKMKSL:MKL:CIDONVIOE"),
		})
	}

	bytes := agamemnon.Compress(agamemnon.EncodeToBytes(kv.Replicas[0].KVStore))
	ds := agamemnon.DecodeToKV(agamemnon.Decompress(bytes))

	fmt.Println(string(ds[0].Value))
	//reflect.DeepEqual(kv, ds)
}

func main()  {
	test_compress()
}
