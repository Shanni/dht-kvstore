package str

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return -1, fmt.Errorf("invalid port: %v", portStr)
	}
	return port, nil
}

func ParseIpPort(ipPortStr string) (string, int, error) {
	ipPortSlice := strings.Split(ipPortStr, ":")
	if len(ipPortSlice) != 2 {
		return "", -1, fmt.Errorf("wrong ip port str: %v", ipPortStr)
	}
	port, err := strconv.Atoi(ipPortSlice[1])
	if err != nil {
		return "", -1, err
	}
	return ipPortSlice[0], port, nil
}
