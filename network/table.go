package network

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type item struct {
	result chan interface{}
	ttl    time.Time
}

type table struct {
	items map[SessionID]item
	ttl   time.Duration
	sync.Mutex
}

func makeResultChan() chan interface{} {
	return make(chan interface{})
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

func (t *table) Put(id SessionID, ch chan interface{}) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = item{
		result: ch,
		ttl:    time.Now().Add(t.ttl),
	}
}

func (t *table) Get(id SessionID) (chan interface{}, bool) {
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

func toPingResult(results chan interface{}) chan *PingResult {
	ch := make(chan *PingResult)
	go func() {
		r := <-results
		if r == nil {
			ch <- nil
		} else {
			ch <- r.(*PingResult)
		}
		close(ch)
	}()
	return ch
}

func toFindResult(results chan interface{}) chan FindResult {
	ch := make(chan FindResult)
	go func() {
		r := <-results
		if r == nil {
			ch <- nil
		} else {
			ch <- r.(FindResult)
		}
		close(ch)
	}()
	return ch
}
