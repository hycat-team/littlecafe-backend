package timeutils

import (
	"testing"
	"time"
)

func TestBusinessLocation(t *testing.T) {
	loc := BusinessLocation()
	if loc == nil {
		t.Fatal("BusinessLocation returned nil")
	}
	_, offset := time.Now().In(loc).Zone()
	if offset != 7*60*60 {
		t.Fatalf("offset = %d, want UTC+7", offset)
	}
}

func TestFormatAPIUsesBusinessTimezone(t *testing.T) {
	input := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	got := FormatAPI(input)
	want := "2024-01-02T10:04:05+07:00"
	if got != want {
		t.Fatalf("FormatAPI = %q, want %q", got, want)
	}
}

func TestGetNow(t *testing.T) {
	before := time.Now().Add(-time.Minute)
	got := GetNow("")
	if got.Before(before) || time.Since(got) > time.Minute {
		t.Fatalf("GetNow() returned an unexpected time: %v", got)
	}
	if _, offset := got.Zone(); offset != 7*60*60 {
		t.Fatalf("default offset = %d, want UTC+7", offset)
	}

	utc := GetNow("UTC")
	if _, offset := utc.Zone(); offset != 0 {
		t.Fatalf("UTC offset = %d, want 0", offset)
	}

	fallback := GetNow("Invalid/Zone")
	if _, offset := fallback.Zone(); offset != 7*60*60 {
		t.Fatalf("fallback offset = %d, want UTC+7", offset)
	}
}
