package log1

import (
	"encoding/hex"
	"fmt"
	"os"
)

const logsEnabled = false

var selfKey = ""

func InitLogs(thisNodeKey string) {
	selfKey = thisNodeKey
}

func Build(key string, msgId []byte, cmd interface{}, isIn bool, port int) string {
	if !logsEnabled {
		return ""
	}
	var inout string
	if isIn {
		inout = "<<<"
	} else {
		inout = ">>>"
	}
	return fmt.Sprintf("%v %v %v %v %v", key, hex.EncodeToString(msgId), cmd, inout, port)
}

func BuildMsgErr(msg string, msgId []byte, cmd interface{}, port int, err error) string {
	return fmt.Sprintf("%v %v %v %v\n%v", msg, hex.EncodeToString(msgId), cmd, port, err)
}

func BuildErr(key string, err error) string {
	return fmt.Sprintf("%v %v", key, err)
}

func Debug(msg string) {
	if logsEnabled {
		fmt.Printf("%v: %v\n", selfKey, msg)
	}
}

func Force(msg string) {
	fmt.Printf("%v: %v\n", selfKey, msg)
}

func Fatal(msg string, err error) {
	fmt.Printf("%v: %v: %v\n", selfKey, msg, err)
	os.Exit(1)
}

const debug = true

func PrintDebug(msg string) {
	if debug {
		fmt.Println(msg)
	}
}

func PrintDebugError(msg string, err error) {
	if debug {
		fmt.Println(msg, err)
	}
}

