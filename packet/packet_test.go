package packet_test

import (
	"bytes"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	"github.com/optmzr/d7024e-dht/packet"
)

func ExamplePacket() {
	r := &packet.Packet{
		PacketId: []byte{123},
	}

	d, err := proto.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}

	rr := &packet.Packet{}
	err = proto.Unmarshal(d, rr)
	if err != nil {
		fmt.Println(err)
	}

	if !bytes.Equal([]byte{123}, rr.PacketId) {
		fmt.Printf("expected '[123]' as PacketId in Request, got '%v'", rr.PacketId)
	} else {
		fmt.Println("got id")
	}

	// Output: got id
}

func ExamplePing() {
	payload := &packet.Ping{
		Challenge: []byte{111},
	}

	r := &packet.Packet{
		PacketId: []byte{123},
		Payload:  &packet.Packet_Ping{payload},
	}

	d, err := proto.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}

	rr := &packet.Packet{}
	err = proto.Unmarshal(d, rr)
	if err != nil {
		fmt.Println(err)
	}

	switch p := rr.Payload.(type) {
	case *packet.Packet_Ping:
		fmt.Printf("got ping: %v", p.Ping.GetChallenge())
	case nil:
		fmt.Printf("expected type '*Packet_Ping' as Request, got '%v'", p)
	}

	// Output: got ping: [111]
}

func ExamplePong() {
	payload := &packet.Pong{
		Challenge: []byte{111},
	}

	r := &packet.Packet{
		PacketId: []byte{123},
		Payload:  &packet.Packet_Pong{payload},
	}

	d, err := proto.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}

	rr := &packet.Packet{}
	err = proto.Unmarshal(d, rr)
	if err != nil {
		fmt.Println(err)
	}

	switch p := rr.Payload.(type) {
	case *packet.Packet_Pong:
		fmt.Printf("got pong: %v", p.Pong.GetChallenge())
	case nil:
		fmt.Printf("expected type '*Packet_Pong' as Request, got '%v'", p)
	}

	// Output: got pong: [111]
}
