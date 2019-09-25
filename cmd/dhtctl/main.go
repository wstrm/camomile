package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/optmzr/d7024e-dht/cmd"
	"log"
	"net/rpc"
)

func put(c *rpc.Client, val string) {
	put := cmd.Put{
		Id:  nil,
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

func get(c *rpc.Client, key []byte) {
	get := cmd.Get{
		Id:  nil,
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

func ping(c *rpc.Client, id []byte) {
	ping := cmd.Ping{Id: id}
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

func exit(c *rpc.Client, id []byte) {
	exit := cmd.Exit{Id: id}
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
		key, err := hex.DecodeString(*getFlag)
		if err != nil {
			log.Fatalln(err)
		}
		get(client, key)
	}
	if *pingFlag != "" {
		id, err := hex.DecodeString(*pingFlag)
		if err != nil {
			log.Fatalln(err)
		}
		ping(client, id)
	}
	if *exitFlag != "" {
		id, err := hex.DecodeString(*exitFlag)
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
