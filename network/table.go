package network

import (
	"sync"
)

type findNodesTable struct {
	items map[PacketID]chan *FindNodesResult
	sync.Mutex
}

func newFindNodesTable() *findNodesTable {
	return &findNodesTable{
		items: make(map[PacketID]chan *FindNodesResult),
		Mutex: sync.Mutex{},
	}
}

func (t *findNodesTable) Put(id PacketID, ch chan *FindNodesResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *findNodesTable) Get(id PacketID) chan *FindNodesResult{
	t.Lock()
	defer t.Unlock()
	return t.items[id]
}

func (t *findNodesTable) Remove(id PacketID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}

type findValueTable struct {
	items map[PacketID]chan *FindValueResult
	sync.Mutex
}

func newFindValueTable() *findValueTable {
	return &findValueTable{
		items: make(map[PacketID]chan  *FindValueResult),
		Mutex: sync.Mutex{},
	}
}

func (t *findValueTable) Put(id PacketID, ch chan *FindValueResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *findValueTable) Get(id PacketID) chan *FindValueResult {
	t.Lock()
	defer t.Unlock()
	return t.items[id]
}

func (t *findValueTable) Remove(id PacketID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}
