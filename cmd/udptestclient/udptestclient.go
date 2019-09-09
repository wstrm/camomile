package main

import "fmt"
import "net"

// You can use this testclient to send a "hello" message to an UDP server.
func main() {
	conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{127, 0, 0, 1}, Port: 8118, Zone: ""})
	defer conn.Close()
	_, err := conn.Write([]byte("hello"))
	if err != nil {
		fmt.Println("Could not send message", err)
	}
}
