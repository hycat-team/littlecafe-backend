package id

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewReturnsUUIDv7(t *testing.T) {
	value, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if value == uuid.Nil {
		t.Fatal("New returned the nil UUID")
	}
	if version := value.Version(); version != 7 {
		t.Fatalf("version = %d, want 7", version)
	}
	if MustNew() == uuid.Nil {
		t.Fatal("MustNew returned the nil UUID")
	}
}
