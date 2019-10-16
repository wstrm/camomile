package main

import "testing"

func TestFlagSplit(t *testing.T) {
	a, b := flagSplit("hej@då")
	if a != "hej" {
		t.Errorf("unexpected part 1: %s", a)
	}
	if b != "då" {
		t.Errorf("unexpected part 2: %s", a)
	}

	a, b = flagSplit("")
	if a != "" || b != "" {
		t.Error("expected empty strings, got:", a, b)
	}
}
