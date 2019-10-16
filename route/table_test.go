package route

import (
	"bytes"
	"fmt"
	"math/rand" // Not cryptographically secure on purpose.
	"sync"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/node"
)

func init() {
	// Make sure the same random ID's are generated on every run.
	rand.Seed(123)
}

func randomID() (id node.ID) {
	buf := make([]byte, cap(id))

	rand.Read(buf)
	copy(id[:], buf[:cap(id)])

	return
}

func zeroID() (id node.ID) {
	copy(id[:], make([]byte, cap(id)))
	return
}

func makeID(prefix []byte) (id node.ID) {
	l := len(prefix)
	b := cap(id) - l
	copy(id[:l], prefix)
	copy(id[l:], make([]byte, b))
	return
}

func TestDistance(t *testing.T) {
	testTable := []struct {
		a     node.ID
		b     node.ID
		dist  Distance
		index int
	}{
		{
			a:     makeID([]byte{1}),
			b:     makeID([]byte{1}),
			dist:  Distance{0},
			index: 255,
		},
		{
			a:     makeID([]byte{1}),
			b:     makeID([]byte{2}),
			dist:  Distance{3},
			index: 6,
		},
		{
			a:     makeID([]byte{1}),
			b:     makeID([]byte{5}),
			dist:  Distance{4},
			index: 5,
		},
	}

	for _, test := range testTable {
		d := distance(test.a, test.b)

		if !bytes.Equal(d[:], test.dist[:]) {
			t.Errorf("unexpected distance for:\n\ta=%x,\n\tb=%x,\ngot: %d, exp: %d", test.a, test.b, d, test.dist)
		}

		i := d.BucketIndex()
		if i != test.index {
			t.Errorf("unexpected index for:\n\ta=%x,\n\tb=%x,\ngot: %d, exp: %d", test.a, test.b, i, test.index)
		}
	}
}

func TestMe(t *testing.T) {
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))
	rtMe := rt.me

	if !me.NodeID.Equal(rtMe.NodeID) {
		t.Errorf("inequal node ID, %v != %v", me.NodeID, rtMe.NodeID)
	}
}

func TestNewTable(t *testing.T) {
	me := Contact{NodeID: zeroID()}
	boots := []Contact{Contact{NodeID: randomID()}, Contact{NodeID: randomID()}}

	_, err := NewTable(me, boots,
		time.Second, time.NewTicker(time.Second))
	if err != nil {
		t.Errorf("cannot create table: %w", err)
	}

	_, err = NewTable(me, []Contact{},
		time.Second, time.NewTicker(time.Second))
	if err == nil {
		t.Error("expected error on empty bootstrap slice")
	}
}

func TestAdd(t *testing.T) {
	me := Contact{NodeID: zeroID()}
	boot := Contact{NodeID: randomID()}

	// Will create the IDs:
	// [ 00000001 000000... ]
	// [ 00000010 000000... ]
	// [ 00000100 000000... ]
	// ...
	// And check that the prefix lengths equal:
	// 7, 6, 5...
	i := uint(0)
	j := 7
	for i < 7 {
		c1 := Contact{NodeID: makeID([]byte{1 << i})}

		rt, _ := NewTable(me, []Contact{boot},
			time.Second, time.NewTicker(time.Second))

		rt.Add(c1)

		c2 := rt.buckets[j].Front().Value.(Contact)

		if !c1.NodeID.Equal(c2.NodeID) {
			t.Errorf("inequal node ID, %v != %v", c1.NodeID, c2.NodeID)
		}

		i++
		j--
	}
}

func TestHead_incremental(t *testing.T) {
	me := Contact{NodeID: zeroID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	for i := 2; i < 50; i++ {
		rt.Add(Contact{NodeID: randomID()})

		contact := rt.Head(boot.NodeID)

		if !contact.NodeID.Equal(boot.NodeID) {
			t.Errorf("bootstrap node was not returned as head (load factor: %d), got: %v, exp: %v",
				i, contact.NodeID, boot.NodeID)
		}
	}
}

func TestHead_panic(t *testing.T) {
	me := Contact{NodeID: zeroID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	i := distance(me.NodeID, boot.NodeID).BucketIndex()

	// Produces an index that if offset by 1 from i.
	c1 := Contact{NodeID: makeID([]byte{1 << uint(i)})}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic as %v shouldn't be in the same bucket as %v", boot.NodeID, c1.NodeID)
		}
	}()

	rt.Head(c1.NodeID) // Unit under test.
}

func TestRemove_incremental(t *testing.T) {
	me := Contact{NodeID: zeroID()}

	var others []Contact
	for i := 0; i < 100; i++ {
		others = append(others, Contact{NodeID: randomID()})
	}

	rt, _ := NewTable(me, others,
		time.Second, time.NewTicker(time.Second))

	// Shuffle the contacts so that they are removed in a random order.
	rand.Shuffle(len(others),
		func(i, j int) { others[i], others[j] = others[j], others[i] })

	for i := 0; i < len(others); i++ {
		a := others[i]

		rt.Remove(a.NodeID)

		contacts := rt.NClosest(me.NodeID, 500).SortedContacts()
		for _, b := range contacts {
			if b.NodeID.Equal(a.NodeID) {
				t.Errorf("node %v still exist in the routing table", a.NodeID)
			}
		}
	}
}

