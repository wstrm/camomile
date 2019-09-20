package node

import (
	"bytes"
	"crypto/rand"
)

const IDLength = 256 // bits.
const IDBytesLength = IDLength / 8

// ID represents a node's ID.
type ID [IDBytesLength]byte

func NewID() (id ID) {
	buf := make([]byte, IDBytesLength)
	rand.Read(buf)
	copy(id[:], buf[:IDBytesLength])
	return id
}

// bytes returns the bytes slice without a fixed size for a node ID.
func (n ID) bytes() []byte {
	return n[:]
}

// Equal compares the node ID with another.
func (a ID) Equal(b ID) bool {
	return bytes.Equal(a.bytes(), b.bytes())
}
