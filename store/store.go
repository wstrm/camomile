package store

//import "log"
import "time"
import "golang.org/x/crypto/blake2b"
import "sync"
import "fmt"

type NodeID [32]byte
type Key NodeID

var db *database

var tExpire = 86400 * time.Second
var tReplicate = 3600 * time.Second
var tRepublish = 86400 * time.Second


type item struct {
	value     string
	expire    time.Time
	republish time.Time
	origPub   NodeID
}

type items struct {
	sync.RWMutex
	m map[Key]item
}

type keys struct {
	sync.RWMutex
	m map[Key]time.Time
}

type replicate struct {
	sync.RWMutex
	time time.Time
}

type database struct {
	items     items
	keys      keys
	ch        chan (Key)
	replicate replicate
}

func setup() {
	db = new(database)
	db.items = items{m: make(map[Key]item)}
	db.keys = keys{m: make(map[Key]time.Time)}
	db.ch = make(chan Key)
	setReplicate()
}

func init() {
	setup()

	go itemHandler()
	go republisher()
}

func setReplicate() {
	db.replicate.Lock()
	db.replicate.time = time.Now().Add(time.Duration(tReplicate))
	db.replicate.Unlock()
}

func getReplicate() time.Time {
	db.replicate.RLock()
	time := db.replicate.time
	db.replicate.RUnlock()

	return time
}

// TODO: Max 1000 chars. Truncate input string.
func AddItem(value string, origPub NodeID) error {
	t := time.Now()
	expire := t.Add(tExpire)
	republish := expire

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

func AddKey(key Key) {
	t := time.Now()
	republish := t.Add(tRepublish)

	db.keys.Lock()
	db.keys.m[key] = republish
	db.keys.Unlock()
}

func GetItem(key Key) (reqItem item, err error) {
	db.items.RLock()
	requestedItem, found := db.items.m[key]
	db.items.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetItem, no item matching key: %x", key)
		return
	}

	return requestedItem, nil
}

func GetRepubTime(key Key) (repubTime time.Time, err error) {
	db.keys.RLock()
	repubTime, found := db.keys.m[key]
	db.keys.RUnlock()

	if !found {
		err = fmt.Errorf("store: GetRepubTime, no time matching key: %x", key)
		return
	}

	return repubTime, nil
}

func evictItem(key Key) {
	db.items.Lock()
	delete(db.items.m, key)
	db.items.Unlock()
}

// Start this as a Goroutine at node start.
func itemHandler() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
		var evictees []Key

		db.items.RLock()
		for key, item := range db.items.m {
			if item.expire.After(now) || item.republish.After(now) {
				evictees = append(evictees, key)
			}
		}
		db.items.RUnlock()

		for _, key := range evictees {
			evictItem(key)
		}
	}
}

// Start as Goroutine.
// Function is responsible for republishing data that should persist on storage network.
func republisher() {
	for {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		now := time.Now()
		replTime := getReplicate()
		replicate := replTime.After(now)

		db.keys.RLock()
		for key, repubTime := range db.keys.m {
			if repubTime.After(now) {
				db.ch <- key
			} else if replicate {
				// Replication event, replicate all stored values to k nodes.
				db.ch <- key
			}
		}
		db.keys.RUnlock()

		// If a replication event just happened, reset the replication timer.
		if replicate {
			setReplicate()
		}
	}
}
