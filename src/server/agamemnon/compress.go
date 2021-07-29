package agamemnon

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
)

func EncodeToBytes(kv interface{}) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(kv)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EncodeToBytes uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func Compress(s []byte) []byte {

	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	zipped.Write(s)
	zipped.Close()
	fmt.Println("compressed size (bytes): ", len(zipbuf.Bytes()))
	return zipbuf.Bytes()
}

func Decompress(s []byte) []byte {

	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatal(err)
	}
	rdr.Close()
	fmt.Println("uncompressed size (bytes): ", len(data))
	return data
}

func DecodeToKV(s []byte) []StoreVal {

	kvStore := []StoreVal{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&kvStore)
	if err != nil {
		log.Fatal(err)
	}
	return kvStore
}
