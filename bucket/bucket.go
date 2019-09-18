package bucket

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"math/bits"
	"net"
)

const idLength = 160
const bytesLength = idLength / 8
const bucketSize = 20

type Buckets [idLength]*Bucket
type Bucket struct{ list.List }

type NodeID [bytesLength]byte

type Contact struct {
	NodeID  NodeID
	Address net.UDPAddr
	Port    uint32
}

type Distance uint64

// index counts the number of leading bits that are zero in an uint64.
func (d Distance) index() int {
	return bits.LeadingZeros64(uint64(d))
}

// bytes returns the bytes slice without a fixed size for a NodeID.
func (n NodeID) bytes() []byte {
	return n[:]
}

// equal compares the NodeID with another.
func (a NodeID) equal(b NodeID) bool {
	return bytes.Equal(a.bytes(), b.bytes())
}

// me returns the contact in the last bucket (the local node).
func (buckets *Buckets) me() Contact {
	lastBucket := buckets[bucketSize-1]
	return lastBucket.Front().Value.(Contact)
}

// distance calculates the XOR metric for Kademlia.
func distance(a, b NodeID) (Distance, error) {
	d := make([]byte, cap(a))

	for i := range a {
		d[i] = a[i] ^ b[i]
	}

	return Distance(binary.BigEndian.Uint64(d)), nil
}

// add adds the contact to the bucket.
func (bucket *Bucket) add(c Contact) {
	// Search for the element in case it already exists and move it to the
	// front.
	for e := bucket.Front(); e != nil; e = e.Next() {
		if c.NodeID.equal(e.Value.(Contact).NodeID) {
			bucket.MoveToFront(e)
			return
		}
	}

	// Make sure the bucket is not larger than the maximum bucket size, k.
	if bucket.Len() < bucketSize {
		bucket.PushFront(c) // Add the contact in the front, last seen.
	}
}

// Add finds the correct bucket to add the contact to and inserts the contact.
func (buckets *Buckets) Add(c Contact) (err error) {
	me := buckets.me()

	d, err := distance(me.NodeID, c.NodeID)
	b := buckets[d.index()]
	b.add(c)

	return
}

/* TODO(opmtzr)
func (b *Buckets) Closest(id []byte) (contact Contact, err error) {}
func (b *Buckets) remove(id []byte)                               {}
*/
