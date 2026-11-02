package tokenutils

import "testing"

func TestRefreshTokenIsRandomAndStoredAsDigest(t *testing.T) {
	a, da, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	b, db, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || da == db {
		t.Fatal("tokens must be unique")
	}
	if da == a || da != Digest(a) {
		t.Fatal("digest contract violated")
	}
}
