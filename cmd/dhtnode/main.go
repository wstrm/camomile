package main

import (
	"fmt"
	"github.com/optmzr/d7024e-dht/cmd"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func handlePacket(conn *net.UDPConn) {
	// TODO: implement what happens with received messages.
	// go-routines go here and connect to channel for message proccessing?
}

func rpcServer() {
	api := new(cmd.API)
	err := rpc.Register(api)
	if err != nil {
		log.Fatalln(err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	fmt.Println("dhtnode")
	go rpcServer()

	// Listen to all addresses on port 8118.
	udpAddress, err := net.ResolveUDPAddr("udp", ":8118")
	if err != nil {
		log.Fatalf("Unable to resolve IP and Port, %v", err)
	}
	conn, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalf("Unable to listen at %v, %v", udpAddress.String(), err)
		return
	}

	defer conn.Close()

	for {
		fmt.Println("Listening for UDP packets on port 8118")
		data := make([]byte, 512)

		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Fatalln(err)
		}
		s := string(data[:n])

		// Print the request data.
		fmt.Println("Received a message from: ", addr)
		fmt.Println(s)
		go handlePacket(conn)
	}
}
