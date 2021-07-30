package agamemnon

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// store 1) outgoing message before sending out and
// 		2) message and data response

type CacheVal struct {
	response []byte
	ttl      int8
}

type ResponseCache struct {
	sync.RWMutex
	cache map[string]*CacheVal
}

var responseCache ResponseCache

func (r *ResponseCache) Init() {
	r.cache = map[string]*CacheVal{}
}

func (r *ResponseCache) Add(msgID []byte, respMsgBytes []byte) bool {
	if IsAllocatePossible(len(msgID) + len(respMsgBytes) + 1) {
		r.Lock()
		defer r.Unlock()
		fmt.Println("ADDING after sending msg", self.Port, msgID, respMsgBytes)
		key := hex.EncodeToString(msgID)
		cacheVal := CacheVal{response: respMsgBytes, ttl: 4}
		r.cache[key] = &cacheVal
		return true
	}
	return false
}

func (r ResponseCache) Get(msgID []byte) []byte {
	r.RLock()
	defer r.RUnlock()

	key := hex.EncodeToString(msgID)
	if val, ok := r.cache[key]; ok {
		fmt.Println("👽👽👽👽 GET ", self.Port, msgID, val.response)
		return val.response
	}
	return nil
}


func (r *ResponseCache) Delete(msgID []byte) ([]byte, bool) {
	r.Lock()
	defer r.Unlock()
	key := hex.EncodeToString(msgID)
	val, ok := r.cache[key]
	if ok {
		delete(r.cache, key)
		return val.response, true
	} else {
		// TODO: change it to log
		fmt.Println("Delete incomingCache: not found", key)
	}
	return nil, false
}

func (r *ResponseCache) flush() {
	//r.cache = make(map[string]*CacheVal)
	for k, _ := range r.cache {
		delete(r.cache, k)
	}
}

func (r *ResponseCache) TTLManager() {
	r.Init()

	for {
		for k, v := range r.cache {
			if v.ttl > 0 {
				v.ttl -= 1
			} else {
				r.Lock()
				delete(r.cache, k)
				r.Unlock()
			}
		}

		// Run the GC explicitly since many items may have been removed
		runtime.GC()

		// Sleep for 1 second
		time.Sleep(time.Second)
		//r.printData()
	}
}

func (r *ResponseCache) printData() {
	fmt.Println("\n\n\n\n\n\n", self.Port)
	for k, v := range r.cache {
		msg, _ := hex.DecodeString(k)
		fmt.Println(msg, v)
	}
}
