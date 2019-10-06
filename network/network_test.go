package network

import (
	"bytes"
	stdlog "log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

const value = "ABC, du är mina tankar."

var nAddr *net.UDPAddr
var mAddr *net.UDPAddr

var nNode route.Contact
var mNode route.Contact

var n Network
var m Network

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	console := zerolog.ConsoleWriter{Out: os.Stderr,
		TimeFormat: time.Stamp,
		NoColor:    true,
	}

	logger := zerolog.New(console).With().Timestamp().Logger()
	log.Logger = logger // Set as global logger.

	// Make sure the standard logger also uses zerolog.
	stdlog.SetFlags(0)
	stdlog.SetOutput(logger)

	var err error

	nAddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:8118")
	panicOnErr(err)

	mAddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:8119")
	panicOnErr(err)

	nNode = route.Contact{
		NodeID:  node.NewID(),
		Address: *nAddr,
	}
	mNode = route.Contact{
		NodeID:  node.NewID(),
		Address: *mAddr,
	}

	n, err = NewUDPNetwork(nNode)
	panicOnErr(err)

	m, err = NewUDPNetwork(mNode)
	panicOnErr(err)

	go func(n Network) {
		err := n.Listen()
		panicOnErr(err)
	}(n)

	go func(m Network) {
		err := m.Listen()
		panicOnErr(err)
	}(m)

	<-n.ReadyCh()
	<-m.ReadyCh()
}

func nextFakeID(a []byte) randRead {
	return func(b []byte) (int, error) {
		copy(b, a)
		return len(a), nil
	}
}

func TestFindValue_value(t *testing.T) {
	rng = nextFakeID([]byte{1})

	// Send a FindValue request to a node at mNode
	ch, err := n.FindValue(store.Key{}, *mAddr)
	if err != nil {
		t.Error(err)
	}

	contacts := []route.Contact{
		route.NewContact(node.NewID(), net.UDPAddr{}),
		route.NewContact(node.NewID(), net.UDPAddr{}),
		route.NewContact(node.NewID(), net.UDPAddr{}),
		route.NewContact(node.NewID(), net.UDPAddr{}),
		route.NewContact(node.NewID(), net.UDPAddr{}),
	}

	// Respond to a FindValue request with a value.
	err = m.SendValue(store.Key{}, value, contacts, SessionID{1}, *nAddr)
	if err != nil {
		t.Error(err)
	}

	r := <-ch
	if r == nil {
		t.Errorf("unexpected nil channel")
	}

	if len(r.Closest()) != 5 {
		t.Errorf("unexpected number of contacts in closest, got: %v, exp: 5", r.Closest())
	}

	res := r.Value()
	if res != value {
		t.Errorf("Expected: %s Got: %s", value, res)
	}
}

func TestFindValue_contacts(t *testing.T) {
	rng = nextFakeID([]byte{2})

	// Send a FindValue request to a node at mNode
	ch, err := n.FindValue(store.Key{}, *mAddr)
	if err != nil {
		t.Error(err)
	}

	// Respond to a FindValue request with a list of contacts
	err = n.SendValue(store.Key{}, value, []route.Contact{}, SessionID{2}, *nAddr)
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

	res, _, err := n.Ping(*mAddr)
	if err != nil {
		t.Error(err)
	}

	err = m.Pong(correctChallenge, SessionID{3}, *nAddr)
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

	res, _, err := n.Ping(*mAddr)
	if err != nil {
		t.Error(err)
	}

	err = m.Pong(wrongChallenge, SessionID{4}, *nAddr)
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

func TestFindNodes_closest(t *testing.T) {
	rng = nextFakeID([]byte{5})

	ch, err := n.FindNodes(node.ID{}, *mAddr)
	if err != nil {
		t.Error(err)
	}

	contacts := []route.Contact{
		route.Contact{
			NodeID:  node.NewID(),
			Address: net.UDPAddr{},
		},
	}

	err = m.SendNodes(contacts, SessionID{5}, *nAddr)
	if err != nil {
		t.Error(err)
	}

	r := <-ch

	if r.Value() != "" {
		t.Errorf("unexpected value in result, got: %v, exp: \"\" (none)", r.Value())
	}

	if len(r.Closest()) != len(contacts) {
		t.Errorf("unexpected length of .Closest(): got: %v, exp: %v", r.Closest(), contacts)
	}

	if r.Closest()[0].NodeID.String() != contacts[0].NodeID.String() {
		t.Errorf("unexpected node ID in .Closest(): got: %v, exp: %v", r.Closest()[0].NodeID, contacts[0].NodeID.String())
	}
}

func TestStore(t *testing.T) {
	rng = nextFakeID([]byte{6})
	value := "ABC, du är mina tankar"
	key := store.Key{1}

	err := n.Store(key, value, *mAddr)
	if err != nil {
		t.Error(err)
	}

	r := <-m.StoreRequestCh()

	if r.Value != value {
		t.Errorf("unexpected value in request, got: %s, exp: %s", r.Value, value)
	}

	if !r.From.NodeID.Equal(nNode.NodeID) {
		t.Errorf("unexpected from node ID in request, got: %v, exp: %v", r.From.NodeID, nNode.NodeID)
	}
}
