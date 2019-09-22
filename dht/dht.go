package dht

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

const α = 3 // Degree of parallelism.

type Key node.ID // TODO(optmzr): use store.Key instead.

type DHT struct {
	rt      *route.Table
	network *Network
}

// TODO(optmzr): Move to network.
type Network struct{}

// TODO(optmzr): Move to network.
func (net *Network) FindNodes(target node.ID) (chan *NodeListResult, error) {
	return make(chan *NodeListResult), nil
}

// TODO(optmzr): Move to network.
type NodeListResult struct {
	From    route.Contact
	Closest []route.Contact
}

// TODO(optmzr): Move to store.
func KeyFromString(str string) (key Key, err error) {
	h, err := hex.DecodeString(str)
	if err != nil {
		err = fmt.Errorf("cannot decode hex string as key: %w", err)
		return
	}

	copy(key[:], h)
	return
}

func NewDHT(me route.Contact, others []route.Contact) (dht *DHT, err error) {
	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	return
}

// Get retrieves the value for a specified key from the network.
func (dht *DHT) Get(hash Key) (value string, err error) {
	value, err = dht.iterativeFindValue(hash)
	return
}

// Put stores the provided value in the network and returns a key.
func (dht *DHT) Put(value string) (hash Key, err error) {
	hash, err = dht.iterativeStore(value)
	return
}

// Join initiates a node lookup of itself to bootstrap the node into the
// network.
func (dht *DHT) Join(me route.Contact) (err error) {
	_, err = dht.iterativeFindNodes(me.NodeID)
	return
}

func (dht *DHT) iterativeFindNodes(target node.ID) ([]route.Contact, error) {
	network := dht.network
	rt := dht.rt

	// Holds a slice of channels that are awaiting a response from the network.
	var await []chan *NodeListResult

	// The first α contacts selected are used to create a *shortlist* for the
	// search.
	sl := dht.rt.NClosest(target, α)

	// Keep a map of acked contacts to make sure we do not contact the same
	// node multiple times.
	acked := make(map[node.ID]bool)

	// If a cycle results in an unchanged `closest` node, then a FindNode
	// network call should be made to each of the closest nodes that has not
	// already been queried.
	rest := false

	// Contacts holds a sorted (slice) copy of the shortlist.
	contacts := sl.SortedContacts()

	// Closest is the node that closest in distance to the target node ID.
	closest := contacts[0]

	for {
		for i, contact := range contacts {
			if i >= α || !rest {
				break // Limit to α contacts per shortlist.
			}
			if acked[contact.NodeID] {
				continue // Ignore already acked contacts.
			}

			ch, err := network.FindNodes(contact.NodeID)
			if err != nil {
				sl.Remove(contact)
			} else {
				await = append(await, ch)
			}
		}

		if len(await) == 0 {
			return nil, errors.New("couldn't connect to any node")
		}

		results := make(chan *NodeListResult)
		for _, ch := range await {
			go func(ch chan *NodeListResult) {
				defer close(ch)

				// Redirect all responses to the results channel.
				r := <-ch
				results <- r
			}(ch)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for {
			result := <-results
			if result != nil {
				// Add node so it is moved to the top of its bucket in the
				// routing table.
				rt.Add(result.From)

				// Add the responding node's closest contacts.
				sl.Add(result.Closest...)
			} else {
				// Network call timed out. Remove the callee from the shortlist.
				sl.Remove(result.From)
			}
		}

		contacts = sl.SortedContacts()
		first := contacts[0]
		if closest.NodeID.Equal(first.NodeID) {
			// Unchanged closest node from last run, re-run but check all the
			// nodes in the shortlist (and not only the α closest).
			if !rest {
				rest = true
				continue
			}

			// Done. Return the contacts in the shortlist sorted by distance.
			return contacts, nil

		} else {
			// New closest node found, continue iteration.
			closest = first
		}
	}
}

func (dht *DHT) iterativeStore(value string) (hash Key, err error) {
	return Key{}, errors.New("Not implemented")
}

func (dht *DHT) iterativeFindValue(hash Key) (value string, err error) {
	return "", errors.New("Not implemented")
}
