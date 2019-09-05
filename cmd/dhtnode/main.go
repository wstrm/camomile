package main

import "fmt"
import "net"
import "log"

func handlePacket(conn *net.UDPConn) {
	// TODO: implement what happens with received messages.
	// go-routines go here and connect to channel for message proccessing?
}

// sendPacket takes packet bytes, and an UDP address.
// The packet bytes must be serialized using Protobuf beforehand.
func sendPacket(packet []byte, address *net.UDPAddr) {
	conn, _ := net.DialUDP("udp", nil, address)
	defer conn.Close()
	conn.Write(packet)
}

func main() {
	fmt.Println("dhtnode")

	// Listen to all addresses on port 8118.
	udpAddress, err := net.ResolveUDPAddr("udp", ":8118")
	conn, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalln("Unable to listen at %s, %v", udpAddress.String(), err)
		return
	}

	defer conn.Close()

	// Say hello to yourself, this should probably be changed to a node in the network.
	sendPacket("Hello, I'm a node", &net.UDPAddr{IP: []byte{127, 0, 0, 1}, Port: 8118, Zone: ""})

	for {
		fmt.Println("Listening for UDP packets on port 8118")
		data := make([]byte, 512)

		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Fatalln(err)
		}
		s := string(data[:n])

		// Print the request data.
		fmt.Println("UDP client sent a message from: ", addr)
		fmt.Println(s)
		go handlePacket(conn)
	}
}
