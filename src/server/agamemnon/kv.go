package agamemnon

import (
	"bytes"
	"fmt"
	"sync"
)

// Constants defining the maximum allowable length in
// bytes of the key and value in the KVStore
const maxKeyLengthBytes = 32
const maxValLengthBytes = 10000

// Data type used to represent a K-V pair
type StoreVal struct {
	Key     []byte
	Value   []byte
	Version int32
}

// In-memory key-value store data structure
type KV struct {
	sync.RWMutex
	KVStore []StoreVal
}

// Get the value and version for a particular key
//
// Arguments:
//		key: key to get the value and version for
// Returns:
//		Byte array containing the value if the key exists
//		Version of the entry if the key exists
//		NO_ERR if key exists, otherwise KEY_DNE_ERR
func (kv *KV) Get(key []byte) ([]byte, int32, uint32) {
	kv.RLock()
	defer kv.RUnlock()
	for _, value := range kv.KVStore {
		if bytes.Equal(key, value.Key) {
			return value.Value, value.Version, NO_ERR
		}
	}

	return nil, 0, KEY_DNE_ERR
}

// GetAll returns a *copy* of all values in kv
// Copy can be accessed without locks later on
func (kv *KV) GetAll() []StoreVal {
	kv.RLock()
	defer kv.RUnlock()
	kvCopy := make([]StoreVal, len(kv.KVStore))
	copy(kvCopy, kv.KVStore)
	return kvCopy
}

// Put a new key-value pair or update an existing one
//
// Arguments:
// 		key: key for the pair
//		value: value for the pair
//		version: pointer to the version of the pair
// Returns:
//		Error code
func (kv *KV) Put(key []byte, value []byte, version *int32) uint32 {
	if len(key) > maxKeyLengthBytes {
		return INVALID_KEY_ERR
	} else if len(value) > maxValLengthBytes {
		return INVALID_VAL_ERR
	} else if !IsAllocatePossible(len(key) + len(value) + 4) {
		return NO_SPC_ERR
	} else {
		storeVal := StoreVal{Key: key, Value: value}
		if version == nil {
			storeVal.Version = 0
		} else {
			storeVal.Version = *version
		}

		kv.Lock()
		defer kv.Unlock()
		for i, value := range kv.KVStore {
			if bytes.Equal(key, value.Key) {
				kv.KVStore[i] = storeVal
				return NO_ERR
			}
		}

		kv.KVStore = append(kv.KVStore, storeVal)
		return NO_ERR
	}
}

// AddFrom adds all values from otherKV to kv
// *Does not checks for duplicates!*
func (kv *KV) AddFrom(otherKV *KV) {
	if kv == otherKV {
		return
	}
	kv.Lock()
	otherKV.Lock()
	defer kv.Unlock()
	defer otherKV.Unlock()
	kv.KVStore = append(kv.KVStore, otherKV.KVStore...)
}

// Remove a key-value pair
//
// Arguments:
// 		key: key to remove
// Returns:
//		NO_ERR if key exists, otherwise KEY_DNE_ERR
func (kv *KV) Remove(key []byte) uint32 {
	kv.Lock()
	defer kv.Unlock()
	for i, value := range kv.KVStore {
		if bytes.Equal(key, value.Key) {
			kv.KVStore[i] = kv.KVStore[len(kv.KVStore)-1]
			kv.KVStore = kv.KVStore[:len(kv.KVStore)-1]
			return NO_ERR
		}
	}

	return KEY_DNE_ERR
}

// RemoveSlice removes a slice of keys which hashcodes are <= endHashCode
func (kv *KV) RemoveSlice(endHashCode uint32) {
	kv.Lock()
	defer kv.Unlock()
	result := make([]StoreVal, 0)
	for _, value := range kv.KVStore {
		if hash(value.Key) > endHashCode {
			result = append(result, value)
		}
	}
	kv.KVStore = result
}

// Removes all the key-value pairs in the system
//
// Returns:
//		NO_ERR
func (kv *KV) RemoveAll() uint32 {
	kv.Lock()
	kv.KVStore = []StoreVal{}
	kv.Unlock()

	return NO_ERR
}
//func (kv *KV) printDataStoreNoLock() {
//	log.Debug(fmt.Sprintf("KV Store (No Lock): %v", kv))
//}
//
//func (kv *KV) printDataStore() {
//	kv.RLock()
//	defer kv.RUnlock()
//	log.Debug(fmt.Sprintf("KV Store (Lock): %v", kv))
//}
//

func (kv *KV) printDataStoreNoLock() {
	fmt.Println("KV Store(No lock):")
	for _, v := range kv.KVStore {
		fmt.Println( "#", v)
	}
}

func (kv *KV) printDataStore() {
	kv.RLock()
	defer kv.RUnlock()
	fmt.Println("KV Store:")
	for _, v := range kv.KVStore {
		fmt.Println( "#", v)
	}
}

func (kv *KV) printDataStoreMsg(msg string) {
	kv.RLock()
	defer kv.RUnlock()
	fmt.Print(msg, "-KV Store:")
	for _, v := range kv.KVStore {
		fmt.Print( "#", v, "; ")
	}
	fmt.Println()
}
