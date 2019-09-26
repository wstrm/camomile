package network

import (
	"crypto/rand"
	"github.com/golang/protobuf/proto"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/packet"
	"github.com/optmzr/d7024e-dht/route"
	"log"
	"net"
)

const UdpPort = ":8118"

type randRead func([]byte) (int, error)

var rng randRead

func init() {
	rng = rand.Read
}

type Key node.ID
type SessionID [node.IDBytesLength]byte

type Network interface {
	Ping(addr net.UDPAddr) (chan *PingResult, error)
	Pong(challenge []byte, sessionID SessionID, addr net.UDPAddr) error
	FindNodes(target node.ID, addr net.UDPAddr) (chan *FindNodesResult, error)
	Store(key Key, value string, addr net.UDPAddr) error
	FindValue(key Key, addr net.UDPAddr) (chan *FindValueResult, error)
	SendValue(key Key, value string, closets []route.Contact, sessionID SessionID, addr net.UDPAddr) error
}

type udpNetwork struct {
	me  node.ID
	fnt *findNodesTable
	fvt *findValueTable
	pt  *pingTable
	fnr chan *FindNodesRequest
	fvr chan *FindValueRequest
	pr  chan *PongRequest
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

type FindNodesResult struct {
	From    route.Contact
	Closest []route.Contact
}

type FindValueResult struct {
	From      route.Contact
	SessionID SessionID
	Closest   []route.Contact
	Key       Key
	Value     string
}

type FindNodesRequest struct {
	SessionID SessionID
	sender    route.Contact
}

type FindValueRequest struct {
	key       Key
	sessionID SessionID
	sender    route.Contact
}

func NewUDPNetwork(id node.ID) (Network, chan *FindNodesRequest, chan *FindValueRequest, chan *PongRequest) {
	n := &udpNetwork{me: id, fvt: newFindValueTable(), fnt: newFindNodesTable(), pt: newPingTable()}

	n.fnr = make(chan *FindNodesRequest)
	n.fvr = make(chan *FindValueRequest)
	n.pr = make(chan *PongRequest)

	go n.listen()

	return n, n.fnr, n.fvr, n.pr
}

func (u *udpNetwork) Ping(addr net.UDPAddr) (chan *PingResult, error) {
	id := generateID()

	payload := &packet.Ping{
		Challenge: generateChallenge(),
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_Ping{Ping: payload},
	}

	err := send(&addr, *p)
	if err != nil {
		return nil, err
	}

	results := make(chan *PingResult)
	u.pt.Put(id, results)

	return results, nil
}

func (u *udpNetwork) Pong(challenge []byte, sessionID SessionID, addr net.UDPAddr) error {
	payload := &packet.Pong{
		Challenge: challenge,
	}
	p := &packet.Packet{
		SessionId: sessionID[:],
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_Pong{Pong: payload},
	}

	return send(&addr, *p)
}

func (u *udpNetwork) FindNodes(target node.ID, addr net.UDPAddr) (chan *FindNodesResult, error) {
	id := generateID()

	payload := &packet.FindNode{
		NodeId: target[:],
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_FindNode{FindNode: payload},
	}

	err := send(&addr, *p)
	if err != nil {
		return nil, err
	}

	results := make(chan *FindNodesResult)
	u.fnt.Put(id, results)

	return results, nil
}

func (u *udpNetwork) Store(key Key, value string, addr net.UDPAddr) error {
	id := generateID()

	payload := &packet.Store{
		Key:   key[:],
		Value: value,
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_Store{Store: payload},
	}

	return send(&addr, *p)
}

func (u *udpNetwork) FindValue(key Key, addr net.UDPAddr) (chan *FindValueResult, error) {
	id := generateID()

	payload := &packet.FindValue{
		Key: key[:],
	}
	p := &packet.Packet{
		SessionId: id[:],
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_FindValue{FindValue: payload},
	}

	err := send(&addr, *p)
	if err != nil {
		return nil, err
	}

	results := make(chan *FindValueResult)
	u.fvt.Put(id, results)

	return results, nil
}

func (u *udpNetwork) SendValue(key Key, value string, closest []route.Contact, sessionID SessionID, addr net.UDPAddr) error {

	var nodes []*packet.NodeInfo
	var contacts []route.Contact

	contacts = append(contacts, closest...)

	for _, c := range contacts {
		p := &packet.NodeInfo{
			NodeId: c.NodeID[:],
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
		SenderId:  u.me.Bytes(),
		Payload:   &packet.Packet_Value{Value: payload},
	}

	err := send(&addr, *p)
	if err != nil {
		return err
	}

	return nil
}

func (u *udpNetwork) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", UdpPort)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Printf("Listening for UDP packets on port %v", UdpPort)
	for {
		data := make([]byte, 1500)
		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Fatalf("Error when reading from UDP from address %v: %s", addr.String(), err)
		}
		b := data[:n]
		go u.handlePacket(b, *addr)
	}
}

