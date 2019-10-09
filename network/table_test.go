package network

import (
	"testing"
	"time"
)

func TestTable_putGet(t *testing.T) {
	// Create ticker that doesn't remove any element during the lifetime of this
	// test.
	ticker := time.NewTicker(time.Hour)
	table := newTable(time.Hour, ticker)

	id := generateID()
	ch := makeResultChan()

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
		t.Errorf("expected: nil, got: %v", res)
	}
}

func TestTable_remove(t *testing.T) {
	// Create ticker that doesn't remove any element during the lifetime of this
	// test.
	ticker := time.NewTicker(time.Hour)
	table := newTable(time.Hour, ticker)

	id := generateID()
	ch := makeResultChan()

	table.Put(id, ch)

	_, ok := table.Get(id)
	if !ok {
		t.Error("expected channel, got nil")
	}

	table.Remove(id)

	_, ok = table.Get(id)
	if ok {
		t.Error("expected no channel")
	}
}

func TestTable_ttl(t *testing.T) {
	tch := make(chan time.Time)
	ticker := &time.Ticker{
		C: tch,
	}

	go func(tch chan time.Time) {
		// Add an hour to mock "now". This will make the item to expire
		// immediately.
		tch <- time.Now().Add(time.Hour)
	}(tch)

	id := generateID()
	ch := makeResultChan()

	table := newTable(time.Nanosecond, ticker)
	table.Put(id, ch)

	select {
	case v := <-ch: // Wait for removal.
		if v != nil {
			t.Errorf("expected to receive nil value from channel, got: %v", v)
		}
	case <-time.After(1 * time.Second):
		// Let's not wait too long for errors (should be removed instantly).
		t.Error("channel didn't receive null within 1 second")
	}

	_, ok := table.Get(id)
	if ok {
		t.Error("expected channel to be removed")
	}
}
