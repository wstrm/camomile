package network

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/packet"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
)

const Size256 = 256 / 8

const networkTimeout = 5 * time.Second

type SessionID [Size256]byte

type randRead func([]byte) (int, error)

var rng randRead

func init() {
	rng = rand.Read
}

type udpNetwork struct {
	conn  *net.UDPConn
	me    route.Contact
	fnt   *table
	fvt   *table
	pt    *pingTable
	fnr   chan *FindNodesRequest
	fvr   chan *FindValueRequest
	pr    chan *PongRequest
	sr    chan *StoreRequest
	ready chan struct{}
}

type PingResult struct {
	From      route.Contact
	Challenge []byte
}

type PongRequest struct {
	From      route.Contact
	SessionID SessionID
	Challenge []byte
}

type StoreRequest struct {
	Value string
	From  route.Contact
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

type Network interface {
	Ping(addr net.UDPAddr) (chan *PingResult, []byte, error)
	Pong(challenge []byte, sessionID SessionID, addr net.UDPAddr) error
	FindNodes(target node.ID, addr net.UDPAddr) (chan Result, error)
	Store(key store.Key, value string, addr net.UDPAddr) error
	FindValue(key store.Key, addr net.UDPAddr) (chan Result, error)
	SendValue(key store.Key, value string, closest []route.Contact, sessionID SessionID, addr net.UDPAddr) error
	SendNodes(closest []route.Contact, sessionID SessionID, addr net.UDPAddr) error
	FindNodesRequestCh() chan *FindNodesRequest
	FindValueRequestCh() chan *FindValueRequest
	StoreRequestCh() chan *StoreRequest
	PongRequestCh() chan *PongRequest
	ReadyCh() chan struct{}
	Listen() error
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

type FindNodesRequest struct {
	SessionID SessionID
	Target    node.ID
	From      route.Contact
}

type FindValueRequest struct {
	Key       store.Key
	SessionID SessionID
	From      route.Contact
}

func NewUDPNetwork(me route.Contact) (Network, error) {
	fvtTicker := time.NewTicker(time.Second)
	fntTicker := time.NewTicker(time.Second)

	n := &udpNetwork{
		me:  me,
		fvt: newTable(networkTimeout, fvtTicker),
		fnt: newTable(networkTimeout, fntTicker),
		pt:  newPingTable(),
	}

	n.fnr = make(chan *FindNodesRequest)
	n.fvr = make(chan *FindValueRequest)
	n.sr = make(chan *StoreRequest)
	n.pr = make(chan *PongRequest)
	n.ready = make(chan struct{})

	return n, nil
}

func (u *udpNetwork) StoreRequestCh() chan *StoreRequest         { return u.sr }
func (u *udpNetwork) FindNodesRequestCh() chan *FindNodesRequest { return u.fnr }
func (u *udpNetwork) FindValueRequestCh() chan *FindValueRequest { return u.fvr }
func (u *udpNetwork) PongRequestCh() chan *PongRequest           { return u.pr }
func (u *udpNetwork) ReadyCh() chan struct{}                     { return u.ready }

func (u *udpNetwork) Ping(addr net.UDPAddr) (chan *PingResult, []byte, error) {
	id := generateID()
	c := generateChallenge()

	payload := &packet.Ping{
		Challenge: c,
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_Ping{Ping: payload},
	}

	err := u.send(addr, *p)
	if err != nil {
		return nil, nil, err
	}

	results := make(chan *PingResult)
	u.pt.Put(id, results)

	return results, c, nil
}

func (u *udpNetwork) Pong(challenge []byte, sessionID SessionID, addr net.UDPAddr) error {
	payload := &packet.Pong{
		Challenge: challenge,
	}
	p := &packet.Packet{
		SessionId: sessionID[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_Pong{Pong: payload},
	}

	return u.send(addr, *p)
}

func (u *udpNetwork) FindNodes(target node.ID, addr net.UDPAddr) (chan Result, error) {
	id := generateID()

	payload := &packet.FindNode{
		NodeId: target[:],
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_FindNode{FindNode: payload},
	}

	results := make(chan Result)
	u.fnt.Put(id, results)

	err := u.send(addr, *p)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (u *udpNetwork) Store(key store.Key, value string, addr net.UDPAddr) error {
	id := generateID()

	payload := &packet.Store{
		Key:   key[:],
		Value: value,
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_Store{Store: payload},
	}

	return u.send(addr, *p)
}

func (u *udpNetwork) FindValue(key store.Key, addr net.UDPAddr) (chan Result, error) {
	id := generateID()

	payload := &packet.FindValue{
		Key: key[:],
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_FindValue{FindValue: payload},
	}

	results := make(chan Result)
	u.fvt.Put(id, results)

	err := u.send(addr, *p)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (u *udpNetwork) SendValue(key store.Key, value string, closest []route.Contact, sessionID SessionID, addr net.UDPAddr) error {
	var nodes []*packet.NodeInfo
	var contacts []route.Contact

	contacts = append(contacts, closest...)

	for _, c := range contacts {
		p := &packet.NodeInfo{
			NodeId: c.NodeID.Bytes(),
			Ip:     c.Address.IP,
			Port:   uint32(c.Address.Port),
		}
		nodes = append(nodes, p)
	}

	internalPayload := &packet.NodeList{
		Nodes: nodes,
	}

	payload := &packet.Value{
		Key:      key[:],
		Value:    value,
		NodeList: internalPayload,
	}
	p := &packet.Packet{
		SessionId: sessionID[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_Value{Value: payload},
	}

	err := u.send(addr, *p)
	if err != nil {
		return err
	}

	return nil
}

func (u *udpNetwork) SendNodes(closest []route.Contact, sessionID SessionID, addr net.UDPAddr) error {
	var nodes []*packet.NodeInfo

	for _, c := range closest {
		p := &packet.NodeInfo{
			NodeId: c.NodeID.Bytes(),
			Ip:     c.Address.IP,
			Port:   uint32(c.Address.Port),
		}
		nodes = append(nodes, p)
	}

	payload := &packet.NodeList{
		Nodes: nodes,
	}

	p := &packet.Packet{
		SessionId: sessionID[:],
		SenderId:  u.me.NodeID.Bytes(),
		Payload:   &packet.Packet_NodeList{NodeList: payload},
	}

	err := u.send(addr, *p)
	if err != nil {
		return err
	}

	return nil
}

func (u *udpNetwork) Listen() (err error) {
	log.Info().
		Msgf("Listening for UDP packets on: %s", u.me.Address.String())

	u.conn, err = net.ListenUDP("udp", &u.me.Address)
	if err != nil {
		return err
	}
	defer u.conn.Close()

	// Notify everyone that we're ready.
	u.ready <- struct{}{}

	// Reusable buffer, can be used between every read loop as it will be copied
	// before sending the data to the packet handler.
	buffer := make([]byte, 65535)

	for {
		n, addr, err := u.conn.ReadFromUDP(buffer)

		log.Debug().
			Msgf("Received packet from: %v", addr)

		if err != nil {
			log.Error().Err(err).
				Msgf("Error when reading from UDP from address %v: %s",
					addr, err)
			continue
		}

		// Make a copy of the current data in the buffer.
		rawPacket := make([]byte, n)
		copy(rawPacket, buffer)

		go u.handlePacket(rawPacket, *addr)
	}
}

func logChannelNotFound(id SessionID) {
	log.Warn().
		Msgf("Channel with ID: %x not found in table", id)
}

func (u *udpNetwork) handlePacket(b []byte, addr net.UDPAddr) {
	p := &packet.Packet{}
	err := proto.Unmarshal(b, p)
	if err != nil {
		log.Error().Err(err).
			Msg("Error unserializing packet")

		return
	}

	switch p.Payload.(type) {
	case *packet.Packet_Value:
		var sessionID SessionID
		var senderID node.ID
		var key store.Key
		var closest []route.Contact
		copy(sessionID[:], p.SessionId)
		copy(senderID[:], p.SenderId)
		copy(key[:], p.GetValue().Key)

		for _, contact := range p.GetValue().GetNodeList().GetNodes() {
			closest = append(closest, route.Contact{
				NodeID: node.IDFromBytes(contact.NodeId),
				Address: net.UDPAddr{
					IP:   contact.Ip,
					Port: int(contact.Port),
					Zone: "",
				},
			})
		}

		ch, ok := u.fvt.Get(sessionID)
		if !ok {
			logChannelNotFound(sessionID)
			return
		}

		ch <- &FindValueResult{
			SessionID: sessionID,
			closest:   closest,
			Key:       key,
			value:     p.GetValue().Value,
		}

		u.fvt.Remove(sessionID)

	case *packet.Packet_NodeList:
		var sessionID SessionID
		var senderID node.ID
		var closest []route.Contact
		copy(sessionID[:], p.SessionId)
		copy(senderID[:], p.SenderId)

		for _, contact := range p.GetNodeList().GetNodes() {
			closest = append(closest, route.Contact{
				NodeID: node.IDFromBytes(contact.NodeId),
				Address: net.UDPAddr{
					IP:   contact.Ip,
					Port: int(contact.Port),
					Zone: "",
				},
			})
		}

		ch, ok := u.fnt.Get(sessionID)
		if !ok {
			logChannelNotFound(sessionID)
			return
		}

		ch <- &FindNodesResult{
			closest: closest,
		}

		u.fnt.Remove(sessionID)

	case *packet.Packet_FindValue:
		var key store.Key
		var senderID node.ID
		var sessionID SessionID
		copy(key[:], p.GetFindValue().Key)
		copy(senderID[:], p.GetSenderId())
		copy(sessionID[:], p.GetSessionId())

		u.fvr <- &FindValueRequest{
			Key:       key,
			SessionID: sessionID,
			From: route.Contact{
				NodeID: senderID,
				Address: net.UDPAddr{
					IP:   addr.IP,
					Port: addr.Port,
				},
			},
		}

	case *packet.Packet_Ping:
		var sessionID SessionID
		var senderID node.ID
		copy(senderID[:], p.GetSenderId())
		copy(sessionID[:], p.GetSessionId())

		u.pr <- &PongRequest{
			From: route.Contact{
				NodeID: senderID,
				Address: net.UDPAddr{
					IP:   addr.IP,
					Port: addr.Port,
				},
			},
			SessionID: sessionID,
			Challenge: p.GetPing().GetChallenge(),
		}

	case *packet.Packet_Pong:
		var sessionID SessionID
		var senderID node.ID
		copy(senderID[:], p.GetSenderId())
		copy(sessionID[:], p.GetSessionId())

		ch, ok := u.pt.Get(sessionID)
		if !ok {
			logChannelNotFound(sessionID)
			return
		}

		ch <- &PingResult{
			From: route.Contact{
				NodeID: senderID,
				Address: net.UDPAddr{
					IP:   addr.IP,
					Port: addr.Port,
				},
			},
			Challenge: p.GetPong().GetChallenge(),
		}

		u.pt.Remove(sessionID)

	case *packet.Packet_FindNode:
		var sessionID SessionID
		var senderID node.ID
		var targetID node.ID
		copy(sessionID[:], p.GetSessionId())
		copy(senderID[:], p.GetSenderId())
		copy(targetID[:], p.GetFindNode().NodeId)

		u.fnr <- &FindNodesRequest{
			SessionID: sessionID,
			Target:    targetID,
			From: route.Contact{
				NodeID: senderID,
				Address: net.UDPAddr{
					IP:   addr.IP,
					Port: addr.Port,
				},
			},
		}

	case *packet.Packet_Store:
		var senderID node.ID
		copy(senderID[:], p.GetSenderId())
		value := p.GetStore().Value

		u.sr <- &StoreRequest{
			Value: value,
			From: route.Contact{
				NodeID: senderID,
				Address: net.UDPAddr{
					IP:   addr.IP,
					Port: addr.Port,
				},
			},
		}

	default:
		log.Debug().
			Msgf("Unhandled packet: %v", p)
	}
}

func generateID() (id SessionID) {
	_, err := rng(id[:])
	if err != nil {
		panic(err)
	}
	return id
}

func generateChallenge() []byte {
	c := make([]byte, Size256)
	_, err := rng(c)
	if err != nil {
		panic(err)
	}
	return c
}

func (u *udpNetwork) send(addr net.UDPAddr, packet packet.Packet) error {
	b, err := proto.Marshal(&packet)
	if err != nil {
		return err
	}
	_, err = u.conn.WriteTo(b, &addr)
	if err != nil {
		return err
	}
	return nil
}

func (id SessionID) String() string {
	return hex.EncodeToString(id[:])
}
