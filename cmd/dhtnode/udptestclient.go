package main

import "net"

// you can use this testclient to send a "hello" message to a UDP server. 
func main() {
  Conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{IP:[]byte{127,0,0,1},Port:8118,Zone:""})
  defer Conn.Close()
  Conn.Write([]byte("hello"))
}
