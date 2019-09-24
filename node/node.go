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

// NewID returns a cryptocraphically secure random ID.
func NewID() (id ID) {
	buf := make([]byte, IDBytesLength)

	_, err := rng(buf)
	if err != nil {
		panic(err)
	}

	copy(id[:], buf[:IDBytesLength])
	return id
}

// IDFromString takes a hexadecimal string and converts it into an ID.
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

// bytes returns the bytes slice without a fixed size for a node ID.
func (n ID) bytes() []byte {
	return n[:]
}

// Equal compares the node ID with another.
func (a ID) Equal(b ID) bool {
	return bytes.Equal(a.bytes(), b.bytes())
}

// String returns the hexadecimal representation of an ID as a string.
func (a ID) String() string {
	return hex.EncodeToString(a[:])
}
