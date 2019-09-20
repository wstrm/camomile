package route

import (
	"container/list"
	"encoding/binary"
	"math/bits"
	"net"
	"sort"

	"github.com/optmzr/d7024e-dht/node"
)

const bucketSize = 20

type bucket struct{ *list.List }
type candidates []Contact

// Table implements a routing table according to the Kademlia specification.
type Table [node.IDLength]*bucket

// Contact contains the node ID and an UDP address.
type Contact struct {
	NodeID   node.ID
	Address  net.UDPAddr
	distance uint64
}

// leadingZeros counts the number of leading bits that are zero in an uint64.
func leadingZeros(distance uint64) int {
	return bits.LeadingZeros64(uint64(distance))
}

// distance calculates the XOR metric for Kademlia.
func distance(a, b node.ID) uint64 {
	d := make([]byte, cap(a))

	for i := range a {
		d[i] = a[i] ^ b[i]
	}

	return binary.BigEndian.Uint64(d)
}

// Len returns the number of candidates.
func (cs candidates) Len() int {
	return len(cs)
}

// Swap swaps the i'th and the j'th node.
func (cs candidates) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

// Less returns true if the distance of the i'th node is less than the j'th
// node.
func (cs candidates) Less(i, j int) bool {
	return cs[i].distance < cs[j].distance
}

// sort sorts the candidates by their distance to the local node.
func (cs candidates) sort() {
	sort.Sort(cs)
}

// me returns the contact in the last bucket (the local node).
func (rt *Table) me() Contact {
	lastBucket := rt[bucketSize-1]
	return lastBucket.Front().Value.(Contact)
}

// add adds the contact to the bucket.
func (b *bucket) add(c Contact) {
	// Search for the element in case it already exists and move it to the
	// front.
	for e := b.Front(); e != nil; e = e.Next() {
		if c.NodeID.Equal(e.Value.(Contact).NodeID) {
			b.MoveToFront(e)
			return
		}
	}

	// Make sure the bucket is not larger than the maximum bucket size, k.
	if b.Len() < bucketSize {
		b.PushFront(c) // Add the contact in the front, last seen.
	}
}

// candidates returns all the candidates in a bucket including the distance to a
// provided node ID.
func (b *bucket) candidates(id node.ID) (c candidates) {
	var contact Contact
	for e := b.Front(); e != nil; e = e.Next() {
		contact = e.Value.(Contact)
		contact.distance = distance(id, contact.NodeID)
		c = append(c, contact)
	}
	return
}

// Add finds the correct bucket to add the contact to and inserts the contact.
func (rt *Table) Add(c Contact) {
	me := rt.me()

	d := distance(me.NodeID, c.NodeID)
	b := rt[leadingZeros(d)]
	b.add(c)
}

// NClosest finds the N closest nodes for a provided node ID.
func (rt *Table) NClosest(id node.ID, n int) (contacts []Contact) {
	me := rt.me()
	d := distance(me.NodeID, id)
	index := leadingZeros(d)

	var b *bucket
	var c candidates

	b = rt[index]
	c = append(c, b.candidates(me.NodeID)...)

	for i := 1; c.Len() < n && (index-i >= 0 || index+i < cap(rt)); i++ {
		if index-i >= 0 {
			b = rt[index-i]
			c = append(c, b.candidates(me.NodeID)...)
		}
		if index+i < cap(rt) {
			b = rt[index+i]
			c = append(c, b.candidates(me.NodeID)...)
		}
	}

	c.sort()

	if c.Len() < n {
		return c
	} else {
		return c[:n]
	}
}

// New creates a new routing table with all the buckets initialized and the
// local node added to the last bucket.
func New(me Contact, other Contact) (rt *Table) {
	rt = new(Table)

	for i := range rt {
		rt[i] = &bucket{list.New()}
	}

	// Add local node to last bucket.
	rt[bucketSize-1].PushFront(me)

	// Add bootstrapping contact.
	rt.Add(other)

	return rt
}
