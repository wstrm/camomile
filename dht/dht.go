package dht

import (
	"encoding/hex"
	"fmt"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
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

func (dht *DHT) iterativeFindNodes(target node.ID) (contacts []route.Contact, err error) {
	network := dht.network
	rt := dht.rt

	// Holds the currently closest contact found.
	var closest route.Contact

	// Holds a slice of nodes that did not answer.
	var dead []route.Contact

	// Holds a slice of channels that are awaiting a response from the network.
	var await []chan int // TODO(optmzr): Change to some message type.

	// The first α contacts selected are used to create a *shortlist* for the
	// search.
	shortlist := dht.rt.NClosest(target, α)

	// Keep a map of acked contacts to make sure we do not contact the same
	// node multiple times.
	acked := make(map[node.ID]bool)

	// If a cycle results in an unchanged `closest` node, then a FindNode
	// network call should be made to each of the closest nodes that has not
	// already been queried.
	rest := false

	for {
		for i, contact := range shortlist {
			if i >= α || !rest {
				break // Limit to α contacts per shortlist.
			}
			if acked[contact.NodeID] {
				continue // Ignore already acked contacts.
			}

			ch, err := network.FindNode(contact.NodeID)
			if err != nil {
				dead = append(dead, contact)
			} else {
				await = append(await, ch)
			}
		}

		results := make(chan int) // TODO(optmzr): Change to some result type.
		for ch := range await {
			go func(ch chan int) { // TODO(optmzr): Change to some message type.
				r := <-ch

				// Network call timed out.
				if r == nil {
					dead = append(dead, contact)
					return
				}

				// Add node so it is moved to the top of its bucket in the
				// routing table.
				rt.Add(r.From) // TODO(optmzr): Result type must contain a From field.

				results <- r
			}(ch)
		}

		if len(await) > 0 {
			// TODO(optmzr)
		}
	}
}

func (dht *DHT) iterativeStore(value string) (hash Key, err error) {
}

func (dht *DHT) iterativeFindValue(hash Key) (value string, err error) {
}
