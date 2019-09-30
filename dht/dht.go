package dht

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
	"golang.org/x/crypto/blake2b"
)

const α = 3  // Degree of parallelism.
const k = 20 // Bucket size.

const tExpire = 86410 * time.Second    // Time after which a key/value pair expires (TTL).
const tReplicate = 3600 * time.Second  // Interval between Kademlia replication events.
const tRepublish = 86400 * time.Second // Time after which the original publisher must republish a key/value pair.

type DHT struct {
	rt *route.Table
	nw network.Network
	me route.Contact
	db *store.Database
}

func New(me route.Contact, others []route.Contact, nw network.Network) (dht *DHT, err error) {
	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	dht.db = store.NewDatabase(tExpire, tReplicate, tRepublish)

	dht.nw = nw
	dht.me = me

	go func(dht *DHT, me route.Contact) {
		<-dht.nw.ReadyCh()
		err := dht.Join(me)
		if err != nil {
			log.Fatalln(err)
		}
	}(dht, me)

	go dht.findNodesRequestHandler()
	go dht.findValueRequestHandler()
	go dht.storeRequestHandler()
	go dht.pongRequestHandler()

	return
}

func (dht *DHT) findValueRequestHandler() {
	for {
		request := <-dht.nw.FindValueRequestCh()

		log.Printf("Find value request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing table.
		dht.rt.Add(request.From)

		var closest []route.Contact
		target := node.ID(request.Key)

		// Try to fetch the value from the local storage.
		item, err := dht.db.GetItem(request.Key)
		if err != nil {
			// No luck.
			// Fetch this nodes contacts that are closest to the requested key.
			closest = dht.rt.NClosest(target, k).SortedContacts()
		} else {
			log.Printf("Found value: %s", item.Value)
		}

		err = dht.nw.SendValue(request.Key, item.Value, closest,
			request.SessionID, request.From.Address)
		if err != nil {
			log.Println(err)
		}
	}
}

func (dht *DHT) findNodesRequestHandler() {
	for {
		request := <-dht.nw.FindNodesRequestCh()

		log.Printf("Find node request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing table.
		dht.rt.Add(request.From)

		// Fetch this nodes contacts that are closest to the requested target.
		closest := dht.rt.NClosest(request.Target, k).SortedContacts()

		err := dht.nw.SendNodes(closest, request.SessionID, request.From.Address)
		if err != nil {
			log.Println(err)
		}
	}
}

func (dht *DHT) storeRequestHandler() {
	for {
		request := <-dht.nw.StoreRequestCh()

		log.Printf("Store value request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing table.
		dht.rt.Add(request.From)

		dht.db.AddItem(request.Value, request.From.NodeID)
	}
}

// Get retrieves the value for a specified key from the network.
func (dht *DHT) Get(hash store.Key) (value string, err error) {
	value, err = dht.iterativeFindValue(hash)
	return
}

// Put stores the provided value in the network and returns a key.
func (dht *DHT) Put(value string) (hash store.Key, err error) {
	hash, err = dht.iterativeStore(value)
	return
}

// Join initiates a node lookup of itself to bootstrap the node into the
// network.
func (dht *DHT) Join(me route.Contact) (err error) {
	contacts, err := dht.iterativeFindNodes(me.NodeID)
	if err != nil {
		return
	}

	logAcquaintedWith(contacts...)
	return
}

// Ping pings a specified camomileID
func (dht *DHT) Ping(target node.ID) (chal []byte, err error) {
	sl := dht.rt.NClosest(target, 1)

	contacts := sl.SortedContacts()
	if !(contacts.Len() > 0 && contacts[0].NodeID.Equal(target)) {
		return nil, fmt.Errorf("ping: could not find target node (%v)", target)
	}

	contact := contacts[0]

	resultCh, challenge, err := dht.nw.Ping(contact.Address)
	if err != nil {
		return nil, fmt.Errorf("not able to send ping request to %v: %w", contact.NodeID, err)
	}

	response := <-resultCh

	dht.rt.Add(response.From)

	if bytes.Equal(challenge, response.Challenge) {
		return response.Challenge, nil
	}
	return nil, fmt.Errorf("challenge mismatch")
}

func (dht *DHT) pongRequestHandler() {
	for {
		request := <-dht.nw.PongRequestCh()

		log.Printf("Pong request from: %v (%x)", request.From.NodeID, request.Challenge)

		dht.rt.Add(request.From)

		err := dht.nw.Pong(request.Challenge, request.SessionID, request.From.Address)
		if err != nil {
			log.Println(err)
		}
	}
}

type awaitChannel struct {
	ch     chan network.Result
	callee route.Contact
}

type awaitResult struct {
	result network.Result
	callee route.Contact
}

func (dht *DHT) walk(call Call) ([]route.Contact, error) {
	nw := dht.nw
	rt := dht.rt
	me := dht.me
	target := call.Target()

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
		await := []awaitChannel{}

		for i, contact := range contacts {
			if i >= α && !rest {
				break // Limit to α contacts per shortlist.
			}
			if sent[contact.NodeID] || contact.NodeID.Equal(me.NodeID) {
				continue // Ignore already contacted contacts or local node.
			}

			ch, err := call.Do(nw, contact.Address)
			if err != nil {
				log.Println(err)
				sl.Remove(contact)
			} else {
				// Mark as contacted.
				sent[contact.NodeID] = true

				// Add to await channel queue.
				await = append(await, awaitChannel{ch: ch, callee: contact})
			}
		}

		results := make(chan awaitResult)
		for _, ac := range await {
			go func(ac awaitChannel) {
				// Redirect all responses to the results channel.
				r := <-ac.ch
				results <- awaitResult{result: r, callee: ac.callee}
			}(ac)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for i := 0; i < len(await); i++ {
			ac := <-results
			result := ac.result
			callee := ac.callee

			if result != nil {
				// Add node so it is moved to the top of its bucket in the
				// routing table.
				rt.Add(callee)

				// Add the responding node's closest contacts.
				sl.Add(result.Closest()...)

				// Update callee with intermediate results.
				stop := call.Result(result)
				if stop {
					break // Callee requested that the walk must be stopped.
				}
			} else {
				// Network call timed out. Remove the callee from the shortlist.
				sl.Remove(callee)
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

func (dht *DHT) iterativeFindNodes(target node.ID) ([]route.Contact, error) {
	return dht.walk(NewFindNodesCall(target))
}

func (dht *DHT) iterativeStore(value string) (hash store.Key, err error) {
	hash = blake2b.Sum256([]byte(value))

	contacts, err := dht.iterativeFindNodes(node.ID(hash))
	if err != nil {
		return
	}

	var stored []route.Contact
	for _, contact := range contacts {
		if e := dht.nw.Store(hash, value, contact.Address); e != nil {
			log.Printf("Failed to store at %s (%s): %v",
				contact.NodeID.String(), contact.Address.String(), e)
		} else {
			stored = append(stored, contact)
		}
	}

	if len(stored) > 0 {
		logStoredAt(hash, stored...)
	}

	return
}

func (dht *DHT) iterativeFindValue(hash store.Key) (value string, err error) {
	call := NewFindValueCall(hash)
	closest, err := dht.walk(call)

	if err != nil {
		return
	}

	if call.value != "" {
		value = call.value
	} else {
		err = fmt.Errorf("Couldn't find any value with the hash: %v", hash)
		return
	}

	// Store at the closest node that did not return any value.
	if len(closest) > 0 {
		first := closest[0]
		if e := dht.nw.Store(hash, value, first.Address); e != nil {
			log.Printf("Failed to store at %s (%s): %v",
				first.NodeID.String(), first.Address.String(), e)
		} else {
			logStoredAt(hash, first)
		}
	}

	return
}

func logStoredAt(hash store.Key, contacts ...route.Contact) {
	log.Printf("Stored value with hash %v at %d nodes:\n",
		hash.String(), len(contacts))
	logContacts(contacts...)
}

func logAcquaintedWith(contacts ...route.Contact) {
	log.Printf("Acquainted with %d nodes:\n", len(contacts))
	logContacts(contacts...)
}

func logContacts(contacts ...route.Contact) {
	for _, contact := range contacts {
		log.Println("\t", contact.NodeID.String())
	}
}
