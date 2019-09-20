package node

import "testing"

func TestEqual(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	if id1.Equal(id2) {
		t.Error("two instances of ids must not equal")
	}

	if !id1.Equal(id1) {
		t.Error("same instance of id must equal itself")
	}
}
