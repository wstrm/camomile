package route

import (
	"net"
	"testing"
)

func randomContacts(n int) (contacts []Contact) {
	for i := 0; i < n; i++ {
		contacts = append(contacts, NewContact(randomID(), net.UDPAddr{}))
	}
	return
}

func TestNewContact(t *testing.T) {
	id := randomID()
	contact := NewContact(id, net.UDPAddr{})
	if !id.Equal(contact.NodeID) {
		t.Errorf("unexpected node ID, got: %v, exp: %v", contact.NodeID, id)
	}
}

func TestNewCandidates(t *testing.T) {
	numContacts := 10
	contacts := randomContacts(numContacts)

	sl := NewCandidates(zeroID(), contacts...)

	slLen := sl.Len()
	if slLen != numContacts {
		t.Errorf("unexpected number of candidates, got: %d, exp: %d", slLen, numContacts)
	}
}

func TestCandidatesSortedContacts(t *testing.T) {
	numContacts := 10
	target := zeroID()
	contacts := randomContacts(numContacts)

	sl := NewCandidates(target, contacts...)
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
	contacts := randomContacts(numContacts)

	sl := NewCandidates(zeroID(), contacts...)

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
	sl := NewCandidates(zeroID(), randomContacts(10)...)
	nonExisting := randomID()

	sorted := sl.SortedContacts()
	for _, contact := range sorted {
		if contact.NodeID.Equal(nonExisting) {
			t.Errorf("contact: %v shouldn't exist (256! ~ infinity chance)", contact)
		}
	}

	// Shouldn't panic.
	sl.Remove(NewContact(nonExisting, net.UDPAddr{}))
}
