package main

import (
	"flag"
	"fmt"
)

var ipText = flag.String("address", "localhost", "the ip-address of the node")
var put = flag.String("put", "", "put value to store")
var get = flag.String("get", "", "key of the value to get")
var ping = flag.Bool("ping", false, "ping the node at address")
var exit = flag.Bool("exit", false, "terminate the node")

func main() {
	flag.Parse()
	fmt.Println("IP: ", *ipText)
	fmt.Println("Put: ", *put)
	fmt.Println("Get: ", *get)
	fmt.Println("Ping :", *ping)
	fmt.Println("Exit: ", *exit)
}
