package ctl

import (
	"log"
	"os"
	"time"

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
	Value string
}

type Get struct {
	Key store.Key
}

type Exit struct{}

func NewAPI(dht *dht.DHT) *API {
	return &API{dht: dht}
}

func (a *API) Ping(ping Ping, reply *[]byte) (err error) {
	log.Printf("Ping: %s\n", ping.NodeID)
	*reply, err = a.dht.Ping(ping.NodeID)
	return
}

func (a *API) Put(put Put, reply *store.Key) (err error) {
	log.Printf("Put: %s\n", put.Value)
	*reply, err = a.dht.Put(put.Value)
	return
}

func (a *API) Get(get Get, reply *string) (err error) {
	log.Printf("Get: %s\n", get.Key)
	*reply, err = a.dht.Get(get.Key)
	return
}

func (a *API) Exit(exit Exit, ok *bool) error {
	log.Println("Terminating node in 5 seconds...")

	*ok = true

	go func() {
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	return nil
}
