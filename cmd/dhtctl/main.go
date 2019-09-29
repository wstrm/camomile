package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"

	"github.com/optmzr/d7024e-dht/ctl"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/store"
)

func put(c *rpc.Client, value string) {
	put := ctl.Put{
		Value: value,
	}
	var key store.Key

	// The RPC call
	err := c.Call("API.Put", put, &key)
	if err != nil {
		log.Fatal("Put error:", err)
	}

	log.Printf("Hash: %v\n", key)
}

func get(c *rpc.Client, key store.Key) {
	get := ctl.Get{
		Key: key,
	}
	var value string

	// The RPC call
	err := c.Call("API.Get", get, &value)
	if err != nil {
		log.Fatalln("Get error:", err)
	}

	log.Printf("Value: %s\n", value)
}

func ping(c *rpc.Client, id node.ID) {
	ping := ctl.Ping{NodeID: id}
	var challenge []byte

	// The RPC call
	err := c.Call("API.Ping", ping, &challenge)
	if err != nil {
		log.Fatalln("Ping error:", err)
	}

	fmt.Printf("Ping response: %x\n", challenge)
}

func exit(c *rpc.Client) {
	var ok bool

	// The RPC call
	err := c.Call("API.Exit", ctl.Exit{}, &ok)
	if err != nil {
		log.Fatalf("Exit error: %v\n", err)
	}

	if ok {
		fmt.Println("Node terminated")
	}
}

func main() {
	// Flags
	var addressFlag = flag.String("address", "localhost:1234", "the address of the node")
	var putFlag = flag.String("put", "", "put value to store")
	var getFlag = flag.String("get", "", "key of the value to get")
	var pingFlag = flag.String("ping", "", "ID of the node to ping")
	var exitFlag = flag.Bool("exit", false, "Terminate the node")

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
	if *exitFlag {
		exit(client)
	}
}
