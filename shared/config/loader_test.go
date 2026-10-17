package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testServiceConfig struct {
	Config `mapstructure:",squash"`
	Google testGoogle `mapstructure:"google"`
}

type testGoogle struct {
	WebClientID string `mapstructure:"web_client_id"`
}

const validBaseConfig = `
app:
  mode: all
  env: development
server:
  port: 8080
  grpc_port: 9090
database:
  host: localhost
  port: 5432
  name: littlecafe_identity
  sslmode: disable
  timezone: Asia/Ho_Chi_Minh
jwt:
  algorithm: ES256
outbox:
  batch_size: 100
rate_limit:
  token_limit: 60
  tokens_per_period: 60
  replenishment_seconds: 60
`

func writeTempConfig(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func clearEnvOverrides(t *testing.T) {
	t.Helper()
	for _, key := range []string{"APP_MODE", "SERVER_ENV", "SERVER_PORT", "GRPC_PORT", "DB_MIGRATE_ON_START", "CONFIG_OVERLAY"} {
		t.Setenv(key, "")
	}
}

func TestLoadMergesBaseAndOverlay(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", validBaseConfig)
	overlay := writeTempConfig(t, dir, "identity.yaml", "google:\n  web_client_id: ${TEST_GOOGLE_CLIENT_ID:-default-client}\n")

	t.Setenv("CONFIG_BASE", base)
	t.Setenv("CONFIG_OVERLAY", overlay)
	t.Setenv("TEST_GOOGLE_CLIENT_ID", "client-123")

	var cfg testServiceConfig
	if err := Load("identity", &cfg); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Google.WebClientID != "client-123" {
		t.Fatalf("WebClientID = %q, want env-expanded value", cfg.Google.WebClientID)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("Server.Port = %d, want base value", cfg.Server.Port)
	}
	if cfg.Database.Name != "littlecafe_identity" {
		t.Fatalf("Database.Name = %q", cfg.Database.Name)
	}
}

func TestLoadSkipsMissingDerivedOverlay(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", validBaseConfig)
	t.Setenv("CONFIG_BASE", base)

	cfg := &Config{}
	if err := Load("order", cfg); err != nil {
		t.Fatalf("Load must skip a missing derived overlay, got: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("Server.Port = %d, want base value", cfg.Server.Port)
	}
}

func TestLoadFailsOnMissingExplicitOverlay(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", validBaseConfig)
	t.Setenv("CONFIG_BASE", base)
	t.Setenv("CONFIG_OVERLAY", filepath.Join(dir, "missing.yaml"))

	cfg := &Config{}
	err := Load("order", cfg)
	if err == nil {
		t.Fatal("Load must fail when an explicit overlay is missing")
	}
	if !strings.Contains(err.Error(), "missing.yaml") {
		t.Fatalf("error should name the missing overlay, got: %v", err)
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", validBaseConfig+"unknown_section:\n  value: 1\n")
	t.Setenv("CONFIG_BASE", base)

	cfg := &Config{}
	if err := Load("order", cfg); err == nil {
		t.Fatal("Load must reject unknown configuration keys")
	}
}

func TestLoadRunsServiceValidator(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", validBaseConfig)
	t.Setenv("CONFIG_BASE", base)

	target := &validatingConfig{}
	err := Load("wallet", target)
	if !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("service Validate must run, got: %v", err)
	}
}

type validatingConfig struct {
	Config `mapstructure:",squash"`
}

func (v *validatingConfig) Validate() error {
	return os.ErrInvalid
}
