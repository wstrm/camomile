package network

import (
	"testing"
)

func TestTableNodes(t *testing.T) {
	table := newFindNodesTable()

	id := generateID()
	ch := make(chan *FindNodesResult)

	table.Put(id, ch)

	c := table.Get(id)

	go func () {
		ch <- nil
	}()

	res := <- c
	if res != nil {
		t.Errorf("Expected: nil Got: %v", res)
	}
}

func TestTableFindValue(t *testing.T) {
	table := newFindValueTable()

	id := generateID()
	ch := make(chan *FindValueResult)

	table.Put(id, ch)

	c := table.Get(id)

	go func () {
		ch <- nil
	}()

	res := <- c
	if res != nil {
		t.Errorf("Expected: nil Got: %v", res)
	}
}
