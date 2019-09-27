package route

import (
	"net"
	"testing"

	"github.com/optmzr/d7024e-dht/node"
)

func randomContacts(target node.ID, n int) (contacts []Contact) {
	for i := 0; i < n; i++ {
		contacts = append(contacts, NewContact(target, randomID(), net.UDPAddr{}))
	}
	return
}

func TestNewContact(t *testing.T) {
	target := zeroID()
	id := randomID()
	contact := NewContact(target, id, net.UDPAddr{})
	if !id.Equal(contact.NodeID) {
		t.Errorf("unexpected node ID, got: %v, exp: %v", contact.NodeID, id)
	}

	expDistance := distance(target, id)
	if contact.distance != expDistance {
		t.Errorf("unexpected distance, got: %v, exp: %v", contact.distance, expDistance)
	}
}

func TestNewCandidates(t *testing.T) {
	numContacts := 10
	contacts := randomContacts(zeroID(), numContacts)

	sl := NewCandidates(contacts...)

	slLen := sl.Len()
	if slLen != numContacts {
		t.Errorf("unexpected number of candidates, got: %d, exp: %d", slLen, numContacts)
	}
}

func TestCandidatesSortedContacts(t *testing.T) {
	numContacts := 10
	target := zeroID()
	contacts := randomContacts(target, numContacts)

	sl := NewCandidates(contacts...)
	sorted := sl.SortedContacts()

	sortedLen := len(sorted)
	if sortedLen != numContacts {
		t.Errorf("unexpected number of contacts, got: %d, exp: %d", sortedLen, numContacts)
	}

	for i := 1; i < sortedLen; i++ {
		c := sorted[i]
		l := sorted[i-1]

		if !l.distance.Less(c.distance) {
			t.Errorf("unsorted contacts found in slice: %v > %v", l, c)
		}
	}
}

func TestCandidatesRemove(t *testing.T) {
	numContacts := 10
	contacts := randomContacts(zeroID(), numContacts)

	sl := NewCandidates(contacts...)

	sl.Remove(contacts[4])

	slLen := sl.Len()
	if slLen != numContacts-1 {
		t.Errorf("unexpected number of candidates, got: %d, exp: %d", slLen, numContacts-1)
	}

	sorted := sl.SortedContacts()
	for _, contact := range sorted {
		if contact.NodeID.Equal(contacts[4].NodeID) {
			t.Errorf("contact: %v is still present in the candidates instance", contact)
		}
	}
}

func TestCandidatesRemoveNonExisting(t *testing.T) {
	sl := NewCandidates(randomContacts(zeroID(), 10)...)
	nonExisting := randomID()

	sorted := sl.SortedContacts()
	for _, contact := range sorted {
		if contact.NodeID.Equal(nonExisting) {
			t.Errorf("contact: %v shouldn't exist (256! ~ infinity chance)", contact)
		}
	}

	// Shouldn't panic.
	sl.Remove(NewContact(zeroID(), nonExisting, net.UDPAddr{}))
}
