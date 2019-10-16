package main

import (
	"net"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

func TestRPCServe(t *testing.T) {
	local, _ := net.ResolveUDPAddr("udp", "localhost:1238")
	me := route.Contact{
		NodeID:  node.NewID(),
		Address: *local,
	}

	others := []route.Contact{
		route.Contact{
			NodeID:  node.NewID(),
			Address: *local,
		},
	}

	nw, _ := network.NewUDPNetwork(me)
	dht, _ := dht.New(me, others, nw)

	go func() {
		err := nw.Listen()
		if err != nil {
			t.Error(err)
		}
	}()

	go rpcServe(dht)
	time.Sleep(1 * time.Second) // TODO: Remove this.
}
