package jwtsigner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/jwttype"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/utils/jwtutils"
)

func writePEM(t *testing.T, dir, name, blockType string, der []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestLoadECDSAPrivateKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal PKCS#8: %v", err)
	}
	sec1, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal SEC1: %v", err)
	}

	dir := t.TempDir()
	paths := map[string]string{
		"pkcs8": writePEM(t, dir, "pkcs8.pem", "PRIVATE KEY", pkcs8),
		"sec1":  writePEM(t, dir, "sec1.pem", "EC PRIVATE KEY", sec1),
	}
	for name, path := range paths {
		got, err := LoadECDSAPrivateKey(path)
		if err != nil {
			t.Fatalf("%s: LoadECDSAPrivateKey: %v", name, err)
		}
		if !got.Equal(key) {
			t.Fatalf("%s: loaded key does not match generated key", name)
		}
	}
}

func TestGenerateTokenES256RoundTrip(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	userID := uuid.New()
	issuer := "littlecafe-identity"
	audience := "littlecafe-api"

	token, err := GenerateTokenES256(
		userID,
		"user@example.com",
		"USER",
		jwttype.AccessToken,
		key,
		issuer,
		audience,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateTokenES256: %v", err)
	}

	claims, err := jwtutils.ValidateTokenES256(token, jwttype.AccessToken, &key.PublicKey)
	if err != nil {
		t.Fatalf("ValidateTokenES256: %v", err)
	}
	if claims.Subject != userID.String() {
		t.Fatalf("Subject = %q, want %q", claims.Subject, userID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("Email = %q", claims.Email)
	}
	if claims.RoleSlug != "USER" {
		t.Fatalf("RoleSlug = %q", claims.RoleSlug)
	}
	if claims.Type != jwttype.AccessToken {
		t.Fatalf("Type = %q", claims.Type)
	}
	if claims.Issuer != issuer {
		t.Fatalf("Issuer = %q", claims.Issuer)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != audience {
		t.Fatalf("Audience = %v", claims.Audience)
	}
	if claims.ID == "" {
		t.Fatal("jti must not be empty")
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		t.Fatalf("ExpiresAt = %v, want a future time", claims.ExpiresAt)
	}

	if _, err := jwtutils.ValidateTokenES256(token, jwttype.RefreshToken, &key.PublicKey); err == nil {
		t.Fatal("token type mismatch must be rejected")
	}
}
