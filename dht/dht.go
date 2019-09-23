package dht

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"golang.org/x/crypto/blake2b"
)

const α = 3 // Degree of parallelism.

type Key node.ID // TODO(optmzr): use store.Key instead.

type DHT struct {
	rt      *route.Table
	network Network
}

// TODO(optmzr): Move to network.
type Network interface {
	FindNodes(target node.ID, address net.UDPAddr) (chan *NodeListResult, error)
	Store(value string, address net.UDPAddr) error
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

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

func New(me route.Contact, others []route.Contact, network Network) (dht *DHT, err error) {
	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	dht.network = network

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
	contacts, err := dht.iterativeFindNodes(me.NodeID)
	logAcquaintedWith(contacts)
	return
}

func (dht *DHT) iterativeFindNodes(target node.ID) ([]route.Contact, error) {
	network := dht.network
	rt := dht.rt

	// The first α contacts selected are used to create a *shortlist* for the
	// search.
	sl := dht.rt.NClosest(target, α)

	// Keep a map of contacts that has been sent to, to make sure we do not
	// contact the same node multiple times.
	sent := make(map[node.ID]bool)

	// If a cycle results in an unchanged `closest` node, then a FindNode
	// network call should be made to each of the closest nodes that has not
	// already been queried.
	rest := false

	// Contacts holds a sorted (slice) copy of the shortlist.
	contacts := sl.SortedContacts()

	// Closest is the node that closest in distance to the target node ID.
	closest := contacts[0]

	for {
		// Holds a slice of channels that are awaiting a response from the
		// network.
		await := [](chan *NodeListResult){}

		for i, contact := range contacts {
			if i >= α && !rest {
				break // Limit to α contacts per shortlist.
			}
			if sent[contact.NodeID] {
				continue // Ignore already contacted contacts.
			}

			ch, err := network.FindNodes(target, contact.Address)
			if err != nil {
				sl.Remove(contact)
			} else {
				// Mark as contacted.
				sent[contact.NodeID] = true

				// Add to await channel queue.
				await = append(await, ch)
			}
		}

		results := make(chan *NodeListResult)
		for _, ch := range await {
			// TODO(optmzr): Should this handle timeouts, or the network
			// package?
			go func(ch chan *NodeListResult) {
				// Redirect all responses to the results channel.
				r := <-ch
				results <- r
			}(ch)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for i := 0; i < len(await); i++ {
			// TODO(optmzr): Should this handle timeouts, or the network
			// package?
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
	hash = blake2b.Sum256([]byte(value))

	contacts, err := dht.iterativeFindNodes(node.ID(hash))
	if err != nil {
		return
	}

	for _, contact := range contacts {
		err = dht.network.Store(value, contact.Address)
		if err != nil {
			return // TODO(optmzr): Collect errors?
		}
	}

	logStoredAt(hash, contacts)

	return
}

func (dht *DHT) iterativeFindValue(hash Key) (value string, err error) {
	return "", errors.New("Not implemented")
}

func logStoredAt(hash Key, contacts []route.Contact) {
	log.Printf("Stored value with hash %v at %d nodes:\n",
		hash.String(), len(contacts))
	logContacts(contacts)
}

func logAcquaintedWith(contacts []route.Contact) {
	log.Printf("Acquainted with %d nodes:\n", len(contacts))
	logContacts(contacts)
}

func logContacts(contacts []route.Contact) {
	for _, contact := range contacts {
		log.Println("\t", contact.NodeID.String())
	}
}
