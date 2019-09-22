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

const UdpPort  = ":8118"

type Key node.ID

func Hash(value string) Key {
	var key = blake2b.Sum256([]byte(value))
	return key
}

func ListenUDP(c chan string){
	udpAddr, err := net.ResolveUDPAddr("udp", UdpPort)
	if err != nil {
		log.Fatalln("Unable to resolve UDP address", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln("Unable to listen at UDP address", err)
	}
	defer conn.Close()

	for {
		log.Printf("Listening for UDP packets on port %v", UdpPort)

		data := make([]byte, 1500)
		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Fatalf("Error when reading from UDP from address %v: %s", addr.String(), err)
		}
		b := data[:n]
		go handlePacket(b, c)
	}
}

func Store(addr *net.UDPAddr, senderID []byte, value string) {
	id := generateID()
	key := Hash(value)

	payload := &packet.Store{
		Key:                  key[:],
		Value:                value,
	}
	p := &packet.Packet{
		PacketId:             id,
		SenderId:             senderID,
		Payload:              &packet.Packet_Store{Store: payload},
	}

	send(addr, *p)
}

func handlePacket(b []byte, c chan string) {
	p := &packet.Packet{}
	err := proto.Unmarshal(b, p)
	if err != nil {
		log.Fatalln("Error unserializing packet", err)
	}
	c <- p.GetStore().Value
}

func generateID() []byte {
	id := make([]byte, node.IDBytesLength)
	_, err := rand.Read(id)
	if err != nil {
		log.Fatalln("Error generating packetID", err)
	}
	return id
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
		log.Fatalln("Error writing via UDP", err)
	}
}

