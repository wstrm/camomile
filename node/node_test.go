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

func TestIDFromStringValid(t *testing.T) {
	id, err := IDFromString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	for _, b := range id {
		if byte(0xff) != b {
			t.Errorf("unexpected byte: %b", b)
		}
	}
}

func TestIDFromStringInvalidLength(t *testing.T) {
	_, err := IDFromString("ffffffffffffffffffffffffffffffffffff")
	if err.Error() != "hex string must be 32 bytes" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestIDFromStringInvalidHex(t *testing.T) {
	_, err := IDFromString("ABC, du är mina tankar")
	if err.Error() != "cannot decode hex string as ID: encoding/hex: invalid byte: U+002C ','" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}
