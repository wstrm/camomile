package store

//import "log"
import "time"
import "golang.org/x/crypto/blake2b"
import "sync"

type NodeID [32]byte
type Key NodeID

var db *database

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
}

// TODO: Max 1000 chars. Truncate input string.
func AddItem(value string, origPub NodeID) error {
	t := time.Now()
	expire := t.Add(time.Second * 86400)
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
	republish := t.Add(time.Second * 86400)

	db.keys.Lock()
	db.keys.m[key] = republish
	db.keys.Unlock()
}

/*
func (i *items) evict(key Key) {
	delete(*items, key)
}

func (k *storedKeys) evict(key Key) {
	delete (*storedKeys, key)
}

// Start this as a Goroutine at node start.
func (i *items) itemHandler() {
	timer := time.NewTimer(time.Second * 1)
	for key, item := range items {
		if item.expire.After(time.Now()) || item.republish.After(time.Now())
			items.evict(key)
	}
}

// Start as Goroutine.
// Function is responsible for republishing data that should persist on storage network.
func (k *storedKeys) republisher(repub chan storedKey) {
	timer := time.NewTimer(time.Second * 1)
	for key :=range storedKeys {
		if key.republish.After(time.Now())
			//TODO: chan to communicate with RPC to send store command.
			repub <- storedKey
	}
}
*/
