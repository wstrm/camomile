package node

import (
	"errors"
	"testing"
)

func TestEqual(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	if id1.Equal(id2) {
		t.Error("two instances of ids must not equal")
	}

	if !id1.Equal(id1) {
		t.Error("same instance of id must equal itself")
	}

	rng = func(_ []byte) (int, error) {
		return 0, errors.New("error")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	NewID() // Should panic.
}
