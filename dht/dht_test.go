package dht

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand" // Insecure on purpose due to testing.
	"net"
	"testing"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

// udpNetwork is a mock that fulfills the network.Network interface.
type udpNetwork struct{}

// Accessed by multiple goroutines, must not be changed except by init().
var others []route.Contact
var me route.Contact

func init() {
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

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
func (net *udpNetwork) FindNodes(target node.ID, address net.UDPAddr) (chan *NodeListResult, error) {
	ch := make(chan *NodeListResult)
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

		// Send fake NodeList.
		ch <- &NodeListResult{
			From:    route.Contact{NodeID: id, Address: address},
			Closest: closest,
		}
	}()
	return ch, nil
}

// Store mocks a Store by immediately returning.
func (net *udpNetwork) Store(_ string, address net.UDPAddr) error {
	return nil
}

func TestJoin(t *testing.T) {
	d, err := New(me, others, new(udpNetwork))
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	err = d.Join(me)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}
}

func TestPut(t *testing.T) {
	d, err := New(me, others, new(udpNetwork))
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	hash, err := d.Put("ABC, du är mina tankar")
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	fmt.Println(hash)
}
