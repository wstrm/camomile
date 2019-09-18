package route

import (
	"math/rand" // Not cryptographically secure on purpose.
	"testing"
)

func init() {
	// Make sure the same random ID's are generated on every run.
	rand.Seed(123)
}

func randomID() (id NodeID) {
	buf := make([]byte, bytesLength)

	rand.Read(buf)
	copy(id[:], buf[:bytesLength])

	return
}

func zeroID() (id NodeID) {
	copy(id[:], make([]byte, bytesLength))
	return
}

func makeID(prefix []byte) (id NodeID) {
	l := len(prefix)
	b := bytesLength - l
	copy(id[:l], prefix)
	copy(id[l:], make([]byte, b))
	return
}

func TestDistance(t *testing.T) {
	testTable := []struct {
		a     NodeID
		b     NodeID
		dist  Distance
		index int
	}{
		{
			a:     randomID(),
			b:     randomID(),
			dist:  1025590944692582766,
			index: 4,
		},
		{
			a:     randomID(),
			b:     randomID(),
			dist:  528613502286187044,
			index: 5,
		},
		{
			a:     randomID(),
			b:     randomID(),
			dist:  15488385851010800719,
			index: 0,
		},
	}

	for _, test := range testTable {
		d, err := distance(test.a, test.b)
		if err != nil {
			t.Error(err)
		}

		if d != test.dist {
			t.Errorf("unexpected distance for:\n\ta=%x,\n\tb=%x,\ngot: %d, expected: %d", test.a, test.b, d, test.dist)
		}

		i := d.index()
		if i != test.index {
			t.Errorf("unexpected index for:\n\ta=%x,\n\tb=%x,\ngot: %d, expected: %d", test.a, test.b, i, test.index)
		}
	}
}

func TestMe(t *testing.T) {
	me := Contact{NodeID: randomID()}

	rt := New(me)
	rtMe := rt.me()

	if !me.NodeID.equal(rtMe.NodeID) {
		t.Errorf("inequal node ID, %v != %v", me.NodeID, rtMe.NodeID)
	}
}

func TestAdd(t *testing.T) {
	me := Contact{NodeID: zeroID()}

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

		rt := New(me)

		err := rt.Add(c1)
		if err != nil {
			t.Error(err)
		}

		c2 := rt[j].Front().Value.(Contact)

		if !c1.NodeID.equal(c2.NodeID) {
			t.Errorf("inequal node ID, %v != %v", c1.NodeID, c2.NodeID)
		}

		i++
		j--
	}
}

func TestDuplicateContact(t *testing.T) {
	me := Contact{NodeID: randomID()}
	c1 := Contact{NodeID: randomID()}
	d, _ := distance(me.NodeID, c1.NodeID)

	rt := New(me)

	err := rt.Add(c1)
	if err != nil {
		t.Error(err)
	}

	err = rt.Add(c1)
	if err != nil {
		t.Error(err)
	}

	bucketLen := rt[d.index()].Len()
	expLen := 1
	if bucketLen != expLen {
		t.Errorf("unexpected bucket length, %d != %d", bucketLen, expLen)
	}
}

func BenchmarkAdd(b *testing.B) {
	rt := New(Contact{NodeID: randomID()})

	for n := 0; n < b.N; n++ {
		err := rt.Add(Contact{NodeID: randomID()})
		if err != nil {
			b.Error(err)
		}
	}
}
