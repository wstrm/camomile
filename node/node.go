package node

import (
	"bytes"
	"crypto/rand"
)

const IDLength = 256 // bits.
const IDBytesLength = IDLength / 8

// ID represents a node's ID.
type ID [IDBytesLength]byte

type randRead func([]byte) (int, error)

var rng randRead

func init() {
	rng = rand.Read
}

func NewID() (id ID) {
	buf := make([]byte, IDBytesLength)

	_, err := rng(buf)
	if err != nil {
		panic(err)
	}

	copy(id[:], buf[:IDBytesLength])
	return id
}

// Bytes returns the bytes slice without a fixed size for a node ID.
func (n ID) Bytes() []byte {
	return n[:]
}

// Equal compares the node ID with another.
func (a ID) Equal(b ID) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}
