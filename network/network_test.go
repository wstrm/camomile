package network_test

import (
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/packet"
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

	network.Send(&net.UDPAddr{IP: []byte{127, 0, 0, 1}, Port: 8118, Zone: ""}, r)


}
