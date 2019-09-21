package dht

import (
	"fmt"

	"github.com/optmzr/d7024e-dht/route"
)

type DHT struct {
	rt *route.Table
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
