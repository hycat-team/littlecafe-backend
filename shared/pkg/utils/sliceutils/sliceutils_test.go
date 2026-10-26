package sliceutils

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestAdjustVectorLength(t *testing.T) {
	original := []float32{1, 2, 3}
	if got := AdjustVectorLength(original, 3); !reflect.DeepEqual(got, original) {
		t.Fatalf("same length = %v", got)
	}
	if got := AdjustVectorLength(original, 2); !reflect.DeepEqual(got, []float32{1, 2}) {
		t.Fatalf("truncate = %v", got)
	}
	if got := AdjustVectorLength(original, 5); !reflect.DeepEqual(got, []float32{1, 2, 3, 0, 0}) {
		t.Fatalf("pad = %v", got)
	}
}

func TestUniqueAndSortStrings(t *testing.T) {
	got := UniqueAndSortStrings([]string{"b", "a", "b", "c"})
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("got %v", got)
	}
}

func TestAppendUniqueStringCaseInsensitive(t *testing.T) {
	base := []string{"Alpha"}
	if got := AppendUniqueStringCaseInsensitive(base, "alpha"); !reflect.DeepEqual(got, base) {
		t.Fatalf("duplicate should be ignored, got %v", got)
	}
	if got := AppendUniqueStringCaseInsensitive(base, "beta"); !reflect.DeepEqual(got, []string{"Alpha", "beta"}) {
		t.Fatalf("append = %v", got)
	}
}

func TestUniqueUUIDs(t *testing.T) {
	if got := UniqueUUIDs(nil); got != nil {
		t.Fatalf("empty = %v, want nil", got)
	}
	a, b := uuid.New(), uuid.New()
	got := UniqueUUIDs([]uuid.UUID{a, b, a})
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("got %v", got)
	}
}

func TestIndexBy(t *testing.T) {
	type item struct {
		ID   string
		Name string
	}
	items := []item{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}}
	got := IndexBy(items, func(i item) string { return i.ID })
	if got["a"].Name != "A" || got["b"].Name != "B" {
		t.Fatalf("got %v", got)
	}
}
