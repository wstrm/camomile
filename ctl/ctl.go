package ctl

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

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

type Forget struct {
	Key store.Key
}

type Exit struct{}

type GetReply struct {
	Value    string
	SenderID node.ID
}

func NewAPI(dht *dht.DHT) *API {
	return &API{dht: dht}
}

func (a *API) Ping(ping Ping, reply *[]byte) (err error) {
	log.Info().Msgf("Ping: %s", ping.NodeID)
	*reply, err = a.dht.Ping(ping.NodeID)
	return
}

func (a *API) Put(put Put, reply *store.Key) (err error) {
	log.Info().Msgf("Put: %s", put.Value)
	*reply, err = a.dht.Put(put.Value)
	return
}

func (a *API) Get(get Get, reply *GetReply) (err error) {
	log.Info().Msgf("Get: %s", get.Key)
	reply.Value, reply.SenderID, err = a.dht.Get(get.Key)
	return
}

func (a *API) Forget(forget Forget) {
	log.Info().Msgf("Forget: %s", forget.Key)
	a.dht.Forget(forget.Key)
}

func (a *API) Exit(exit Exit, ok *bool) error {
	log.Info().Msg("Terminating node in 5 seconds...")

	*ok = true

	go func() {
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	return nil
}
