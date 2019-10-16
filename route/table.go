package route

import (
	"bytes"
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/optmzr/d7024e-dht/node"
)

const BucketSize = 32

type bucket struct {
	*list.List
	lastAccess time.Time
	rw         sync.RWMutex
}

// Table implements a routing table according to the Kademlia specification.
type Table struct {
	buckets   [node.IDLength]*bucket
	me        Contact
	tRefresh  time.Duration
	refreshCh chan int
}

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

// touch updates the last access timestamp to now.
func (b *bucket) touch() {
	b.rw.Lock()
	b.lastAccess = time.Now()
	b.rw.Unlock()
}

// add adds the contact to the bucket, it'll return false if the bucket is full.
func (b *bucket) add(c Contact) (ok bool) {
	b.touch()

	b.rw.Lock()
	defer b.rw.Unlock()

	// Search for the element in case it already exists and move it to the
	// front.
	for e := b.Front(); e != nil; e = e.Next() {
		if c.NodeID.Equal(e.Value.(Contact).NodeID) {
			b.MoveToFront(e)
			// Successfully "added", in reality, the position in the list was
			// just updated.
			return true
		}
	}

	// Make sure the bucket is not larger than the maximum bucket size, k.
	if b.Len() < BucketSize {
		b.PushFront(c) // Add the contact in the front, last seen.
		return true
	}

	return false // Full bucket, contact was not added.
}

// head retrieves the oldest contact in a bucket. The bucket must have at least
// one contact, or else it'll panic.
func (b *bucket) head() Contact {
	b.touch()

	b.rw.RLock()
	defer b.rw.RUnlock()

	e := b.Back()
	if e == nil {
		panic("head must not be called on empty bucket")
	}
	return e.Value.(Contact)
}

// remove a contact from a bucket. If the contact doesn't exist the bucket is
// left unchanged.
func (b *bucket) remove(id node.ID) {
	b.touch()

	b.rw.Lock()
	defer b.rw.Unlock()

	// Small optimization: As the old contacts are usually those that are
	// evicted, iterate through the list backwards to search the oldest contacts
	// first.
	for e := b.Back(); e != nil; e = e.Prev() {
		if id.Equal(e.Value.(Contact).NodeID) {
			b.Remove(e)
		}
	}
}

// contacts returns all the contacts in a bucket including the distance to a
// provided node ID.
func (b *bucket) contacts(id node.ID) (c Contacts) {
	b.touch()

	b.rw.RLock()
	defer b.rw.RUnlock()

	var contact Contact
	for e := b.Front(); e != nil; e = e.Next() {
		contact = e.Value.(Contact)
		contact.distance = distance(id, contact.NodeID)
		c = append(c, contact)
	}
	return
}

func (b *bucket) len() int {
	b.rw.RLock()
	defer b.rw.RUnlock()
	return b.Len()
}

// Add finds the correct bucket to add the contact to and inserts the contact.
// It will return false if the bucket is full.
func (rt *Table) Add(c Contact) (ok bool) {
	me := rt.me

	// Do not add local node to routing table.
	if me.NodeID.Equal(c.NodeID) {
		return true // OK, the node already know of itself.
	}

	d := distance(me.NodeID, c.NodeID)
	b := rt.buckets[d.BucketIndex()]
	return b.add(c)
}

// Head retrieves the oldest contact in a bucket for a specified id.
// The bucket must have at least one contact, or else it'll panic.
func (rt *Table) Head(id node.ID) Contact {
	d := distance(rt.me.NodeID, id)
	b := rt.buckets[d.BucketIndex()]
	return b.head()
}

// Remove a contact from a bucket. If the contact doesn't exist the bucket is
// left unchanged.
func (rt *Table) Remove(id node.ID) {
	d := distance(rt.me.NodeID, id)
	b := rt.buckets[d.BucketIndex()]
	b.remove(id)
}

// Centrality returns the centrality metric according to the formula:
//	Let:
//		Ca = Number of contacts in the bucket corresponding to the target.
//		Cb = Number of contacts in the buckets further away from the target.
//	Then, the centrality, C is:
//		C = Ca + Cb.
func (rt *Table) Centrality(target node.ID) int {
	d := distance(rt.me.NodeID, target)
	index := d.BucketIndex()
	b := rt.buckets[index]

	cb := b.len()

	ca := 0
	for i := 0; i <= index; i++ {
		b = rt.buckets[i]
		ca += b.len()
	}

	return ca + cb
}

// NClosest finds the N closest nodes for a provided node ID.
func (rt *Table) NClosest(target node.ID, n int) (sl *Candidates) {
	me := rt.me
	d := distance(me.NodeID, target)
	index := d.BucketIndex()

	b := rt.buckets[index]
	sl = NewCandidates(target, b.contacts(me.NodeID)...)

	for i := 1; sl.Len() < n && (index-i >= 0 || index+i < cap(rt.buckets)); i++ {
		if index-i >= 0 {
			b = rt.buckets[index-i]
			sl.Add(b.contacts(me.NodeID)...)
		}
		if index+i < cap(rt.buckets) {
			b = rt.buckets[index+i]
			sl.Add(b.contacts(me.NodeID)...)
		}
	}

	if sl.Len() >= n {
		// Create new truncated shortlist with only the N closest nodes.
		sl = NewCandidates(target, sl.SortedContacts()[:n]...)
	}

	return
}

// RefreshCh returns a channel that will be published to when the routing table
// requests a bucket refresh.
func (rt *Table) RefreshCh() chan int {
	return rt.refreshCh
}

// refreshHandler checks for buckets that haven't been touched in tRefresh time
// and sends a refresh request with the bucket index to the refresh channel.
func (rt *Table) refreshHandler(ticker *time.Ticker) {
	go func() {
		for now := range ticker.C {
			for i, b := range rt.buckets {
				b.rw.RLock()
				refresh := now.After(b.lastAccess.Add(rt.tRefresh))
				b.rw.RUnlock()

				if refresh {
					rt.refreshCh <- i
				}
			}
		}
	}()
}

// NewTable creates a new routing table with all the buckets initialized and the
// local node added to the last bucket. At least one bootstrapping node must be
// provided.
func NewTable(me Contact, others []Contact,
	tRefresh time.Duration, refreshTicker *time.Ticker) (rt *Table, err error) {

	if len(others) == 0 {
		err = errors.New("at least one bootstrap contact must be provided")
		return
	}

	rt = new(Table)
	rt.me = me
	rt.refreshCh = make(chan int)
	rt.tRefresh = tRefresh

	// Create all the buckets.
	for i := range rt.buckets {
		rt.buckets[i] = &bucket{List: list.New()}
	}

	// Add bootstrapping contacts.
	for _, other := range others {
		rt.Add(other)
	}

	go rt.refreshHandler(refreshTicker)

	return
}
