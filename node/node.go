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
	_, err := rng(id[:])
	if err != nil {
		panic(err)
	}
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

// RandomIDWithPrefix generates random IDs with the provided ID as a prefix.
//
// The first bit after the prefix is XOR:ed to ensure that each random ID has an
// unique XOR metric distance.
//
// Example:
//	With the following input ID: [11111111 11111111 11111111 11111111 ]
// 	Generates:
//  	* [01010010 00100111 00100000 11010000]
//  	* [10011001 01011100 10111111 10111011]
//  	* [11000000 00111110 00010000 01010111]
//  	* [11100100 10100010 01001100 10010010]
//  	* [11110001 00000010 00100001 00011101]
//  	* [11111010 01110101 11000001 10110101]
//  	* [11111101 11101000 10111111 01000101]
//  	* [11111110 11001111 11101010 00111110]
//  	* [11111111 00100010 00101000 11001010]
//  	* [11111111 10110111 00000110 10110001]
//  	* [11111111 11010000 00010100 11100110]
//  	* [11111111 11100000 11110010 00000011]
//  	* [11111111 11110101 00011111 11011101]
//  	* [11111111 11111010 00100000 01010011]
//  	* [11111111 11111101 01110110 01000110]
//  	* [11111111 11111110 10111100 01011101]
//  	* [11111111 11111111 00101100 10101111]
//  	* [11111111 11111111 10001100 00011110]
//  	* [11111111 11111111 11000111 00001111]
//  	* [11111111 11111111 11100111 11000000]
//  	* [11111111 11111111 11110001 00101010]
//		* and so on...
func RandomIDWithPrefix(id ID) chan ID {
	ch := make(chan ID)

	go func(ch chan ID) {
		for i := 1; i < IDLength; i++ {
			j := (i - 1) / 8 // Byte position.

			// Create new instance of ID every iteration.
			nextID := NewID()

			// Copy prefix of input ID into the new ID.
			copy(nextID[:j], id[:j])

			lastBitPos := uint(i - j*8)
			prefixMask := uint(0xff<<(8-lastBitPos)) & 0xff
			suffixMask := ^prefixMask & 0xff

			// Extract left-most bits and add to n (prefix).
			n := uint(id[j]) & prefixMask

			// XOR the first bit after the prefix to ensure that each random ID
			// has an unique XOR metric distance.
			n ^= 1 << (8 - lastBitPos)

			// Extract right-most bits and add to n (suffix).
			n |= uint(nextID[j]) & suffixMask

			nextID[j] = byte(n)

			ch <- nextID
		}

		close(ch)
	}(ch)

	return ch
}

// IDFromBytes reads the bytes in a slice into an ID.
func IDFromBytes(b []byte) (id ID) {
	copy(id[:], b)
	return
}

// Bytes returns a copy of the bytes slice without a fixed size for a node ID.
func (n ID) Bytes() []byte {
	b := make([]byte, IDBytesLength)
	copy(b, n[:])
	return b
}

// Equal compares the node ID with another.
func (a ID) Equal(b ID) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

// String returns the hexadecimal representation of an ID as a string.
func (a ID) String() string {
	return hex.EncodeToString(a[:])
}
