package store

import "fmt"
import "log"
import "time"
import "golang.org/x/crypto/blake2b"

type NodeID [20]byte
type Key NodeID

type item struct {
	value		[255]byte
	expire		time.Time
	republish	time.Time
	origPub		NodeID
}

type storedKey struct {
	key		Key
	republish	time.Time
}

type items map[Key]item
type storedKeys map[Key]storedKey

func addItem(value [255]byte, origPub NodeID) (error) {
	t := time.Now()
	expire := t.Add(time.Second * 86400)
	republish := expire

	h, err := blake2b.New(160, value)
	if err != nil {
		return err
	}
	key := h.sum()

	newItem := new(item)

	newItem.value = value
	newItem.expire = expire
	newItem.republish = republish
	newItem.origPub = origPub

	items[key] = newItem
}

func addStoredKey(key Key) {
	t := time.Now()
	republish := t.Add(time.Second * 86400)

	newStoredKey := new(storedKey)

	newStoredKey.key = key
	newStoredKey.republish = republish

	storedKeys[key] = newStoredKey
}

func evictItem(key Key) {
	delete(items, key)
}

// Start this as a Goroutine at node start.
func itemHandler() {
	timer := time.NewTimer(time.Second * )

}
