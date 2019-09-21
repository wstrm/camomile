package network_test

import (
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/packet"
	"log"
	"net"
	"testing"
)

func TestStore(t *testing.T) {
	payload := &packet.Store{
		Key:                  []byte{111},
		Value:                "ABC, du är mina tankar",
	}

	r := &packet.Packet{
		PacketId: []byte{123},
		SenderId: []byte{100},
		Payload: &packet.Packet_Store{payload},
	}

	udpAddress, err := net.ResolveUDPAddr("udp", ":8118")

	network.Send(udpAddress, r)

	conn, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalf("Unable to listen at %v, %v", udpAddress.String(), err)
		return
	}
	defer conn.Close()
}
