package dht

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/optmzr/d7024e-dht/node"
)

const α = 3 // Degree of parallelism.

type Key node.ID // TODO(optmzr): use store.Key instead.

type DHT struct {
	rt *route.Table
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
	value, err = iterativeFindValue(hash)
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
	_, err = iterativeFindNode(me)
	return
}

func (dht *DHT) iterativeFindNodes(target node.ID) ([]route.Contact, error) {
	network := dht.network
	rt := dht.rt

	// Holds the currently closest contact found.
	var closest route.Contact

	// Holds a slice of channels that are awaiting a response from the network.
	var await []chan int // TODO(optmzr): Change to some message type.

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
	contacts := shortlist.SortedContacts()

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

			ch, err := network.FindNode(contact.NodeID)
			if err != nil {
				sl.Remove(contact)
			} else {
				await = append(await, ch)
			}
		}

		if len(await) == 0 {
			return nil, errors.New("couldn't connect to any node")
		}

		results := make(chan int) // TODO(optmzr): Change to some result type.
		for ch := range await {
			go func(ch chan int) { // TODO(optmzr): Change to some message type.
				r := <-ch

				// Network call timed out.
				if r == nil {
					sl.Remove(contact)
					return
				}

				// Add node so it is moved to the top of its bucket in the
				// routing table.
				rt.Add(r.From) // TODO(optmzr): Result type must contain a From field.

				results <- r
			}(ch)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for {
			result := <-results
			if result != nil {
				// Add the responding node's closest contacts.
				sl.Add(result.Closest...)
			} else {
				// Network response timed out.
				sl.Remove(contact)
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
}

func (dht *DHT) iterativeFindValue(hash Key) (value string, err error) {
}
