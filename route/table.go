package route

import (
	"bytes"
	"container/list"
	"errors"

	"github.com/optmzr/d7024e-dht/node"
)

const bucketSize = 20

type bucket struct{ *list.List }

// Table implements a routing table according to the Kademlia specification.
type Table [node.IDLength]*bucket

// Distance represents the distance between two node IDs.
type Distance [node.IDBytesLength]byte

func (d Distance) BucketIndex() int {
	// Count number of leading zeros.
	for i, b := range d {
		for j := 7; j >= 0; j-- {
			if (b>>uint(j))&1 != 0 {
				return i*8 + (8 - j) - 1
			}
		}
	}

	// If distance is zero, set index to be:
	// 	i = (distance capacity)*8 - 1
	// i.e. the distances 0001 and 0000 have the same prefix (000).
	return cap(d)*8 - 1
}

func (a Distance) Less(b Distance) bool {
	return bytes.Compare(a[:], b[:]) < 0
}

// distance calculates the XOR metric for Kademlia.
func distance(a, b node.ID) (d Distance) {
	for i := range a {
		d[i] = a[i] ^ b[i]
	}
	return
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

// contacts returns all the contacts in a bucket including the distance to a
// provided node ID.
func (b *bucket) contacts(id node.ID) (c Contacts) {
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
	b := rt[d.BucketIndex()]
	b.add(c)
}

// NClosest finds the N closest nodes for a provided node ID.
func (rt *Table) NClosest(target node.ID, n int) (sl *Candidates) {
	me := rt.me()
	d := distance(me.NodeID, target)
	index := d.BucketIndex()

	b := rt[index]
	sl = NewCandidates(b.contacts(me.NodeID)...)

	for i := 1; sl.Len() < n && (index-i >= 0 || index+i < cap(rt)); i++ {
		if index-i >= 0 {
			b = rt[index-i]
			sl.Add(b.contacts(me.NodeID)...)
		}
		if index+i < cap(rt) {
			b = rt[index+i]
			sl.Add(b.contacts(me.NodeID)...)
		}
	}

	if sl.Len() >= n {
		// Create new truncated shortlist with only the N closest nodes.
		sl = NewCandidates(sl.SortedContacts()[:n]...)
	}

	return
}

// NewTable creates a new routing table with all the buckets initialized and the
// local node added to the last bucket. At least one bootstrapping node must be
// provided.
func NewTable(me Contact, others []Contact) (rt *Table, err error) {
	if len(others) == 0 {
		err = errors.New("at least one bootstrap contact must be provided")
		return
	}

	rt = new(Table)

	// Create all the buckets.
	for i := range rt {
		rt[i] = &bucket{list.New()}
	}

	// Add local node to last bucket.
	rt[bucketSize-1].PushFront(me)

	// Add bootstrapping contacts.
	for _, other := range others {
		rt.Add(other)
	}

	return
}
