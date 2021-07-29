package test

import (
	"agamemnon/src/server/agamemnon"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

func init_data()  {


}

func test_compress(t *testing.T)  {
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

	bytes := agamemnon.Compress(agamemnon.EncodeToBytes(kv))
	ds := agamemnon.DecodeToKV(agamemnon.Decompress(bytes))

	fmt.Println(ds[0])
	reflect.DeepEqual(kv, ds)
}

func main()  {

}