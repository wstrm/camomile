package bucket

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

	// Not cryptographically secure, so that the same ID's are generated on every run.
	rand.Read(buf)
	copy(id[:], buf[:bytesLength])

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