func TestDuplicateContact(t *testing.T) {
	me := Contact{NodeID: makeID([]byte{1})}
	boot := Contact{NodeID: zeroID()}
	c1 := Contact{NodeID: makeID([]byte{2})}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	rt.Add(c1)
	rt.Add(c1)

	numContacts := 0
	for _, bucket := range rt.buckets {
		numContacts += bucket.Len()
	}

	expLen := 2 // A bootstrap node and c1 (and no duplicate of c1).
	if numContacts != expLen {
		t.Errorf("unexpected number of contacts, %d != %d", numContacts, expLen)
	}
}
func TestAddLocalNode(t *testing.T) {
	me := Contact{NodeID: makeID([]byte{1})}
	boot := Contact{NodeID: zeroID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	rt.Add(me)

	numContacts := 0
	for _, bucket := range rt.buckets {
		numContacts += bucket.Len()
	}

	expLen := 1 // A single bootstrap node.
	if numContacts != expLen {
		t.Errorf("unexpected number of contacts, %d != %d", numContacts, expLen)
	}
}

func TestNClosest(t *testing.T) {
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	var contacts []Contact
	var contact Contact
	for i := 0; i < 30; i++ {
		contact = Contact{NodeID: randomID()}
		contacts = append(contacts, contact)
		rt.Add(contact)
	}

	closest := rt.NClosest(me.NodeID, 500)
	n := closest.Len()
	if n != 31 { // 1 bootstrap node.
		t.Errorf("unexpected number of contacts, got: %d, exp: %d", n, 32)
	}

	for _, contact := range contacts {
		found := false
		for _, c := range closest.SortedContacts() {
			if contact.NodeID.Equal(c.NodeID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("contact with node ID: %v doesn't exist", contact.NodeID)
		}
	}

	// Fetch 20 closest contacts for local node.
	closest = rt.NClosest(me.NodeID, 20)
	n = closest.Len()
	if n != 20 {
		t.Errorf("unexpected number of contacts, got: %d, exp: %d", n, 20)
	}

	// Fetch 30 closest contacts for bootstrap node.
	closest = rt.NClosest(boot.NodeID, 30)
	n = closest.Len()
	if n != 30 {
		t.Errorf("unexpected number of contacts, got: %d, exp: %d", n, 30)
	}
}

func BenchmarkAdd(b *testing.B) {
	rt, _ := NewTable(
		Contact{NodeID: randomID()},
		[]Contact{Contact{NodeID: randomID()}},
		time.Second, time.NewTicker(time.Second))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rt.Add(Contact{NodeID: randomID()})
	}
}

func BenchmarkDistance(b *testing.B) {
	id1 := randomID()
	id2 := randomID()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		distance(id1, id2)
	}
}

func BenchmarkNClosest(b *testing.B) {
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	var contacts []Contact
	var contact Contact

	contacts = append(contacts, me, boot)

	for i := 0; i < 100; i++ {
		contact = Contact{NodeID: randomID()}
		contacts = append(contacts, contact)
		rt.Add(contact)
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rt.NClosest(contacts[n%len(contacts)].NodeID, 10)
	}
}

func TestNClosest_concurrent(t *testing.T) {
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot},
		time.Second, time.NewTicker(time.Second))

	var contacts []Contact
	var wg sync.WaitGroup

	contacts = append(contacts, me, boot)

	fmt.Print("Add call order: ")
	for i := 0; i < 100; i++ {
		contact := Contact{NodeID: randomID()}
		contacts = append(contacts, contact)

		wg.Add(1)

		go func(i int) {
			time.Sleep(time.Duration(rand.Int()%2) * time.Millisecond)
			rt.Add(contact)
			fmt.Printf("%d ", i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println()

	fmt.Print("NClosest call order: ")
	for n := 0; n < 100; n++ {
		wg.Add(1)
		go func(n int) {
			time.Sleep(time.Duration(rand.Int()%2) * time.Millisecond)
			rt.NClosest(contacts[n%len(contacts)].NodeID, 10)
			fmt.Printf("%d ", n)
			wg.Done()
		}(n)
	}
	wg.Wait()
	fmt.Println()
}

func TestRefreshCh(t *testing.T) {
	me := Contact{NodeID: makeID([]byte{0xff})}
	boot := Contact{NodeID: makeID([]byte{0x7f})}
	tExpire := time.Second

	tch := make(chan time.Time)
	ticker := &time.Ticker{
		C: tch,
	}

	go func(tch chan time.Time) {
		// Add an hour to mock "now". This will make the channel to fire a
		// refresh event immediately.
		tch <- time.Now().Add(time.Hour)
	}(tch)

	rt, _ := NewTable(me, []Contact{boot}, tExpire, ticker)

	// Every bucket should be untouched on initialization, producing a refresh
	// event for all of them (in order).
	exp := 0
	for {
		if exp == 256 {
			break
		}

		i := <-rt.RefreshCh()
		if i != exp {
			t.Errorf("unexpected bucket index from refresh channel, got: %d, exp: %d", i, exp)
		}

		exp++
	}
}

func TestCentrality(t *testing.T) {
	me := Contact{NodeID: zeroID()}

	var others []Contact
	for i := 0; i < 254; i++ {
		others = append(others, Contact{NodeID: makeID([]byte{byte(i)})})
	}

	rt, _ := NewTable(me, others,
		time.Second, time.NewTicker(time.Second))

	c := rt.Centrality(zeroID())
	exp := 127
	if c != exp {
		t.Errorf("unexpected centrality, got: %d, exp: %d", c, exp)
	}

	c = rt.Centrality(makeID([]byte{byte(254)}))
	exp = 64
	if c != exp {
		t.Errorf("unexpected centrality, got: %d, exp: %d", c, exp)
	}
}
