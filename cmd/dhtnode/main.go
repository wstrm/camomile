package main

import "fmt"
import "net"

func handlePacket(conn *net.UDPConn) {
    // TODO: implement what happens with received messages
    // go-routines go here and connect to channel for message proccessing?
}

// function takes a msg string, and an UDP address struct as defined in 
// net/UDPAddr docs.
// the message string will have to be serialized with protobuf beforehand.
func sendPacket(msg string, address *net.UDPAddr){
    Conn, _ := net.DialUDP("udp", nil, address)
    defer Conn.Close()
    Conn.Write([]byte(msg))
}

func main() {
	fmt.Println("dhtnode")

    // listen to all addresses on port 8118
    udpAddress, err := net.ResolveUDPAddr("udp", ":8118")
    conn, err := net.ListenUDP("udp", udpAddress)
    if err != nil {
        fmt.Println("could not listen for incoming UDP packets, is this port already in use?")
        fmt.Println(err)
        return
    }

    defer conn.Close()

    // say hello to yourelf, this should probably be changed to a node in the network.
    sendPacket("hello, I'm a node", &net.UDPAddr{IP:[]byte{127,0,0,1},Port:8118,Zone:""})

    for {
        fmt.Println("listening for UDP packets on port 8118")
        data := make([]byte, 512)

        n, addr, err := conn.ReadFromUDP(data)
        if err != nil { panic(err)  }
        s := string(data[:n])

        // print the request data
        fmt.Println("UDP client sent a message from: ", addr)
        fmt.Println(s)
        go handlePacket(conn)
    }
}
