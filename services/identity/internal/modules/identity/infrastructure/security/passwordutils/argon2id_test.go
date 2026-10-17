package passwordutils

import "testing"

func TestArgon2id(t *testing.T) {
	encoded, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(encoded, "correct horse battery staple") {
		t.Fatal("valid password rejected")
	}
	if Verify(encoded, "wrong") {
		t.Fatal("invalid password accepted")
	}
}
