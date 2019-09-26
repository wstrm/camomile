package ctl

import (
	"log"
	"os"

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

type Exit struct{}

func NewAPI(dht *dht.DHT) *API {
	return &API{dht: dht}
}

func (a *API) Ping(ping Ping, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Put(put Put, reply *store.Key) (err error) {
	log.Printf("Put: %s\n", put.Val)
	*reply, err = a.dht.Put(put.Val)
	return
}

func (a *API) Get(get Get, reply *string) (err error) {
	log.Printf("Get: %s\n", get.Key)
	*reply, err = a.dht.Get(get.Key)
	return
}

func (a *API) Exit(exit Exit, reply *bool) error {
	log.Println("Exit")
	*reply = true
	defer os.Exit(0)
	return nil
}
