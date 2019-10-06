package network

import (
	"bytes"
	"log"
	"net"
	"testing"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

const value = "ABC, du är mina tankar."

var addr *net.UDPAddr
var n Network

func init() {
	addr, _ = net.ResolveUDPAddr("udp", ":8118")
	me := route.Contact{
		NodeID:  node.NewID(),
		Address: *addr,
	}

	var err error
	n, err = NewUDPNetwork(me)
	if err != nil {
		log.Fatalln(err)
	}

	go func(n Network) {
		err := n.Listen()
		if err != nil {
			log.Fatalln(err)
		}
	}(n)

	<-n.ReadyCh()
}

func nextFakeID(a []byte) randRead {
	return func(b []byte) (int, error) {
		copy(b, a)
		return len(a), nil
	}
}

func TestFindValue_value(t *testing.T) {
	rng = nextFakeID([]byte{1})

	// Send a findvalue request to a node att addr
	ch, err := n.FindValue(store.Key{}, *addr)
	if err != nil {
		t.Error(err)
	}

	// Respond to a findvalue request with a value
	err = n.SendValue(store.Key{}, value, []route.Contact{}, SessionID{1}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <-ch
	if r == nil {
		t.Errorf("unexpected nil channel")
	}
	res := r.Value()

	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

func TestFindValue_contacts(t *testing.T) {
	rng = nextFakeID([]byte{2})

	// Send a findvalue request to a node att addr
	ch, err := n.FindValue(store.Key{}, *addr)
	if err != nil {
		t.Error(err)
	}

	// Respond to a findvalue request with a list of contacts
	err = n.SendValue(store.Key{}, value, []route.Contact{}, SessionID{2}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <-ch
	if r == nil {
		t.Errorf("unexpected nil channel")
	}
	res := r.Value()

	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

func TestPingPongShow_correctChallengeReply(t *testing.T) {
	rng = nextFakeID([]byte{3})

	correctChallenge := []byte{254}

	res, _, err := n.Ping(*addr)
	if err != nil {
		t.Error(err)
	}
	err = n.Pong(correctChallenge, SessionID{3}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <-res
	rc := r.Challenge

	comp := bytes.Compare(rc, correctChallenge)

	if comp != 0 {
		t.Errorf("Got: %v Expected: %v", rc, correctChallenge)
	}
}

func TestPingPongShow_wrongChallengeReply(t *testing.T) {
	rng = nextFakeID([]byte{4})

	correctChallenge := []byte{254}
	wrongChallenge := []byte{0}

	res, _, err := n.Ping(*addr)
	if err != nil {
		t.Error(err)
	}
	err = n.Pong(wrongChallenge, SessionID{4}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <-res
	rc := r.Challenge

	comp := bytes.Compare(rc, correctChallenge)

	if comp == 0 {
		t.Errorf("Got: %v Expected: %v", rc, correctChallenge)
	}
}

func TestFindNodes_value(t *testing.T) {
	rng = nextFakeID([]byte{5})

	ch, err := n.FindNodes(node.ID{}, *addr)
	if err != nil {
		t.Error(err)
	}

	contacts := []route.Contact{
		route.Contact{
			NodeID:  node.NewID(),
			Address: net.UDPAddr{},
		},
	}
	err = n.SendNodes(contacts, SessionID{5}, *addr)
	if err != nil {
		t.Error(err)
	}

	r := <-ch

	if len(r.Closest()) != len(contacts) {
		t.Errorf("unexpected length of .Closest(): got: %v, exp: %v", r.Closest(), contacts)
	}

	if r.Closest()[0].NodeID.String() != contacts[0].NodeID.String() {
		t.Errorf("unexpected node ID in .Closest(): got: %v, exp: %v", r.Closest()[0].NodeID, contacts[0].NodeID.String())
	}
}
