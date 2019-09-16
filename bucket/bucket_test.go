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
		a   NodeID
		b   NodeID
		exp uint64
	}{
		{
			a:   randomID(),
			b:   randomID(),
			exp: 1025590944692582766,
		},
	}

	for _, test := range testTable {
		d, err := distance(test.a, test.b)
		if err != nil {
			t.Error(err)
		}

		if d != test.exp {
			t.Errorf("unexpected distance for:\n\ta=%x,\n\tb=%x,\ngot: %d, expected: %d\n", test.a, test.b, d, test.exp)
		}
	}
}
