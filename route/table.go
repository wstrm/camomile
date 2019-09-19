package route

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"math/bits"
	"net"
	"sort"
)

const idLength = 160
const bytesLength = idLength / 8
const bucketSize = 20

type Table [idLength]*Bucket
type Bucket struct{ *list.List }

type NodeID [bytesLength]byte

type Contact struct {
	NodeID   NodeID
	Address  net.UDPAddr
	distance Distance
}

type Candidates []Contact

type Distance uint64

func (cs Candidates) Len() int {
	return len(cs)
}

func (cs Candidates) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs Candidates) Less(i, j int) bool {
	return cs[i].distance < cs[j].distance
}

func (cs Candidates) sort() {
	sort.Sort(cs)
}

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
func (rt *Table) me() Contact {
	lastBucket := rt[bucketSize-1]
	return lastBucket.Front().Value.(Contact)
}

// distance calculates the XOR metric for Kademlia.
func distance(a, b NodeID) Distance {
	d := make([]byte, cap(a))

	for i := range a {
		d[i] = a[i] ^ b[i]
	}

	return Distance(binary.BigEndian.Uint64(d))
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

// candidates returns all the candidates in a bucket including the distance to a
// provided NodeID.
func (bucket *Bucket) candidates(id NodeID) (candidates Candidates) {
	var contact Contact
	for e := bucket.Front(); e != nil; e = e.Next() {
		contact = e.Value.(Contact)
		contact.distance = distance(id, contact.NodeID)
		candidates = append(candidates, contact)
	}
	return
}

// Add finds the correct bucket to add the contact to and inserts the contact.
func (rt *Table) Add(c Contact) {
	me := rt.me()

	d := distance(me.NodeID, c.NodeID)
	b := rt[d.index()]
	b.add(c)

	return
}

func (rt *Table) NClosest(id NodeID, n int) (contacts []Contact) {
	me := rt.me()
	d := distance(me.NodeID, id)
	index := d.index()

	var bucket *Bucket
	var candidates Candidates

	bucket = rt[index]
	candidates = append(candidates, bucket.candidates(me.NodeID)...)

	for i := 1; len(candidates) < n && (index-i >= 0 || index+i < cap(rt)); i++ {
		if index-i >= 0 {
			bucket = rt[index-i]
			candidates = append(candidates, bucket.candidates(me.NodeID)...)
		}
		if index+i < cap(rt) {
			bucket = rt[index+i]
			candidates = append(candidates, bucket.candidates(me.NodeID)...)
		}
	}

	candidates.sort()

	if candidates.Len() < n {
		return candidates
	} else {
		return candidates[:n]
	}
}

// New creates a new routing table with all the buckets initialized and the
// local node added to the last bucket.
func New(me Contact, other Contact) (rt *Table) {
	rt = new(Table)

	for i := range rt {
		rt[i] = &Bucket{list.New()}
	}

	// Add local node to last bucket.
	rt[bucketSize-1].PushFront(me)

	// Add bootstrapping contact.
	rt.Add(other)

	return rt
}

/* TODO(opmtzr)
func (b *Table) remove(id []byte)                               {}
*/
