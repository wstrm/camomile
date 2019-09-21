package network

import (
	"crypto/rand"
	"github.com/golang/protobuf/proto"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/packet"
	"golang.org/x/crypto/blake2b"
	"log"
	"net"
)

type key node.ID

func Hash(value string) key {
	var key = blake2b.Sum256([]byte(value))
	return key
}

func Store(addr *net.UDPAddr, senderID []byte, value string) {
	// Generate packetID
	id := make([]byte, 4)
	_, err := rand.Read(id)
	if err != nil {
		log.Fatalln(err)
	}

	key := Hash(value)

	payload := &packet.Store{
		Key:                  key[:],
		Value:                value,
	}
	packet := &packet.Packet{
		PacketId:             id,
		SenderId:             senderID,
		Payload:              &packet.Packet_Store{Store: payload},
	}

	send(addr, *packet)
}

func send(addr *net.UDPAddr, packet packet.Packet) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	b, err := proto.Marshal(&packet)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = conn.Write(b)
	if err != nil {
		log.Fatalln(err)
	}
}
