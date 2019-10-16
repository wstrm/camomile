package main

import (
	"net"
	"net/rpc"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/ctl"
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

	time.Sleep(2 * time.Second) // TODO: Remove this.

	go rpcServe(dht)

	time.Sleep(1 * time.Second) // TODO: Remove this.

	client, err := rpc.DialHTTP("tcp", defaultRPCAddress)
	if err != nil {
		t.Error("dialing:", err)
	}

	put := ctl.Put{
		Value: "something",
	}
	var key store.Key
	err = client.Call("API.Put", put, &key)
	if err != nil {
		t.Error(err)
	}

	get := ctl.Get{
		Key: [32]byte{},
	}
	var value ctl.GetReply
	err = client.Call("API.Get", get, &value)
	if err == nil {
		t.Error("expected error, value shouldn't exist")
	}

	forget := ctl.Forget{
		Key: [32]byte{},
	}
	var ok bool
	err = client.Call("API.Forget", forget, &ok)
	if err != nil {
		t.Error(err)
	}
}
