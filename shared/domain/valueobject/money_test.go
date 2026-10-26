package valueobject

import "testing"

func TestMoneyVND(t *testing.T) {
	if _, err := NewMoneyVND(-1); err == nil {
		t.Fatal("negative amounts must be rejected")
	}
	zero, err := NewMoneyVND(0)
	if err != nil || zero.Int64() != 0 {
		t.Fatalf("zero = %v, %v", zero, err)
	}
	amount, err := NewMoneyVND(150_000)
	if err != nil || amount.Int64() != 150_000 {
		t.Fatalf("amount = %v, %v", amount, err)
	}
}
