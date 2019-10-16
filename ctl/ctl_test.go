package ctl

import (
	"net"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
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

	time.Sleep(1 * time.Second) // TODO: Remove this.

	api := NewAPI(dht)

	var pingReply []byte
	err := api.Ping(Ping{NodeID: others[0].NodeID}, &pingReply)
	if err != nil {
		t.Error(err)
	}

	var putReply store.Key
	err = api.Put(Put{Value: "something"}, &putReply)
	if err != nil {
		t.Error(err)
	}

	var getReply GetReply
	err = api.Get(Get{Key: putReply}, &getReply)
	if err != nil {
		t.Error(err)
	}

	var forgetReply bool
	err = api.Forget(Forget{Key: putReply}, &forgetReply)
	if err != nil {
		t.Error(err)
	}

	var exitReply bool
	err = api.Exit(Exit{}, &exitReply)
	if err != nil {
		t.Error(err)
	}
}
