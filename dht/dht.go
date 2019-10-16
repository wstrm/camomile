package dht

import (
	"bytes"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

const α = 3                // Degree of parallelism.
const k = route.BucketSize // Bucket size.

const tExpire = 86410 * time.Second    // Time after which a key/value pair expires (TTL).
const tReplicate = 3600 * time.Second  // Interval between replication events.
const tRepublish = 86400 * time.Second // Time after which the original publisher must republish a key/value pair.
const tRefresh = 3600 * time.Second    // Time after which the routing table requests a refresh of an untouched bucket.

type DHT struct {
	rt *route.Table
	nw network.Network
	me route.Contact
	db *store.Database
}

func New(me route.Contact, others []route.Contact, nw network.Network) (dht *DHT, err error) {
	refreshTicker := time.NewTicker(60 * time.Second)

	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others, tRefresh, refreshTicker)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)

	dht.db = store.NewDatabase(tExpire, tReplicate, tRepublish, iHTicker, rHTicker)

	dht.nw = nw
	dht.me = me

	go func(dht *DHT, me route.Contact) {
		<-dht.nw.ReadyCh() // Wait for network.

		retryInterval := 1 * time.Second
		for {
			err := dht.Join(me)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to join the DHT network, retrying in %v", retryInterval)
			} else {
				break // Join successful, exit retry loop.
			}

			time.Sleep(retryInterval)
		}
	}(dht, me)

	go dht.findNodesRequestHandler()
	go dht.findValueRequestHandler()
	go dht.storeRequestHandler()
	go dht.pongRequestHandler()
	go dht.republishRequestHandler()
	go dht.replicateRequestHandler()
	go dht.refreshRequestHandler()

	return
}

// Forget removes the key and associated value from the local items DB and
// therefore stop republishing it on the network.
func (dht *DHT) Forget(hash store.Key) {
	dht.db.ForgetItem(hash)
}

// Get retrieves the value for a specified key from the network.
func (dht *DHT) Get(hash store.Key) (value string, sender node.ID, err error) {
	value, sender, err = dht.iterativeFindValue(hash)
	return
}

// Put stores the provided value in the network and returns a key.
func (dht *DHT) Put(value string) (hash store.Key, err error) {
	hash, err = dht.iterativeStore(value, network.StoreClassPublish)
	if err != nil {
		return
	}
	dht.db.AddLocalItem(hash, value)
	return
}

// Join initiates a node lookup of itself to bootstrap the node into the
// network.
func (dht *DHT) Join(me route.Contact) (err error) {
	_, err = dht.iterativeFindNodes(me.NodeID)
	if err != nil {
		return
	}

	for id := range node.IDWithPrefixGenerator(me.NodeID) {
		_, err = dht.iterativeFindNodes(id)
		if err != nil {
			return
		}
	}

	return
}

// Ping pings a specified node ID.
func (dht *DHT) Ping(target node.ID) (chal []byte, err error) {
	sl := dht.rt.NClosest(target, 1)

	contacts := sl.SortedContacts()
	if !(contacts.Len() > 0 && contacts[0].NodeID.Equal(target)) {
		return nil, fmt.Errorf("ping: could not find target node (%v)", target)
	}

	contact := contacts[0]

	resultCh, challenge, err := dht.nw.Ping(contact.Address)
	if err != nil {
		return nil, fmt.Errorf("ping request failed for: %v: %w",
			contact.NodeID, err)
	}

	response := <-resultCh
	if response == nil {
		return nil, fmt.Errorf("ping response from: %v timed out", contact.NodeID)
	}

	go dht.addNode(contact)

	if bytes.Equal(challenge, response.Challenge) {
		return response.Challenge, nil
	}
	return nil, fmt.Errorf("challenge mismatch")
}

// addNode attempts to add a node to the routing table. If the bucket is full
// for the given node, the least recently seen node will be pinged and evicted
// if it doesn't respond. If the bucket already contain the node, it'll be moved
// to the top of the bucket.
func (dht *DHT) addNode(contact route.Contact) {
	rt := dht.rt

	ok := rt.Add(contact)
	if ok {
		return
	}

	old := rt.Head(contact.NodeID).NodeID
	// Check if the oldest node is still alive.
	// If the node answers, it'll be moved to the top of the bucket by the Ping
	// method.
	_, err := dht.Ping(old)

	if err != nil {
		// Either challenge mismatch or dead node, remove it.
		rt.Remove(old)

		// Re-try to add new node.
		ok = rt.Add(contact)
		if !ok {
			log.Warn().Msg("Unable to add new node even after old node was evicted")
		}
		return
	}
}

func (dht *DHT) iterativeFindNodes(target node.ID) ([]route.Contact, error) {
	return dht.walk(NewFindNodesCall(target))
}

func (dht *DHT) iterativeStore(value string, class network.StoreClass) (hash store.Key, err error) {
	hash = store.KeyFromValue(value)

	contacts, err := dht.iterativeFindNodes(node.ID(hash))
	if err != nil {
		return
	}

	// Do not replicate the value over more than k nodes.
	if len(contacts) > k {
		contacts = contacts[:k]
	}

	var stored []route.Contact
	for _, contact := range contacts {
		if e := dht.nw.Store(hash, value, class, contact.Address); e != nil {
			logFailedStoreAt(contact, e)
		} else {
			stored = append(stored, contact)
		}
	}

	if len(stored) > 0 {
		logStoredAt(hash, stored...)
	}

	return
}

func (dht *DHT) iterativeFindValue(hash store.Key) (value string, sender node.ID, err error) {
	call := NewFindValueCall(hash)
	closest, err := dht.walk(call)

	if err != nil {
		return
	}

	if call.value != "" {
		value = call.value
		sender = call.sender
	} else {
		err = fmt.Errorf("couldn't find any value with the hash: %v", hash)
		return
	}

	// Store at the closest node that did not return any value.
	if len(closest) > 0 {
		first := closest[0]
		if e := dht.nw.Store(hash, value, network.StoreClassReplicate, first.Address); e != nil {
			logFailedStoreAt(first, e)
		} else {
			logStoredAt(hash, first)
		}
	}

	return
}

func logFailedStoreAt(contact route.Contact, err error) {
	log.Error().Err(err).Msgf("Failed to store at %v (%v)", contact.NodeID, contact.Address)
}

func logStoredAt(hash store.Key, contacts ...route.Contact) {
	log.Info().Msgf("Stored value with hash %v at %d nodes:\n%s", hash.String(), len(contacts), tabbedContactList(contacts...))
}

func tabbedContactList(contacts ...route.Contact) (cl string) {
	for _, contact := range contacts {
		cl += "\t" + contact.NodeID.String() + "\n"
	}
	return
}
