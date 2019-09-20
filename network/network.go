package network

import (
	"github.com/golang/protobuf/proto"
	"github.com/optmzr/d7024e-dht/packet"
	"log"
	"net"
)


func Send(addr *net.UDPAddr, packet *packet.Packet) /*Todo: return chan*/ {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	b, err := proto.Marshal(packet)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = conn.Write(b)
	if err != nil {
		log.Fatalln(err)
	}
}
