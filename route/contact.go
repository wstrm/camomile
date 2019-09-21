package dht

import (
	"net"
	"sort"

	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

// Contact contains the node ID and an UDP address.
type Contact struct {
	NodeID   node.ID
	Address  net.UDPAddr
	distance uint64
}

// Contacts implements a sortable list of contacts.
type Contacts []Contact

// Shortlist implements a set of contacts.
type Shortlist struct {
	contacts map[node.ID]route.Contact
}

// Len returns the number of candidates.
func (cs Contacts) Len() int {
	return len(cs)
}

// Swap swaps the i'th and the j'th node.
func (cs Contacts) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

// Less returns true if the distance of the i'th node is less than the j'th
// node.
func (cs Contacts) Less(i, j int) bool {
	return cs[i].distance < cs[j].distance
}

// sort sorts the candidates by their distance to the local node.
func (cs Contacts) sort() {
	sort.Sort(cs)
}

func (sl *Shortlist) Add(contacts ...Contact) {
	for contact := range contacts {
		sl.contacts[contact.NodeID] = contact
	}
}

func (sl *Shortlist) Remove(contact Contact) {
	delete(sl.contacts, contact.NodeID)
}

func (sl *Shortlist) Len() int {
	return len(sl.contacts)
}

// SortedContacts returns all the contacts in the shortlist set sorted by their
// distance.
func (sl *Shortlist) SortedContacts() Contacts {
	var contacts Contacts

	for contact := range sl.contacts {
		contacts = append(contacts, contact)
	}
	contacts.sort()

	return contacts
}

// NewShortlist creates a new shortlist set with the provided contacts.
func NewShortlist(contacts []route.Contact) *Shortlist {
	sl := new(Shortlist)
	for contact := range contacts {
		sl.contacts[contact.NodeID] = contact
	}
}
