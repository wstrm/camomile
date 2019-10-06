package network

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type item struct {
	result chan Result
	ttl    time.Time
}

type table struct {
	items map[SessionID]item
	ttl   time.Duration
	sync.Mutex
}

func newTable(ttl time.Duration, ticker *time.Ticker) *table {
	t := &table{
		ttl:   ttl,
		items: make(map[SessionID]item),
	}

	go func() {
		for now := range ticker.C {
			t.Lock()
			for k, v := range t.items {
				if now.After(v.ttl) {
					log.Debug().Msgf("Session timed out (ID: %v)", k)
					v.result <- nil // Signal removal of channel.
					delete(t.items, k)
				}
			}
			t.Unlock()
		}
	}()

	return t
}

func (t *table) Put(id SessionID, ch chan Result) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = item{
		result: ch,
		ttl:    time.Now().Add(t.ttl),
	}
}

func (t *table) Get(id SessionID) (chan Result, bool) {
	t.Lock()
	defer t.Unlock()
	i, ok := t.items[id]
	return i.result, ok
}

func (t *table) Remove(id SessionID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}

type pingTable struct {
	items map[SessionID]chan *PingResult
	sync.Mutex
}

func newPingTable() *pingTable {
	return &pingTable{
		items: make(map[SessionID]chan *PingResult),
	}
}

func (t *pingTable) Put(id SessionID, ch chan *PingResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *pingTable) Get(id SessionID) (chan *PingResult, bool) {
	t.Lock()
	defer t.Unlock()
	ch, ok := t.items[id]
	return ch, ok
}

func (t *pingTable) Remove(id SessionID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}
