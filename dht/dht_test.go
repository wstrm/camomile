package dht

import (
	"bytes"
	"math/rand" // Insecure on purpose due to testing.
	"net"
	"testing"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

// udpNetwork is a mock that fulfills the network.Network interface.
type udpNetwork struct{}

// findNodesResult is a mock that fulfills the network.Result interface.
type findNodesResult struct {
	from    route.Contact
	closest []route.Contact
}

func (r *findNodesResult) From() route.Contact {
	return r.from
}

func (r *findNodesResult) Closest() []route.Contact {
	return r.closest
}

// Accessed by multiple goroutines, must not be changed except by init().
var others []route.Contact
var me route.Contact

func init() {
	//log.SetFlags(0)
	//log.SetOutput(ioutil.Discard)

	id := node.NewID()
	me = route.NewContact(id, id, net.UDPAddr{
		IP:   net.IP{10, 10, 10, 254},
		Port: 123,
		Zone: "",
	})

	for i := 0; i < 100; i++ {
		others = append(others, route.NewContact(id, node.NewID(), net.UDPAddr{
			IP:   net.IP{10, 10, 10, byte(i)},
			Port: 123,
			Zone: "",
		}))
	}
}

// FindNodes mocks a FindNodes call by returning a NodeListResult with some
// random contacts as closest.
func (net *udpNetwork) FindNodes(target node.ID, address net.UDPAddr) (chan network.Result, error) {
	ch := make(chan network.Result)
	go func() {
		var id node.ID
		found := false

		if address.IP.Equal(me.Address.IP) {
			id = me.NodeID
			found = true
		} else {
			// Find this nodes ID by the address in the others slice.
			for _, contact := range others {
				if address.IP.Equal(contact.Address.IP) {
					id = contact.NodeID
					found = true
				}
			}
		}

		if !found {
			panic("address used doesn't exist in the test contacts")
		}

		// Pick some random contacts as closest.
		i := rand.Int()
		l := len(others)
		closest := []route.Contact{
			others[i%l],
			others[(i+1)%l],
			others[(i+2)%l],
		}

		// Send fake FindNodesResult.
		ch <- &findNodesResult{
			from:    route.Contact{NodeID: id, Address: address},
			closest: closest,
		}
	}()
	return ch, nil
}

func (net *udpNetwork) Ping(addr net.UDPAddr) (chan *network.PingResult, error) { return nil, nil }
func (net *udpNetwork) Pong(challenge []byte, sessionID network.SessionID, addr net.UDPAddr) error {
	return nil
}
func (net *udpNetwork) FindValue(key store.Key, addr net.UDPAddr) (chan *network.FindValueResult, error) {
	return nil, nil
}
func (net *udpNetwork) SendValue(key store.Key, value string, closets []route.Contact, sessionID network.SessionID, addr net.UDPAddr) error {
	return nil
}
func (net *udpNetwork) Store(key store.Key, value string, addr net.UDPAddr) error {
	return nil
}

func TestJoin(t *testing.T) {
	d, err := New(me, others[:1], new(udpNetwork))
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	err = d.Join(me)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}
}

func TestPut(t *testing.T) {
	d, err := New(me, others[:1], new(udpNetwork))
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	hash, err := d.Put("ABC, du är mina tankar")
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	expHash := store.Key{
		189, 224, 233, 246, 233, 211, 250, 189, 91, 246, 132, 158, 23, 159, 10,
		238, 72, 86, 48, 246, 213, 193, 196, 57, 133, 23, 204, 21, 67, 251, 147,
		134,
	}

	if !bytes.Equal(hash[:], expHash[:]) {
		t.Errorf("unexpected hash, got: %v, exp: %v", hash, expHash)
	}
}
