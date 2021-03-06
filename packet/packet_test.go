package packet_test

import (
	"bytes"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	"github.com/optmzr/d7024e-dht/packet"
)

func ExamplePacket() {
	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
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

	if !bytes.Equal([]byte{123}, rr.GetSessionId()) {
		fmt.Printf("expected '[123]' as SessionId in payload, got '%v'", rr.SessionId)
	} else if !bytes.Equal([]byte{100}, rr.GetSenderId()) {
		fmt.Printf("expected '[100]' as SenderId in payload, got '%v'", rr.SenderId)
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
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_Ping{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_Ping:
		fmt.Printf("got ping: %v", rr.GetPing().GetChallenge())
	case nil:
		fmt.Printf("expected type '*Packet_Ping' as payload, got '%v'", p)
	}

	// Output: got ping: [111]
}

func ExamplePong() {
	payload := &packet.Pong{
		Challenge: []byte{111},
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_Pong{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_Pong:
		fmt.Printf("got pong: %v", rr.GetPong().GetChallenge())
	case nil:
		fmt.Printf("expected type '*Packet_Pong' as payload, got '%v'", p)
	}

	// Output: got pong: [111]
}

func ExampleStore() {
	payload := &packet.Store{
		Key:   []byte{111},
		Value: "ABC, du är mina tankar",
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_Store{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_Store:
		fmt.Printf("got store: %v=%v", rr.GetStore().GetKey(), p.Store.GetValue())
	case nil:
		fmt.Printf("expected type '*Packet_Store' as payload, got '%v'", p)
	}

	// Output: got store: [111]=ABC, du är mina tankar
}

func ExampleValue() {
	payload := &packet.Value{
		Key:   []byte{111},
		Value: "ABC, du är mina tankar",
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_Value{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_Value:
		fmt.Printf("got value: %v=%v", rr.GetValue().GetKey(), rr.GetValue().GetValue())
	case nil:
		fmt.Printf("expected type '*Packet_Value' as payload, got '%v'", p)
	}

	// Output: got value: [111]=ABC, du är mina tankar
}

func ExampleFindValue() {
	payload := &packet.FindValue{
		Key: []byte{111},
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_FindValue{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_FindValue:
		fmt.Printf("got find_value: %v", rr.GetFindValue().GetKey())
	case nil:
		fmt.Printf("expected type '*Packet_FindValue' as payload, got '%v'", p)
	}

	// Output: got find_value: [111]
}

func ExampleFindNode() {
	payload := &packet.FindNode{
		NodeId: []byte{111},
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_FindNode{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_FindNode:
		fmt.Printf("got find_node: %v", rr.GetFindNode().GetNodeId())
	case nil:
		fmt.Printf("expected type '*Packet_FindNode' as payload, got '%v'", p)
	}

	// Output: got find_node: [111]
}

func ExampleNodeList() {
	nodes := []*packet.NodeInfo{
		&packet.NodeInfo{
			NodeId: []byte{111},
			Ip:     []byte{1, 1, 1, 1},
			Port:   1337,
		},
		&packet.NodeInfo{
			NodeId: []byte{112},
			Ip:     []byte{1, 1, 1, 2},
			Port:   1338,
		},
	}
	payload := &packet.NodeList{
		Nodes: nodes,
	}

	r := &packet.Packet{
		SessionId: []byte{123},
		SenderId:  []byte{100},
		Payload:   &packet.Packet_NodeList{payload},
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

	switch p := rr.GetPayload().(type) {
	case *packet.Packet_NodeList:
		n := rr.GetNodeList().GetNodes()
		if fmt.Sprintf("%v", n) == fmt.Sprintf("%v", nodes) {
			fmt.Printf("got node_list")
		} else {
			fmt.Printf("expected node_list: %v, got: %v", nodes, n)
		}
	case nil:
		fmt.Printf("expected type '*Packet_NodeList' as payload, got '%v'", p)
	}

	// Output: got node_list
}
