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

const UdpPort  = ":8118"

type randRead func([]byte) (int, error)

var rng randRead

func init() {
	rng = rand.Read
}

type Key node.ID
type PacketID [node.IDBytesLength]byte

type Network interface {
	FindNodes(target node.ID, addr *net.UDPAddr) (chan *FindNodesResult, error)
	Store(key Key, value string, addr *net.UDPAddr) error
	FindValue(key Key, addr *net.UDPAddr) (chan *FindValueResult, error)
	SendValue(key Key, value string, packetID PacketID, addr *net.UDPAddr) error
}

type udpNetwork struct {
	me node.ID
	fnt *findNodesTable
	fvt *findValueTable
}

type FindNodesResult struct {
	From route.Contact
	Closest []route.Contact
}

type FindValueResult struct {
	From route.Contact
	PacketID PacketID
	Key Key
	Value string
}

func NewUDPNetwork(id node.ID) Network {
	n := &udpNetwork{me:id, fvt: newFindValueTable(), fnt: newFindNodesTable()}
	go n.listen()
	return n
}

func (u *udpNetwork) FindNodes(target node.ID, addr *net.UDPAddr) (chan *FindNodesResult, error) {
	panic("implement me")
}

func (u *udpNetwork) Store(key Key, value string, addr *net.UDPAddr) error {
	id := generateID()

	payload := &packet.Store{
		Key:                  key[:],
		Value:                value,
	}
	p := &packet.Packet{
		PacketId:             id[:],
		SenderId:             u.me.Bytes(),
		Payload:              &packet.Packet_Store{Store: payload},
	}

	return send(addr, *p)
}

func (u *udpNetwork) FindValue(key Key, addr *net.UDPAddr) (chan *FindValueResult, error) {
	id := generateID()

	payload := &packet.FindValue{
		Key:                  key[:],
	}
	p := &packet.Packet{
		PacketId:             id[:],
		SenderId:             u.me.Bytes(),
		Payload:              &packet.Packet_FindValue{FindValue: payload},
	}

	err := send(addr, *p)
	if err != nil {
		return nil, err
	}

	results := make(chan *FindValueResult)
	u.fvt.Put(id, results)

	return results, nil
}

func (u *udpNetwork) SendValue(key Key, value string, packetID PacketID, addr *net.UDPAddr) error {
	payload := &packet.Value{
		Key:                  key[:],
		Value:                value,
	}
	p := &packet.Packet{
		PacketId:             packetID[:],
		SenderId:             u.me.Bytes(),
		Payload:              &packet.Packet_Value{Value: payload},
	}

	err := send(addr, *p)
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

	for {
		log.Printf("Listening for UDP packets on port %v", UdpPort)

		data := make([]byte, 1500)
		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Fatalf("Error when reading from UDP from address %v: %s", addr.String(), err)
		}
		b := data[:n]
		go u.handlePacket(b, addr)
	}
}

func (u *udpNetwork) handlePacket(b []byte, addr *net.UDPAddr) {
	p := &packet.Packet{}
	err := proto.Unmarshal(b, p)
	if err != nil {
		log.Println("Error unserializing packet", err)
		return
	}

	switch p.Payload.(type) {
	case *packet.Packet_Value:

		var packetID PacketID
		var senderID node.ID
		var key Key
		copy(packetID[:], p.PacketId)
		copy(senderID[:], p.SenderId)
		copy(key[:], p.GetValue().Key)

		ch := u.fvt.Get(packetID)
		if ch == nil {
			log.Println("Channel not found in table")
			return
		}

		ch <- &FindValueResult{
			From:  route.Contact{
				NodeID:  senderID,
				Address: *addr,
			},
			PacketID: packetID,
			Key:   key,
			Value: p.GetValue().Value,
		}

		u.fvt.Remove(packetID)

	default:
		log.Println("Unhandled packet", p)
	}
}

func generateID() (id PacketID) {
	_, err := rng(id[:])
	if err != nil {
		log.Fatalln("Error generating PacketID", err)
	}
	log.Println(id)
	return id
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

