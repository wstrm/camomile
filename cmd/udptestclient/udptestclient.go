package main

import "net"

// You can use this testclient to send a "hello" message to an UDP server.
func main() {
	conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{127, 0, 0, 1}, Port: 8118, Zone: ""})
	defer conn.Close()
	conn.Write([]byte("hello"))
}
