package cmd

import (
	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/store"
)

type API struct {
	dht *dht.DHT
}

type Ping struct {
	NodeID node.ID
}

type Put struct {
	Val string
}

type Get struct {
	Key store.Key
}

type Exit struct {
	NodeID node.ID
}

func (a *API) Ping(ping Ping, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Put(put Put, reply *store.Key) (err error) {
	*reply, err = a.dht.Put(put.Val)
	return
}

func (a *API) Get(get Get, reply *string) (err error) {
	*reply, err = a.dht.Get(get.Key)
	return
}

func (a *API) Exit(exit Exit, reply *bool) error {
	*reply = true
	return nil
	//TODO
}
