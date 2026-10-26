package jwtutils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/jwttype"
)

func newPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func signToken(t *testing.T, key any, method jwt.SigningMethod, tokenType jwttype.JwtType) string {
	t.Helper()
	claims := &CustomClaims{
		Email:    "user@example.com",
		RoleSlug: "USER",
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestValidateTokenES256(t *testing.T) {
	key := newPrivateKey(t)
	valid := signToken(t, key, jwt.SigningMethodES256, jwttype.AccessToken)

	claims, err := ValidateTokenES256(valid, jwttype.AccessToken, &key.PublicKey)
	if err != nil {
		t.Fatalf("ValidateTokenES256: %v", err)
	}
	if claims.Email != "user@example.com" || claims.RoleSlug != "USER" {
		t.Fatalf("claims = %+v", claims)
	}

	if _, err := ValidateTokenES256(valid, jwttype.AccessToken, nil); err == nil {
		t.Fatal("nil public key must fail")
	}

	if _, err := ValidateTokenES256(valid, jwttype.RefreshToken, &key.PublicKey); err == nil {
		t.Fatal("token type mismatch must fail")
	}

	hsToken := signToken(t, []byte("secret"), jwt.SigningMethodHS256, jwttype.AccessToken)
	if _, err := ValidateTokenES256(hsToken, jwttype.AccessToken, &key.PublicKey); err == nil {
		t.Fatal("non-ES256 token must fail")
	}

	if _, err := ValidateTokenES256("not-a-token", jwttype.AccessToken, &key.PublicKey); err == nil {
		t.Fatal("garbage token must fail")
	}
}

func TestLoadECDSAPublicKey(t *testing.T) {
	key := newPrivateKey(t)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	dir := t.TempDir()

	validPath := filepath.Join(dir, "public.pem")
	if err := os.WriteFile(validPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	loaded, err := LoadECDSAPublicKey(validPath)
	if err != nil || loaded == nil {
		t.Fatalf("LoadECDSAPublicKey = %v, %v", loaded, err)
	}

	if _, err := LoadECDSAPublicKey(filepath.Join(dir, "missing.pem")); err == nil {
		t.Fatal("missing file must fail")
	}

	invalidPEM := filepath.Join(dir, "invalid.pem")
	if err := os.WriteFile(invalidPEM, []byte("not pem"), 0o600); err != nil {
		t.Fatalf("write invalid pem: %v", err)
	}
	if _, err := LoadECDSAPublicKey(invalidPEM); err == nil {
		t.Fatal("invalid PEM must fail")
	}

	badDER := filepath.Join(dir, "bad-der.pem")
	if err := os.WriteFile(badDER, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("bad")}), 0o600); err != nil {
		t.Fatalf("write bad der: %v", err)
	}
	if _, err := LoadECDSAPublicKey(badDER); err == nil {
		t.Fatal("invalid PKIX must fail")
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	nonECDSA, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal rsa: %v", err)
	}
	rsaPath := filepath.Join(dir, "rsa.pem")
	if err := os.WriteFile(rsaPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: nonECDSA}), 0o600); err != nil {
		t.Fatalf("write rsa pem: %v", err)
	}
	if _, err := LoadECDSAPublicKey(rsaPath); err == nil {
		t.Fatal("non-ECDSA key must fail")
	}
}
