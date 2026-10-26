package entities

import (
	"reflect"
	"testing"
)

func TestVectorValue(t *testing.T) {
	var empty Vector
	value, err := empty.Value()
	if err != nil || value != nil {
		t.Fatalf("empty Value = %v, %v", value, err)
	}

	value, err = Vector{1, 2.5}.Value()
	if err != nil || value != "[1,2.5]" {
		t.Fatalf("Value = %v, %v", value, err)
	}
}

func TestVectorScan(t *testing.T) {
	var vector Vector
	if err := vector.Scan(nil); err != nil || vector != nil {
		t.Fatalf("Scan(nil) = %v, %v", vector, err)
	}
	if err := vector.Scan("[1,2.5]"); err != nil || !reflect.DeepEqual(vector, Vector{1, 2.5}) {
		t.Fatalf("Scan(string) = %v, %v", vector, err)
	}
	if err := vector.Scan([]byte("[3,4]")); err != nil || !reflect.DeepEqual(vector, Vector{3, 4}) {
		t.Fatalf("Scan([]byte) = %v, %v", vector, err)
	}
	if err := vector.Scan("[]"); err != nil || len(vector) != 0 {
		t.Fatalf("Scan(empty) = %v, %v", vector, err)
	}
	if err := vector.Scan("1,2"); err == nil {
		t.Fatal("missing brackets must fail")
	}
	if err := vector.Scan("[a]"); err == nil {
		t.Fatal("invalid float must fail")
	}
	if err := vector.Scan(42); err == nil {
		t.Fatal("unsupported type must fail")
	}
	if got := (Vector{}).GormDataType(); got != "vector" {
		t.Fatalf("GormDataType = %q", got)
	}
}
