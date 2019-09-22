package network_test

import (
	"github.com/optmzr/d7024e-dht/network"
	"log"
	"net"
	"testing"
)

const value  = "ABC, du är mina tankar."

func TestStore(t *testing.T) {
	udpAddress, err := net.ResolveUDPAddr("udp", network.UdpPort)
	if err != nil {
		log.Fatalln("Unable to resolve UDP address", err)
	}

	ch := make(chan string)
	go network.ListenUDP(ch)

	network.Store(udpAddress, []byte{100}, value)
	s := <- ch
	log.Println(s)
	if s != value {
		t.Errorf("Expected: %s, got: %s", value, s)
	}
}