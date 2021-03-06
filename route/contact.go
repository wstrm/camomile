package route

import (
	"net"
	"sort"

	"github.com/optmzr/d7024e-dht/node"
)

// Contact contains the node ID and an UDP address.
type Contact struct {
	NodeID   node.ID
	Address  net.UDPAddr
	distance Distance
}

// Contacts implements a sortable list of contacts.
type Contacts []Contact

type contactMap map[node.ID]Contact

// Candidates implements a set of contacts.
type Candidates struct {
	target   node.ID
	contacts contactMap
}

func NewContact(id node.ID, address net.UDPAddr) Contact {
	return Contact{
		NodeID:  id,
		Address: address,
	}
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
	return cs[i].distance.Less(cs[j].distance)
}

// sort sorts the candidates by their distance to the local node.
func (cs Contacts) sort() {
	sort.Sort(cs)
}

func (sl *Candidates) Add(contacts ...Contact) {
	for _, contact := range contacts {
		sl.contacts[contact.NodeID] = contact
	}
}

func (sl *Candidates) Remove(contact Contact) {
	delete(sl.contacts, contact.NodeID)
}

func (sl *Candidates) Len() int {
	return len(sl.contacts)
}

// SortedContacts returns all the contacts in the shortlist set sorted by their
// distance.
func (sl *Candidates) SortedContacts() Contacts {
	var contacts Contacts

	for _, contact := range sl.contacts {
		contact.distance = distance(sl.target, contact.NodeID)
		contacts = append(contacts, contact)
	}
	contacts.sort()

	return contacts
}

// NewCandidates creates a new shortlist set with the provided contacts.
func NewCandidates(target node.ID, contacts ...Contact) *Candidates {
	sl := new(Candidates)
	sl.target = target
	sl.contacts = make(contactMap)

	for _, contact := range contacts {
		sl.contacts[contact.NodeID] = contact
	}

	return sl
}
