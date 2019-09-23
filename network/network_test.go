package network

import (
	"github.com/optmzr/d7024e-dht/node"
	"net"
	"testing"
)

const value  = "ABC, du är mina tankar."

func TestFindValue(t *testing.T) {
	udpAddr, err := net.ResolveUDPAddr("udp", UdpPort)
	if err != nil {
		panic(err)
	}

	n := NewUDPNetwork(node.NewID())

	rng = func(_ []byte) (int, error) {
		return 0, nil
	}

	ch, err := n.FindValue(Key{}, udpAddr)

	err = n.SendValue(Key{}, value, PacketID{}, udpAddr)
	if err != nil {
		t.Error(err)
	}

	r := <- ch
	res := r.Value

	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

