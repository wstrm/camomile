package route

import (
	"bytes"
	"math/rand" // Not cryptographically secure on purpose.
	"testing"

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

	rt, _ := NewTable(me, []Contact{boot})
	rtMe := rt.me()

	if !me.NodeID.Equal(rtMe.NodeID) {
		t.Errorf("inequal node ID, %v != %v", me.NodeID, rtMe.NodeID)
	}
}

func TestNewTable(t *testing.T) {
	me := Contact{NodeID: zeroID()}
	boots := []Contact{Contact{NodeID: randomID()}, Contact{NodeID: randomID()}}

	_, err := NewTable(me, boots)
	if err != nil {
		t.Errorf("cannot create table: %w", err)
	}

	_, err = NewTable(me, []Contact{})
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

		rt, _ := NewTable(me, []Contact{boot})

		rt.Add(c1)

		c2 := rt[j].Front().Value.(Contact)

		if !c1.NodeID.Equal(c2.NodeID) {
			t.Errorf("inequal node ID, %v != %v", c1.NodeID, c2.NodeID)
		}

		i++
		j--
	}
}

func TestDuplicateContact(t *testing.T) {
	me := Contact{NodeID: makeID([]byte{1})}
	boot := Contact{NodeID: zeroID()}
	c1 := Contact{NodeID: makeID([]byte{2})}
	d := distance(me.NodeID, c1.NodeID)

	rt, _ := NewTable(me, []Contact{boot})

	rt.Add(c1)
	rt.Add(c1)

	bucketLen := rt[d.BucketIndex()].Len()
	expLen := 1
	if bucketLen != expLen {
		t.Errorf("unexpected bucket length, %d != %d", bucketLen, expLen)
	}
}

func TestNClosest(t *testing.T) {
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot})

	var contacts []Contact
	var contact Contact
	for i := 0; i < 30; i++ {
		contact = Contact{NodeID: randomID()}
		contacts = append(contacts, contact)
		rt.Add(contact)
	}

	closest := rt.NClosest(me.NodeID, 500)
	n := closest.Len()
	if n != 32 { // 1 bootstrap node and 1 local node.
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

	closest = rt.NClosest(me.NodeID, 20)
	n = closest.Len()
	if n != 20 {
		t.Errorf("unexpected number of contacts, got: %d, exp: %d", n, 20)
	}
}

func BenchmarkAdd(b *testing.B) {
	b.StopTimer()
	rt, _ := NewTable(
		Contact{NodeID: randomID()},
		[]Contact{Contact{NodeID: randomID()}})
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		rt.Add(Contact{NodeID: randomID()})
	}
}

func BenchmarkDistance(b *testing.B) {
	b.StopTimer()
	id1 := randomID()
	id2 := randomID()
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		distance(id1, id2)
	}
}

func BenchmarkNClosest(b *testing.B) {
	b.StopTimer()
	me := Contact{NodeID: randomID()}
	boot := Contact{NodeID: randomID()}

	rt, _ := NewTable(me, []Contact{boot})

	var contacts []Contact
	var contact Contact

	contacts = append(contacts, me, boot)

	for i := 0; i < 100; i++ {
		contact = Contact{NodeID: randomID()}
		contacts = append(contacts, contact)
		rt.Add(contact)
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		rt.NClosest(contacts[n%len(contacts)].NodeID, 10)
	}
}
