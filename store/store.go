package store

//import "log"
import "time"
import "golang.org/x/crypto/blake2b"
import "sync"
import "fmt"

type NodeID [32]byte
type Key NodeID

var db *database

var tExpire int = 86400
var tReplicate int = 3600
var tRepublish int = 86400

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

type database struct {
	items	items
	keys	keys
	ch	chan(Key)
}

func init() {
	db = new(database)
	db.items = items{m: make(map[Key]item)}
	db.keys = keys{m: make(map[Key]time.Time)}
	db.ch = make(chan Key)

	go itemHandler()
	go republisher()
}

func setTimers(tExpire int, tReplicate int, tRepublish int) {
	tExpire = tExpire
	tReplicate = tReplicate
	tRepublish = tRepublish
}

// TODO: Max 1000 chars. Truncate input string.
func AddItem(value string, origPub NodeID) error {
	t := time.Now()
	expire := t.Add(time.Second * time.Duration(tExpire))
	republish := expire

	var key Key
	key = blake2b.Sum256([]byte(value))

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
	republish := t.Add(time.Second * time.Duration(tRepublish))

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
	delete(db.items.m, key)
}

func evictKey(key Key) {
	delete (db.keys.m, key)
}

// Start this as a Goroutine at node start.
func itemHandler() {
	for true {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C
		for key, item := range db.items.m {
			if item.expire.After(time.Now()) || item.republish.After(time.Now()) {
				evictItem(key)
			}
		}
	}
}


// Start as Goroutine.
// Function is responsible for republishing data that should persist on storage network.
func republisher() {
	for true {
		timer := time.NewTimer(time.Second * 1)
		<-timer.C
		for key, repubTime := range db.keys.m {
			if repubTime.After(time.Now()) {
				db.ch <-key
			}
		}
	}
}

