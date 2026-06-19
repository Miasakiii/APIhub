package util

import (
	"testing"
)

func TestGenerateID_Length(t *testing.T) {
	id := GenerateID()
	if len(id) != 32 {
		t.Fatalf("expected 32 chars, got %d: %q", len(id), id)
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := GenerateID()
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestGenerateID_HexChars(t *testing.T) {
	id := GenerateID()
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex char %c in ID %s", c, id)
		}
	}
}
