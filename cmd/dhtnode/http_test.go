package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/rs/zerolog/log"
)

func TestHTTPHandler(t *testing.T) {
	local, _ := net.ResolveUDPAddr("udp", "localhost:1234")
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
			log.Error(err)
		}
	}()

	handler := newHTTPHandler(dht)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	_, err := http.Head(ts.URL)
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}

	_, err = http.PostForm(ts.URL, url.Values{"value": {"ABC, du är mina tankar"}})
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}

	_, err = http.Get(ts.URL + "/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386")
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}

	_, err = http.PostForm(ts.URL, url.Values{"value": {""}})
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}

	_, err = http.Get(ts.URL + "/invalid")
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", ts.URL+"/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386", nil)
	_, err = client.Do(req)
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}
}

func TestNewHTTPHandler(t *testing.T) {
	local, _ := net.ResolveUDPAddr("udp", "localhost:1234")
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
			log.Error(err)
		}
	}()

	go httpServe(dht)

	_, err := http.Get(defaultHTTPAddress)
	if err != nil {
		t.Errorf("unexpected response: %v", err)
	}
}
