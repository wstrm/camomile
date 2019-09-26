package network

import (
	"bytes"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"net"
	"testing"
)

const value  = "ABC, du är mina tankar."
const wrongValue = "CBA, du är i mina tankar"

var addr *net.UDPAddr
var n Network

func init() {
	addr, _ = net.ResolveUDPAddr("udp", UdpPort)
	nodeID := node.NewID()

	n, _,_,_ = NewUDPNetwork(nodeID)
}

func nextFakeID(a []byte) randRead {
	return func(b []byte) (int, error) {
		copy(a, b)
		return len(a), nil
	}
}

func TestFindValue_value(t *testing.T) {
	rng = nextFakeID([]byte{})

	// Send a findvalue request to a node att addr
	ch, err := n.FindValue(Key{}, *addr)

	// Responde to a finvalue request with a value
	err = n.SendValue(Key{}, value, []route.Contact{}, SessionID{}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <- ch
	res := r.Value

	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

func TestFindValue_contacts(t *testing.T) {
	rng = nextFakeID([]byte{})

	// Send a findvalue request to a node att addr
	ch, err := n.FindValue(Key{}, *addr)

	// Responde to a finvalue request with a list of contacts
	err = n.SendValue(Key{}, value, []route.Contact{}, SessionID{}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <- ch
	res := r.Value

	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

func TestPingPongShow_correctChallengeReply(t *testing.T) {
	rng = nextFakeID([]byte{254})

	correctChallenge := []byte{254}

	res, err := n.Ping(*addr)
	if err != nil {
		t.Error(err)
	}
	err = n.Pong(correctChallenge, SessionID{}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <- res
	rc := r.Challenge

	comp := bytes.Compare(rc, correctChallenge)

	if comp != 0 {
		t.Errorf("Got: %v Expected: %v", rc, correctChallenge)
	}
}

func TestPingPongShow_wrongChallengeReply(t *testing.T) {
	rng = nextFakeID([]byte{254})

	correctChallenge := []byte{254}
	wrongChallenge := []byte{0}

	res, err := n.Ping(*addr)
	if err != nil {
		t.Error(err)
	}
	err = n.Pong(wrongChallenge, SessionID{}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <- res
	rc := r.Challenge

	comp := bytes.Compare(rc, correctChallenge)

	if comp == 0 {
		t.Errorf("Got: %v Expected: %v", rc, correctChallenge)
	}
}
