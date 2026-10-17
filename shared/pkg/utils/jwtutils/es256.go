package jwtutils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadECDSAPublicKey reads an EC public key (PKIX PEM) for ES256 verification.
// Every service except identity holds only this key (FR-022).
func LoadECDSAPublicKey(path string) (*ecdsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read EC public key: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("decode EC public key: invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC public key: %w", err)
	}
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("PKIX key is not an ECDSA public key")
	}
	return ecPub, nil
}
