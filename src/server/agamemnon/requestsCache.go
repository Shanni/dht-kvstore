package agamemnon

import (
	"encoding/hex"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
)

// store 1) incoming message and
//		2) message forwarding

// clientAddr: request source
// ACK: store request "ack", if the request is forwarded to
// 		other node, e.g. PUT sending replicas
type RequestCacheVal struct {
	id         []byte
	clientAddr *net.UDPAddr
	ackedAddr  []*net.UDPAddr
	requiredAck bool
	timestamp  time.Time

	ttl        int8
}

type RequestCache struct {
	sync.RWMutex
	cache map[string]*RequestCacheVal
}

var incomingCache RequestCache

func (r *RequestCache) Init() {
	r.cache = map[string]*RequestCacheVal{}
}

func (r RequestCache) Get(msgID []byte) *RequestCacheVal {
	r.RLock()
	defer r.RUnlock()

	key := hex.EncodeToString(msgID)
	if val, ok := r.cache[key]; ok {
		return val
	}
	return nil
}

//// TODO: remove - refractor
//func CacheRequest(msgID []byte, clientAddr *net.UDPAddr) {
//	requestCacheMutex.Lock()
//	defer requestCacheMutex.Unlock()
//
//	key := hex.EncodeToString(msgID)
//
//	cacheVal := RequestCacheVal{id: msgID, clientAddr: clientAddr, ackedAddr:  []*net.UDPAddr{}}
//	RequestCache[key] = &cacheVal
//}

func (r *RequestCache) Add(msgID []byte, clientAddr *net.UDPAddr) {
	r.Lock()
	defer r.Unlock()

	key := hex.EncodeToString(msgID)

	cacheVal := RequestCacheVal{id: msgID, clientAddr: clientAddr, ackedAddr:  []*net.UDPAddr{}, ttl: 4}
	r.cache[key] = &cacheVal
}

func (r *RequestCache) Delete(msgID []byte) (*net.UDPAddr, bool) {
	r.Lock()
	defer r.Unlock()
	key := hex.EncodeToString(msgID)
	val, ok := r.cache[key]
	if ok {
		delete(r.cache, key)
		return val.clientAddr, true
	} else {
		// TODO: change it to log
		fmt.Println("DeleteRequestCache: not found", key)
	}
	return nil, false
}

//// set total number of ack needed
//func setRequiredAsk(msdId []byte, num int) {
//	if reqCacheValue := GetCacheRequest(msdId); reqCacheValue != nil {
//		requestCacheMutex.Lock()
//		defer requestCacheMutex.Unlock()
//		reqCacheValue.requiredAck = num
//	}
//}

//func requestAcked(msdId []byte, addr *net.UDPAddr) {
//	if reqCacheValue := GetCacheRequest(msdId); reqCacheValue != nil {
//		for _, address := range reqCacheValue.ackedAddr {
//			if address == addr {
//			 	return
//			}
//		}
//
//		requestCacheMutex.Lock()
//		defer requestCacheMutex.Unlock()
//		reqCacheValue.ackedAddr = append(reqCacheValue.ackedAddr, addr)
//	}
//}

func (r *RequestCache) TTLManager() {
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

		time.Sleep(10 * time.Second)
		//r.printData()
	}
}

func requestCacheToString(msg string)  {
	fmt.Println("debugggg", len(incomingCache.cache))
	for k, v := range incomingCache.cache {
		fmt.Println(msg, "-resquestCache-", k, v)
	}
}
