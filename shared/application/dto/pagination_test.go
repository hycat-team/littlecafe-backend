package dto

import "testing"

func TestCursorRoundTrip(t *testing.T) {
	encoded := EncodeCursor("order-123")
	if encoded == "" {
		t.Fatal("EncodeCursor returned an empty string")
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil || decoded != "order-123" {
		t.Fatalf("DecodeCursor = %q, %v", decoded, err)
	}
	if _, err := DecodeCursor("!!!"); err == nil {
		t.Fatal("invalid cursors must fail to decode")
	}
}
