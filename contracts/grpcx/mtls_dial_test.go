package grpcx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clearTLSEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"GRPC_TLS_CA_FILE", "GRPC_TLS_CERT_FILE", "GRPC_TLS_KEY_FILE", "GRPC_MTLS_REQUIRED"} {
		t.Setenv(key, "")
	}
}

func TestServerOptionsWithoutTLS(t *testing.T) {
	clearTLSEnv(t)
	options, err := ServerOptionsFromEnv()
	if err != nil {
		t.Fatalf("ServerOptionsFromEnv: %v", err)
	}
	if len(options) != 2 {
		t.Fatalf("options = %d, want the locale interceptors only", len(options))
	}
}

func TestServerOptionsRequiresCompleteConfiguration(t *testing.T) {
	clearTLSEnv(t)
	t.Setenv("GRPC_MTLS_REQUIRED", "true")
	if _, err := ServerOptionsFromEnv(); err == nil {
		t.Fatal("missing certificates must fail when mTLS is required")
	}

	clearTLSEnv(t)
	t.Setenv("GRPC_TLS_CA_FILE", "ca.pem")
	if _, err := ServerOptionsFromEnv(); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("partial configuration must fail, got %v", err)
	}
}

func TestServerOptionsInvalidFiles(t *testing.T) {
	clearTLSEnv(t)
	dir := t.TempDir()
	t.Setenv("GRPC_TLS_CA_FILE", filepath.Join(dir, "missing-ca.pem"))
	t.Setenv("GRPC_TLS_CERT_FILE", filepath.Join(dir, "missing-cert.pem"))
	t.Setenv("GRPC_TLS_KEY_FILE", filepath.Join(dir, "missing-key.pem"))
	if _, err := ServerOptionsFromEnv(); err == nil {
		t.Fatal("missing files must fail")
	}
}

func TestClientTransportCredentialsRequiresFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := ClientTransportCredentials(MTLSFiles{
		CAFile:   filepath.Join(dir, "missing-ca.pem"),
		CertFile: filepath.Join(dir, "missing-cert.pem"),
		KeyFile:  filepath.Join(dir, "missing-key.pem"),
	}); err == nil {
		t.Fatal("missing CA must fail")
	}

	ca := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(ca, []byte("not-a-cert"), 0o600); err != nil {
		t.Fatalf("write ca: %v", err)
	}
	if _, err := ClientTransportCredentials(MTLSFiles{CAFile: ca}); err == nil || !strings.Contains(err.Error(), "invalid CA") {
		t.Fatalf("invalid CA = %v", err)
	}
}

func TestServerOptionRequiresCA(t *testing.T) {
	dir := t.TempDir()
	if _, err := ServerOption(MTLSFiles{CAFile: filepath.Join(dir, "missing.pem")}); err == nil {
		t.Fatal("missing CA must fail")
	}
}

func TestDialService(t *testing.T) {
	if _, err := DialService("", MTLSFiles{}); err == nil {
		t.Fatal("empty URL must fail")
	}
	if _, err := DialService("https://identity:9090", MTLSFiles{
		CAFile:   filepath.Join(t.TempDir(), "missing.pem"),
		CertFile: filepath.Join(t.TempDir(), "missing.pem"),
		KeyFile:  filepath.Join(t.TempDir(), "missing.pem"),
	}); err == nil {
		t.Fatal("missing credentials must fail")
	}
}
