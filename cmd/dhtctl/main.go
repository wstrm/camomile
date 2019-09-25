package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"

	"github.com/optmzr/d7024e-dht/cmd"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/store"
)

func put(c *rpc.Client, val string) {
	put := cmd.Put{
		Val: val,
	}
	var reply bool

	// The RPC call
	err := c.Call("API.Put", put, &reply)
	if err != nil {
		log.Fatal("Put error:", err)
	}

	// Debug (reply)
	if reply {
		fmt.Println("RPC successful")
	}
}

func get(c *rpc.Client, key store.Key) {
	get := cmd.Get{
		Key: key,
	}
	var reply bool

	// The RPC call
	err := c.Call("API.Get", get, &reply)
	if err != nil {
		log.Fatal("Get error:", err)
	}

	// Debug (reply)
	if reply {
		fmt.Println("RPC successful")
	}
}

func ping(c *rpc.Client, id node.ID) {
	ping := cmd.Ping{NodeID: id}
	var reply bool

	// The RPC call
	err := c.Call("API.Ping", ping, &reply)
	if err != nil {
		log.Fatal("Ping error:", err)
	}

	// Debug (reply)
	if reply {
		fmt.Println("RPC successful")
	}
}

func exit(c *rpc.Client, id node.ID) {
	exit := cmd.Exit{NodeID: id}
	var reply bool

	// The RPC call
	err := c.Call("API.Exit", exit, &reply)
	if err != nil {
		log.Fatal("Ping error:", err)
	}

	// Debug (reply)
	if reply {
		fmt.Println("RPC successful")
	}
}

func main() {
	// Flags
	var addressFlag = flag.String("address", "localhost:1234", "the address of the node")
	var putFlag = flag.String("put", "", "put value to store")
	var getFlag = flag.String("get", "", "key of the value to get")
	var pingFlag = flag.String("ping", "", "ID of the node to ping")
	var exitFlag = flag.String("exit", "", "ID of the node to terminate")

	// Parse input
	flag.Parse()

	// Dial RPC Server
	client, err := rpc.DialHTTP("tcp", *addressFlag)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Execute tasks
	if "" != *putFlag {
		put(client, *putFlag)
	}
	if "" != *getFlag {
		key, err := store.KeyFromString(*getFlag)
		if err != nil {
			log.Fatalln(err)
		}
		get(client, key)
	}
	if *pingFlag != "" {
		id, err := node.IDFromString(*pingFlag)
		if err != nil {
			log.Fatalln(err)
		}
		ping(client, id)
	}
	if *exitFlag != "" {
		id, err := node.IDFromString(*exitFlag)
		if err != nil {
			log.Fatalln(err)
		}
		exit(client, id)
	}

	// Debug
	fmt.Println("Address: ", *addressFlag)
	fmt.Println("Put: ", *putFlag)
	fmt.Println("Get: ", *getFlag)
	fmt.Println("Ping:", *pingFlag)
	fmt.Println("Exit: ", *exitFlag)
}
