package dht

import (
	"net"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/store"
)

type Call interface {
	Do(nw network.Network, address net.UDPAddr) (ch chan network.Result, err error)
	Result(result network.Result) (stop bool)
	Target() (target node.ID)
}

func NewFindNodesCall(target node.ID) *FindNodesCall {
	return &FindNodesCall{
		target: target,
	}
}

type FindNodesCall struct {
	target node.ID
}

func (q *FindNodesCall) Do(nw network.Network, address net.UDPAddr) (chan network.Result, error) {
	return nw.FindNodes(q.target, address)
}

func (q *FindNodesCall) Result(_ network.Result) (_ bool) { return }
func (q *FindNodesCall) Target() node.ID                  { return q.target }

func NewFindValueCall(hash store.Key) *FindValueCall {
	return &FindValueCall{
		hash: hash,
	}
}

type FindValueCall struct {
	hash  store.Key
	value string
}

func (q *FindValueCall) Do(nw network.Network, address net.UDPAddr) (chan network.Result, error) {
	return nw.FindValue(q.hash, address)
}

func (q *FindValueCall) Result(result network.Result) (stop bool) {
	// TODO: Value validation could be added here, where the value received is
	// checked towards the expected hash.

	q.value = result.Value()
	if q.value != "" {
		stop = true
	} else {
		stop = false
	}

	return
}

func (q *FindValueCall) Target() node.ID { return node.ID(q.hash) }
