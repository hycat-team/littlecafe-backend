package cryptoutils

import (
	"bytes"
	"testing"
)

func TestAESGCMRoundTripAndAAD(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	ciphertext, err := Encrypt(key, []byte("bank snapshot"), []byte("withdrawal-id"))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Decrypt(key, ciphertext, []byte("withdrawal-id"))
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "bank snapshot" {
		t.Fatal("unexpected plaintext")
	}
	if _, err := Decrypt(key, ciphertext, []byte("other")); err == nil {
		t.Fatal("wrong associated data must fail")
	}
}
