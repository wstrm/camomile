package store

//import "log"
import "time"
import "golang.org/x/crypto/blake2b"
import "sync"
import "fmt"

type NodeID [32]byte
type Key NodeID

type item struct {
	value     string
	expire    time.Time
	republish time.Time
	origPub   NodeID
}

type localItem struct {
	value     string
	repubTime time.Time
}

type items struct {
	sync.RWMutex
	m map[Key]item
}

type localItems struct {
	sync.RWMutex
	m map[Key]localItem
}

type replicate struct {
	sync.RWMutex
	time time.Time
}

type Database struct {
	items      items
	localItems localItems
	ch         chan (localItem)
	replicate  replicate
	tExpire    time.Duration
	tReplicate time.Duration
	tRepublish time.Duration
}

func NewDatabase(tExpire, tReplicate, tRepublish time.Duration) *Database {
	db := new(Database)

	db.tExpire = tExpire
	db.tReplicate = tReplicate
	db.tRepublish = tRepublish

	db.items = items{m: make(map[Key]item)}
	db.localItems = localItems{m: make(map[Key]localItem)}
	db.ch = make(chan localItem)

	go db.itemHandler()
	go db.republishHandler()

	return db
}

func (db *Database) setReplicate() {
	db.replicate.Lock()
	db.replicate.time = time.Now().Add(time.Duration(db.tReplicate))
	db.replicate.Unlock()
}

func (db *Database) getReplicate() time.Time {
	db.replicate.RLock()
	time := db.replicate.time
	db.replicate.RUnlock()

	return time
}

// TODO: Max 1000 chars. Truncate input string.
func (db *Database) AddItem(value string, origPub NodeID) error {
	t := time.Now()
	expire := t.Add(db.tExpire)
	republish := t.Add(db.tRepublish)

	key := Key(blake2b.Sum256([]byte(value)))

	newItem := item{}

	newItem.value = value
	newItem.expire = expire
	newItem.republish = republish
	newItem.origPub = origPub

	// TODO: Handle duplicate keys.
	db.items.Lock()
	db.items.m[key] = newItem
	db.items.Unlock()

	return nil
}

func (db *Database) AddLocalItem(key Key, value string) {
	t := time.Now()

	newLocalItem := localItem{}

	newLocalItem.value = value
	newLocalItem.repubTime = t.Add(db.tRepublish)

	db.localItems.Lock()
	db.localItems.m[key] = newLocalItem
	db.localItems.Unlock()
}

func (db *Database) GetItem(key Key) (reqItem item, err error) {
	db.items.RLock()
	requestedItem, found := db.items.m[key]
	db.items.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetItem, no item matching key: %x", key)
		return
	}

	return requestedItem, nil
}

/*
func (db *Database) GetRepubTime(key Key) (repubTime time.Time, err error) {
	db.localItems.RLock()
	localItem, found := db.localItems.m[key]
	db.localItems.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetRepubTime, no time matching key: %x", key)
		return
	}

	return localItem.repubTime, nil
}
*/

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

func (db *Database) evictItem(key Key) {
	db.items.Lock()
	delete(db.items.m, key)
	db.items.Unlock()
}

// Start this as a Goroutine at node start.
func (db *Database) itemHandler() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
		var evictees []Key

		db.items.RLock()
		for key, item := range db.items.m {
			if now.After(item.expire) || now.After(item.republish) {
				evictees = append(evictees, key)
			}
		}
		db.items.RUnlock()

		for _, key := range evictees {
			db.evictItem(key)
		}
	}
}

// Start as Goroutine.
// Function is responsible for republishing data that should persist on storage network.
func (db *Database) republishHandler() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
		replicate := now.After(db.getReplicate()) // This line is a suspect if something doesnt work.

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
