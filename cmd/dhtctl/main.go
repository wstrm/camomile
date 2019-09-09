package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func executePut() {
	//TODO
}

func executeGet() {
	//TODO
}

func executePing() {
	//TODO
}

func executeExit() {
	//TODO
}

func main() {
	// Flags
	var address = flag.String("address", "localhost", "the ip-address of the node")
	var put = flag.String("put", "", "put value to store")
	var get = flag.String("get", "", "key of the value to get")
	var ping = flag.Bool("ping", false, "ping the node at address")
	var exit = flag.Bool("exit", false, "terminate the node")

	// Parse input and validate ip-address from CLI
	flag.Parse()
	udpAddr := net.UDPAddr{IP: net.ParseIP(*address), Port: 8118}
	if udpAddr.IP == nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", "Unable to parse IP address.")
		os.Exit(1)
	}

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
	fmt.Println("IP: ", udpAddr)
	fmt.Println("Put: ", *put)
	fmt.Println("Get: ", *get)
	fmt.Println("Ping :", *ping)
	fmt.Println("Exit: ", *exit)
}
