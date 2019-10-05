package network

import (
	"testing"
)

func TestTableNodes(t *testing.T) {
	table := newTable()

	id := generateID()
	ch := make(chan Result)

	table.Put(id, ch)

	c, ok := table.Get(id)
	if !ok {
		t.Error("expected channel, got nil")
	}

	go func() {
		ch <- nil
	}()

	res := <-c
	if res != nil {
		t.Errorf("Expected: nil Got: %v", res)
	}
}

func TestTablePing(t *testing.T) {
	table := newPingTable()

	id := generateID()
	ch := make(chan *PingResult)

	table.Put(id, ch)

	c, ok := table.Get(id)
	if !ok {
		t.Error("expected channel, got nil")
	}

	go func() {
		ch <- nil
	}()

	res := <-c
	if res != nil {
		t.Errorf("Expected: nil Got: %v", res)
	}
}
