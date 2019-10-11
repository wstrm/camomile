package store

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/blake2b"

	"github.com/optmzr/d7024e-dht/node"
)

// Key should be a checksum made with blake2b256 hash algorithm, in binary and at a length of 32 bytes.
type Key node.ID

// item is an item stored by the kademlia network on this node.
// This contains timers that decide the retention of the object along with the stored value and identifier of the node that made the store request to the network initially.
type item struct {
	Value     string
	expire    time.Time
	republish time.Time
	origPub   node.ID
}

// localItem contains a timer and the value that this node has stored on the kademlia network.
type localItem struct {
	Value     string
	repubTime time.Time
}

// remoteItems holds multiple items, and a Mutex lock for the datastructure.
type remoteItems struct {
	sync.RWMutex
	m map[Key]item
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
	ch          chan string // only needs to return string value
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

	db.remoteItems = remoteItems{m: make(map[Key]item)}
	db.localItems = localItems{m: make(map[Key]localItem)}
	db.ch = make(chan string)

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

// LocalItemCh returns the database communication channel.
func (db *Database) LocalItemCh() chan string {
	return db.ch
}

// truncate truncates supplied string to a maximum of a 1000 characters. Returns a string.
func truncate(s string) string {
	if len(s) > 1000 {
		return s[:1000]
	}
	return s
}

// AddItem adds an value to the remoteItems database that a node in the kademlia network has sent to this node.
func (db *Database) AddItem(value string, origPub node.ID) {
	t := time.Now()
	expire := t.Add(db.tExpire)
	republish := t.Add(db.tRepublish)

	key := Key(blake2b.Sum256([]byte(truncate(value))))

	newItem := item{}

	newItem.Value = truncate(value)
	newItem.expire = expire
	newItem.republish = republish
	newItem.origPub = origPub

	db.remoteItems.Lock()
	db.remoteItems.m[key] = newItem
	db.remoteItems.Unlock()

	log.Info().Msgf("Stored:\n\tValue: %s\n\tHash: %v", value, key)
}

// AddLocalItem adds an value to the local item database that this node has requested to be stored on the kademlia network.
func (db *Database) AddLocalItem(key Key, value string) {
	t := time.Now()

	newLocalItem := localItem{}

	newLocalItem.Value = value
	newLocalItem.repubTime = t.Add(db.tRepublish)

	db.localItems.Lock()
	db.localItems.m[key] = newLocalItem
	db.localItems.Unlock()
}

// GetItem returns an item stored on this node that originated from the kademlia network.
// Also updates the expiration time of the item.
func (db *Database) GetItem(key Key) (reqItem item, err error) {
	newExpirationTime := time.Now().Add(db.tExpire)

	db.remoteItems.Lock()
	requestedItem, found := db.remoteItems.m[key]
	requestedItem.expire = newExpirationTime
	db.remoteItems.m[key] = requestedItem
	db.remoteItems.Unlock()

	if !found {
		err = fmt.Errorf("store: GetItem, no item matching key: %v", key)
		return
	}

	return requestedItem, nil
}

// GetLocalItem return a stored local item to the republishing function.
// Throws an error if no match is found in the local item DB.
func (db *Database) GetLocalItem(key Key) (reqLocalItem localItem, err error) {
	db.localItems.RLock()
	localItem, found := db.localItems.m[key]
	db.localItems.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetLocalItem, no localItem matching key: %v", key)
		return
	}

	return localItem, nil
}

// evictItem evicts an item that other nodes has stored on this node.
// The internal map delete mechanism is encapsulated within mutex and should therefore be thread safe.
func (db *Database) evictItem(key Key) {
	db.remoteItems.Lock()
	delete(db.remoteItems.m, key)
	db.remoteItems.Unlock()
}

// ForgetItem removes an item from the local items to stop it from beeing
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
			if now.After(item.expire) || now.After(item.republish) {
				evictees = append(evictees, key)
			}
		}
		db.remoteItems.RUnlock()

		for _, key := range evictees {
			db.evictItem(key)
		}
	}
}

// republishHandler checks stored localItems that's due for renewal at remote nodes.
// This function should be run as a goroutine.
func (db *Database) republishHandler(ticker *time.Ticker) {
	for now := range ticker.C {
		replicate := now.After(db.getReplicate())

		db.localItems.RLock()
		for _, localItem := range db.localItems.m {
			if now.After(localItem.repubTime) {
				db.ch <- localItem.Value
			}
		}
		db.localItems.RUnlock()

		// Replication event, replicate all stored values to k nodes.
		if replicate {
			db.remoteItems.RLock()
			for _, remoteItem := range db.remoteItems.m {
				db.ch <- remoteItem.Value
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
