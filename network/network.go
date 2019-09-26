package network

import (
	"net"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

const Size256 = 256 / 8

type SessionID [Size256]byte

type PingResult struct {
	Challenge []byte
}

type PongRequest struct {
	SessionID SessionID
	Challenge []byte
}

type FindNodesResult struct {
	closest []route.Contact
}

type FindValueResult struct {
	SessionID SessionID
	closest   []route.Contact
	Key       store.Key
	value     string
}

type FindNodesRequest struct{}

type FindValueRequest struct{}

type Network interface {
	Ping(addr net.UDPAddr) (chan *PingResult, error)
	Pong(challenge []byte, sessionID SessionID, addr net.UDPAddr) error
	FindNodes(target node.ID, addr net.UDPAddr) (chan Result, error)
	Store(key store.Key, value string, addr net.UDPAddr) error
	FindValue(key store.Key, addr net.UDPAddr) (chan Result, error)
	SendValue(key store.Key, value string, closets []route.Contact, sessionID SessionID, addr net.UDPAddr) error
}

type Result interface {
	Closest() []route.Contact
	Value() string
}

func (r *FindNodesResult) Closest() []route.Contact {
	return r.closest
}

func (r *FindNodesResult) Value() string {
	return ""
}

func (r *FindValueResult) Closest() []route.Contact {
	return r.closest
}

func (r *FindValueResult) Value() string {
	return r.value
}
