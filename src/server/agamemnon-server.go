package main

import (
	lib "agamemnon/src/server/agamemnon"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

func usage() {
	fmt.Printf("Usage: go run src/server/agamemnon-server.go PORT SERVERLIST_FILE \n\n")
}

// TODO: change back
//func findCurrentIP() []net.IP {
//	host, _ := os.Hostname()
//	addrs, _ := net.LookupIP(host)
//
//	var IPs []net.IP
//	for _, addr := range addrs {
//		if ipv4 := addr.To4(); ipv4 != nil {
//			IPs = append(IPs, ipv4)
//		}
//	}
//	return IPs
//}
func getCurrentIPs() []string {
	ifaces, _ := net.Interfaces()
	ips := []string{}
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.To4().String())
			}
		}
	}
	return ips
}

func readNodeInfo(info string) (string, int, error) {
	addr := strings.Split(info, ":")
	if len(addr) != 2 {
		return "", 0, fmt.Errorf("Invalid IP:PORT - %s.", info)
	}

	port, err := strconv.Atoi(addr[1])
	if err != nil {
		return "", 0, fmt.Errorf("Invalid :PORT - %s.", info)
	}

	//TODO: check for valid IP here

	return addr[0], port, nil
}

// $go run server.go PORT servers.txt
func main() {
	if len(os.Args) < 3 {
		log.Println("Error: number of arguments is incorrect")
		usage()
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usage()
		fmt.Println("Error: ", err)
	}
	serverList, err := os.Open(os.Args[2])
	if err != nil {
		log.Println("Error: invalid file name")
		usage()
		return
	}
	defer serverList.Close()

	var nodes []*lib.Node
	thisNodeIPs := getCurrentIPs()
	scanner := bufio.NewScanner(serverList)
	for scanner.Scan() {
		nodeIp, nodePort, err := readNodeInfo(scanner.Text())
		if err != nil {
			usage()
			panic(err)
		}
		isSelf := false
		if nodePort == port {
			for _, ip := range thisNodeIPs {
				if ip == nodeIp {
					isSelf = true
				}
			}
		}

		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", nodeIp, nodePort))
		if err != nil {
			panic("Error resolving server address")
		}

		newNode := lib.Node{
			Ip:       nodeIp,
			Port:     nodePort,
			HashCode: lib.HashCodeForIpPort(fmt.Sprintf("%v:%v", nodeIp, nodePort)),
			Addr:     addr,
			IsSelf:   isSelf,
		}
		nodes = append(nodes, &newNode)

		if isSelf {
			lib.SetSelf(&newNode)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].HashCode < nodes[j].HashCode
	})

	fmt.Println("\n\n\n\n\n")
	lib.StartServer(port, nodes)
}
