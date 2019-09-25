package store

import "time"
import "golang.org/x/crypto/blake2b"
import "sync"
import "fmt"

// Key should be a checksum made with blake2b256 hash algorithm, in binary and at a length of 32 bytes.
type Key node.ID

// item is an item stored by the kademlia network on this node.
// This contains timers that decide the retention of the object along with the stored value and identifier of the node that made the store request to the network initially.
type item struct {
	value     string
	expire    time.Time
	republish time.Time
	origPub   node.ID
}

// localItem contains a timer and the value that this node has stored on the kademlia network.
type localItem struct {
	value     string
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
	remoteItems      remoteItems
	localItems localItems
	ch         chan localItem
	replicate  replicate
	tExpire    time.Duration
	tReplicate time.Duration
	tRepublish time.Duration
}

// NewDatabase instantiates a new database object with the given time constants, returns a Database pointer and a channel.
// Spins up the two governing handlers as go routines, responsible for maintaining the database.
func NewDatabase(tExpire, tReplicate, tRepublish time.Duration) (*Database, chan localItem) {
	db := new(Database)

	db.tExpire = tExpire
	db.tReplicate = tReplicate
	db.tRepublish = tRepublish

	db.remoteItems = remoteItems{m: make(map[Key]item)}
	db.localItems = localItems{m: make(map[Key]localItem)}
	db.ch = make(chan localItem)

	go db.itemHandler()
	go db.republishHandler()

	return db, db.ch
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

// truncate truncates supplied string to a maximum of a 1000 characters. Returns a string.
func truncate(s string) string {
    if len(s) > 1000 {
        return s[:1000]
    }
    return s
}

// AddItem adds an value to the remoteItems database that a node in the kademlia network has sent to this node. 
func (db *Database) AddItem(value string, origPub node.ID) error {
	t := time.Now()
	expire := t.Add(db.tExpire)
	republish := t.Add(db.tRepublish)

	key := Key(blake2b.Sum256([]byte(truncate(value))))

	newItem := item{}

	newItem.value = truncate(value)
	newItem.expire = expire
	newItem.republish = republish
	newItem.origPub = origPub

	db.remoteItems.Lock()
	db.remoteItems.m[key] = newItem
	db.remoteItems.Unlock()

	return nil
}

// AddLocalItem adds an value to the local item database that this node has requested to be stored on the kademlia network.
func (db *Database) AddLocalItem(key Key, value string) {
	t := time.Now()

	newLocalItem := localItem{}

	newLocalItem.value = value
	newLocalItem.repubTime = t.Add(db.tRepublish)

	db.localItems.Lock()
	db.localItems.m[key] = newLocalItem
	db.localItems.Unlock()
}

// GetItem returns an item stored on this node that originated from the kademlia network.
func (db *Database) GetItem(key Key) (reqItem item, err error) {
	db.remoteItems.RLock()
	requestedItem, found := db.remoteItems.m[key]
	db.remoteItems.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetItem, no item matching key: %x", key)
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
		err = fmt.Errorf("store: GetLocalItem, no localItem matching key: %x", key)
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

// itemHandler checks for expired items every second and remove them if they're outdated.
// This function should be run as a goroutine.
func (db *Database) itemHandler() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
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
func (db *Database) republishHandler() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
		replicate := now.After(db.getReplicate())

		db.localItems.RLock()
		for _, localItem := range db.localItems.m {
			if now.After(localItem.repubTime) {
				db.ch <- localItem
			} else if replicate {
				// Replication event, replicate all stored values to k nodes.
				db.ch <- localItem
			}
		}
		db.localItems.RUnlock()

		// If a replication event just happened, reset the replication timer.
		if replicate {
			db.setReplicate()
		}
	}
}
