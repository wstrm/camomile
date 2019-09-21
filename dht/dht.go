package dht

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

type Key node.ID // TODO(optmzr): use store.Key instead.

type DHT struct {
	rt *route.Table
}

// TODO(optmzr): Move to store.
func KeyFromString(str string) (key Key, err error) {
	h, err := hex.DecodeString(str)
	if err != nil {
		err = fmt.Errorf("cannot decode hex string as key: %w", err)
		return
	}

	copy(key[:], h)
	return
}

func NewDHT(me route.Contact, others []route.Contact) (dht *DHT, err error) {
	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	return
}

// Get retrieves the value for a specified key from the network.
func (dht *DHT) Get(hash Key) (value string, err error) {
	// TODO(optmzr)
	return "", errors.New("Not Implemented")
}

// Put stores the provided value in the network and returns a key.
func (dht *DHT) Put(value string) (hash Key, err error) {
	// TODO(optmzr)
	return Key{}, errors.New("Not Implemented")
}

// Join initiates a node lookup of itself to bootstrap the node into the
// network.
func (dht *DHT) Join(me route.Contact) (err error) {
	// TODO(optmzr)
	return errors.New("Not Implemented")
}
