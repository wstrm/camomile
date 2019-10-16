package store

import (
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/blake2b"

	"github.com/optmzr/d7024e-dht/node"
)

// Key should be a checksum made with blake2b256 hash algorithm, in binary and at a length of 32 bytes.
type Key node.ID

type Item struct {
	Key   Key
	Value string
}

// item is an item stored by the kademlia network on this node.
// This contains timers that decide the retention of the object along with the stored value and identifier of the node that made the store request to the network initially.
type remoteItem struct {
	value  string
	expire time.Time
}

// localItem contains a timer and the value that this node has stored on the kademlia network.
type localItem struct {
	value     string
	republish time.Time
}

// remoteItems holds multiple items, and a Mutex lock for the datastructure.
type remoteItems struct {
	sync.RWMutex
	m map[Key]remoteItem
}

// localItems holds multiple local items, and a Mutex lock for the datastructure.
type localItems struct {
	sync.RWMutex
	m map[Key]localItem
}

// replicate stores the time at which to run the database replication event, protected by a Mutex lock.
type replicate struct {
	sync.RWMutex
	time time.Time
}

// Database object that contains the 2 datastructures holding remote and local items.
// Time constants dictate the behaviour of the database according to the kademlia algorithm.
// The channel enables the database to signal DHT when to send republish events.
type Database struct {
	remoteItems remoteItems
	localItems  localItems
	replicateCh chan Item
	republishCh chan Item
	replicate   replicate
	tExpire     time.Duration
	tReplicate  time.Duration
	tRepublish  time.Duration
}

// NewDatabase instantiates a new database object with the given time constants, returns a Database pointer and a channel.
// Spins up the two governing handlers as go routines, responsible for maintaining the database.
func NewDatabase(tExpire, tReplicate, tRepublish time.Duration, iHTicker, rHTicker *time.Ticker) *Database {
	db := new(Database)

	db.tExpire = tExpire
	db.tReplicate = tReplicate
	db.tRepublish = tRepublish
	db.setReplicate()

	db.remoteItems = remoteItems{m: make(map[Key]remoteItem)}
	db.localItems = localItems{m: make(map[Key]localItem)}

	db.replicateCh = make(chan Item)
	db.republishCh = make(chan Item)

	go db.itemHandler(iHTicker)
	go db.republishHandler(rHTicker)

	return db
}

// setReplicate, a set function for the replication interval time of the database.
func (db *Database) setReplicate() {
	db.replicate.Lock()
	db.replicate.time = time.Now().Add(time.Duration(db.tReplicate))
	db.replicate.Unlock()
}

// getReplicate, a get function for the replication interval time of the database. Returns a time.Time object.
func (db *Database) getReplicate() time.Time {
	db.replicate.RLock()
	time := db.replicate.time
	db.replicate.RUnlock()

	return time
}

// ItemCh returns the database communication channel.
func (db *Database) ReplicateCh() chan Item {
	return db.replicateCh
}

// ItemCh returns the database communication channel.
func (db *Database) RepublishCh() chan Item {
	return db.republishCh
}

// AddItem adds an value to the remoteItems database that a node in the Kademlia network has sent to this node.
func (db *Database) AddItem(key Key, value string, centrality int, k int, touch bool) {
	db.remoteItems.RLock()
	_, ok := db.remoteItems.m[key]
	db.remoteItems.RUnlock()

	if ok && !touch {
		return
	}

	value = truncate(value)
	t := time.Now()

	// The expiration time should be "exponentially inversely proportional to
	// the number between the current node and the node whose ID closest to the
	// key". According to the following formula:
	//	Let:
	//		Ca = Number of contacts in the bucket corresponding to the key.
	//		Cb = Number of contacts in the buckets further away from the key.
	//		C = Ca + Cb (the centrality parameter).
	//	Then, the expiration, E, is:
	//		E = {
	//			24 hours; if C > k,
	//			24*exp(k/C) hours; otherwise.
	//		}
	var expire time.Time
	if centrality > k {
		expire = t.Add(db.tExpire)
	} else {
		n := float64(db.tExpire.Seconds())
		p := math.Exp(float64(k) / float64(centrality))
		r := int64(n * p)
		d := time.Duration(r) * time.Second

		expire = t.Add(d)
	}

	item := remoteItem{
		value:  value,
		expire: expire,
	}

	db.remoteItems.Lock()
	db.remoteItems.m[key] = item
	db.remoteItems.Unlock()

	log.Info().Msgf("Stored:\n\tValue: %s\n\tHash: %v", value, key)
}