func (u *udpNetwork) handlePacket(b []byte, addr net.UDPAddr) {
	p := &packet.Packet{}
	err := proto.Unmarshal(b, p)
	if err != nil {
		log.Println("Error unserializing packet", err)
		return
	}

	switch p.Payload.(type) {
	case *packet.Packet_Value:
		var sessionID SessionID
		var senderID node.ID
		var key Key
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

		ch := u.fvt.Get(sessionID)
		if ch == nil {
			log.Println("Channel not found in table")
			return
		}

		ch <- &FindValueResult{
			From: route.Contact{
				NodeID:  senderID,
				Address: addr,
			},
			SessionID: sessionID,
			Closest:   closest,
			Key:       key,
			Value:     p.GetValue().Value,
		}

		u.fvt.Remove(sessionID)

	case *packet.Packet_FindValue:
		var key Key
		var senderID node.ID
		var sessionID SessionID
		copy(key[:], p.GetFindValue().Key)
		copy(senderID[:], p.GetSessionId())
		copy(sessionID[:], p.GetSessionId())

		u.fvr <- &FindValueRequest{
			key:       key,
			sessionID: sessionID,
			sender: route.Contact{
				NodeID:  senderID,
				Address: addr,
			},
		}

	case *packet.Packet_Ping:
		var sessionID SessionID
		var senderID node.ID
		var challenge []byte
		copy(senderID[:], p.GetSessionId())
		copy(sessionID[:], p.GetSessionId())
		copy(challenge, p.GetPing().GetChallenge())

		u.pr <- &PongRequest{
			From: route.Contact{
				NodeID:  senderID,
				Address: addr,
			},
			SessionID: sessionID,
			Challenge: challenge,
		}

	case *packet.Packet_Pong:
		var sessionID SessionID
		var senderID node.ID
		copy(senderID[:], p.GetSessionId())
		copy(sessionID[:], p.GetSessionId())

		ch := u.pt.Get(sessionID)
		if ch == nil {
			log.Println("handlePackage: Channel not found in table")
			return
		}

		ch <- &PingResult{
			From: route.Contact{
				NodeID:  senderID,
				Address: addr,
			},
			Challenge: p.GetPong().Challenge,
		}

	case *packet.Packet_FindNode:
		var sessionID SessionID
		var senderID node.ID
		copy(sessionID[:], p.GetSessionId())
		copy(senderID[:], p.GetSenderId())

		u.fnr <- &FindNodesRequest{
			SessionID: sessionID,
			sender: route.Contact{
				NodeID:  senderID,
				Address: addr,
			},
		}

	default:
		log.Println("Unhandled packet", p)
	}
}

func generateID() (id SessionID) {
	_, err := rng(id[:])
	if err != nil {
		panic(err)
	}
	return id
}

func generateChallenge() (c []byte) {
	_, err := rng(c)
	if err != nil {
		panic(err)
	}
	return
}

func send(addr *net.UDPAddr, packet packet.Packet) error {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	b, err := proto.Marshal(&packet)
	if err != nil {
		return err
	}
	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}
