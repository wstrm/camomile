package main

import (
	"flag"
	"fmt"
	"net"
)

func strToIP(str string) net.IP {
	return net.ParseIP(str)
}

func executePut() {

}

func executeGet() {

}

func executePing() {

}

func executeExit() {

}

func main() {
	// Flags
	var ip = flag.String("address", "localhost", "the ip-address of the node")
	var put = flag.String("put", "", "put value to store")
	var get = flag.String("get", "", "key of the value to get")
	var ping = flag.Bool("ping", false, "ping the node at address")
	var exit = flag.Bool("exit", false, "terminate the node")

	// Parse input and validate ip-address from CLI
	flag.Parse()
	// TODO: netIP, err := strToIP(ip)

	// Execute tasks given via CLI
	if "" != *put {
		executePut()
	}
	if "" != *get {
		executeGet()
	}
	if *ping {
		executePing()
	}
	if *exit {
		executeExit()
	}

	// Debug output
	fmt.Println("IP: ", *ip)
	fmt.Println("Put: ", *put)
	fmt.Println("Get: ", *get)
	fmt.Println("Ping :", *ping)
	fmt.Println("Exit: ", *exit)
}
