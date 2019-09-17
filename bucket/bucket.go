package bucket

import (
	"encoding/binary"
	"fmt"
	"net"
)

const idLength = 160
const bytesLength = idLength / 8
const bucketSize = 20

type Buckets [idLength]Bucket
type Bucket [bucketSize]Contact

type NodeID [bytesLength]byte

type Contact struct {
	NodeID  NodeID
	Address net.UDPAddr
	Port    uint32
}

func distance(a, b NodeID) (uint64, error) {
	x := len(a)
	y := len(b)

	if x != y {
		return 0, fmt.Errorf("inequal byte slice lengths: %d != %d", x, y)
	}

	d := make([]byte, bytesLength)

	for i := range a {
		d[i] = a[i] ^ b[i]
	}

	return binary.BigEndian.Uint64(d), nil
}

/* TODO(opmtzr)
func (b *Buckets) Add(c Contact) (err error)                      {}
func (b *Buckets) Closest(id []byte) (contact Contact, err error) {}
func (b *Buckets) remove(id []byte)                               {}
*/
