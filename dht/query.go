package dht

import (
	"net"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

type NetworkResult struct {
	From    route.Contact
	Closest []route.Contact
}

type NetworkFindNodesResult struct {
	NetworkResult
}

type NetworkFindValueResult struct {
	NetworkResult
	Value string
}

type query interface {
	Call(nw Network, address net.UDPAddr) (ch chan *NetworkResult, err error)
	Update(result *NetworkResult) (done bool)
	Target() (target node.ID)
}

type queryFindNodes struct {
	target node.ID
}

func (q *queryFindNodes) Call(nw Network, address net.UDPAddr) (ch chan *NetworkResult, err error) {
	return nw.FindNodes(q.target, address)
}

func (q *queryFindNodes) Update(_ *NetworkResult) bool {
	return false // Do the whole walk, never return a done.
}

func (q *queryFindNodes) Target() node.ID {
	return q.target
}

type queryFindValue struct {
	hash  Key
	value string
}

func (q *queryFindValue) Call(nw Network, address net.UDPAddr) (ch chan *NetworkResult, err error) {
	return nw.FindValue(q.hash, address)
}

func (q *queryFindValue) Update(result *NetworkResult) bool {
	vr, ok := result.(*NetworkFindValueResult)
	if !ok {
		return false
	}
	q.value = vr.Value
	return true
}

func (q *queryFindValue) Target() node.ID {
	return node.ID(q.hash)
}
