package node

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// NewID creates a new cryptographically unique ID.
func NewID() (id ID) {
	buf := make([]byte, IDBytesLength)

	_, err := rng(buf)
	if err != nil {
		panic(err)
	}

	copy(id[:], buf[:IDBytesLength])
	return id
}

// IDFromString parses a hexadecimal representation of an ID into an ID.
func IDFromString(str string) (id ID, err error) {
	i, err := hex.DecodeString(str)
	if err != nil {
		err = fmt.Errorf("cannot decode hex string as ID: %w", err)
		return
	}

	if len(i) != IDBytesLength {
		err = fmt.Errorf("hex string must be %d bytes", IDBytesLength)
		return
	}

	copy(id[:], i)
	return
}

// IDFromBytes reads the bytes in a slice into an ID.
func IDFromBytes(b []byte) (id ID) {
	copy(id[:], b)
	return
}

// Bytes returns the bytes slice without a fixed size for a node ID.
func (n ID) Bytes() []byte {
	return n[:]
}

// Equal compares the node ID with another.
func (a ID) Equal(b ID) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

// String returns the hexadecimal representation of an ID as a string.
func (a ID) String() string {
	return hex.EncodeToString(a[:])
}
