package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/optmzr/d7024e-dht/ctl"
	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

func rpcServe(dht *dht.DHT) {
	api := ctl.NewAPI(dht)

	err := rpc.Register(api)
	if err != nil {
		log.Fatalln(err)
	}

	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	err = http.Serve(l, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	address, err := net.ResolveUDPAddr("udp", network.UdpPort)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO
	others := []route.Contact{
		route.Contact{
			NodeID:  node.NewID(),
			Address: *address,
		},
		route.Contact{
			NodeID:  node.NewID(),
			Address: *address,
		},
		route.Contact{
			NodeID:  node.NewID(),
			Address: *address,
		},
	}

	me := route.Contact{
		NodeID:  node.NewID(),
		Address: *address,
	}

	log.Printf("My node ID is: %v", me.NodeID)

	nw, err := network.NewUDPNetwork(me)
	if err != nil {
		log.Fatalln(err)
	}

	dht, err := dht.New(me, others, nw)
	if err != nil {
		log.Fatalln(err)
	}

	go rpcServe(dht)

	err = nw.Listen()
	if err != nil {
		log.Fatalln(err)
	}
}
