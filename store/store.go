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

type items map[Key]item
type storedKeys map[Key]time.Time

func (i *items) add(value [255]byte, origPub NodeID) (error) {
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

	i[key] = newItem
}

func (k *storedKeys) add(key Key) {
	t := time.Now()
	republish := t.Add(time.Second * 86400)

	newStoredKey := new(storedKey)

	newStoredKey.key = key
	newStoredKey.republish = republish

	k[key] = newStoredKey
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