// AddLocalItem adds an value to the local item database that this node has requested to be stored on the kademlia network.
func (db *Database) AddLocalItem(key Key, value string) {
	value = truncate(value)

	t := time.Now()

	item := localItem{
		value:     value,
		republish: t.Add(db.tRepublish),
	}

	db.localItems.Lock()
	db.localItems.m[key] = item
	db.localItems.Unlock()
}

// GetItem returns an item stored on this node that originated from the kademlia network.
// Also updates the expiration time of the item.
func (db *Database) GetItem(key Key) (item Item, err error) {
	newExpirationTime := time.Now().Add(db.tExpire)

	db.remoteItems.Lock()
	defer db.remoteItems.Unlock()

	remoteItem, found := db.remoteItems.m[key]
	if !found {
		err = fmt.Errorf("no item matching key: %v", key)
		return
	}

	remoteItem.expire = newExpirationTime
	db.remoteItems.m[key] = remoteItem

	item = Item{Key: key, Value: remoteItem.value}
	return
}

// evictRemoteItem evicts an item that other nodes has stored on this node.
// The internal map delete mechanism is encapsulated within mutex and should therefore be thread safe.
func (db *Database) evictRemoteItem(key Key) {
	log.Debug().Msgf("Evicting: %v", key)
	db.remoteItems.Lock()
	delete(db.remoteItems.m, key)
	db.remoteItems.Unlock()
}

// ForgetItem removes an item from the local items to stop it from being
// republished on the Kademlia network and eventually cease to exist.
func (db *Database) ForgetItem(key Key) {
	db.localItems.Lock()
	delete(db.localItems.m, key)
	db.localItems.Unlock()
}

// itemHandler checks for expired items every second and remove them if they're outdated.
// This function should be run as a goroutine.
func (db *Database) itemHandler(ticker *time.Ticker) {
	for now := range ticker.C {
		var evictees []Key

		db.remoteItems.RLock()
		for key, item := range db.remoteItems.m {
			if now.After(item.expire) {
				evictees = append(evictees, key)
			}
		}
		db.remoteItems.RUnlock()

		for _, key := range evictees {
			db.evictRemoteItem(key)
		}
	}
}

// republishHandler checks stored localItems that's due for renewal at remote nodes.
// This function should be run as a goroutine.
func (db *Database) republishHandler(ticker *time.Ticker) {
	for now := range ticker.C {
		replicate := now.After(db.getReplicate())

		db.localItems.Lock()
		for key, localItem := range db.localItems.m {
			if now.After(localItem.republish) {

				// Update republish timestamp.
				localItem.republish = now.Add(db.tRepublish)
				db.localItems.m[key] = localItem

				db.republishCh <- Item{Key: key, Value: localItem.value}
			}
		}
		db.localItems.Unlock()

		// Replication event, replicate all stored values to k nodes.
		if replicate {
			db.remoteItems.RLock()
			for key, remoteItem := range db.remoteItems.m {
				db.replicateCh <- Item{Key: key, Value: remoteItem.value}
			}
			db.remoteItems.RUnlock()
		}

		// If a replication event just happened, reset the replication timer.
		if replicate {
			db.setReplicate()
		}
	}
}

// KeyFromString parses a hexadecimal representation of the key into a Key.
func KeyFromString(str string) (key Key, err error) {
	h, err := hex.DecodeString(str)
	if err != nil {
		err = fmt.Errorf("cannot decode hex string as key: %w", err)
		return
	}

	copy(key[:], h)
	return
}

// String returns the hexadecimal representation of a Key.
func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

func KeyFromValue(value string) Key {
	return blake2b.Sum256([]byte(truncate(value)))
}

// truncate truncates supplied string to a maximum of a 1000 characters. Returns a string.
func truncate(s string) string {
	if len(s) > 1000 {
		return s[:1000]
	}
	return s
}

func (item Item) String() string {
	return fmt.Sprintf("%v: %s", item.Key, item.Value)
}
