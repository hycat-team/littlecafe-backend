package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func validConfig() *Config {
	return &Config{
		App:    App{Mode: "all", Env: "development"},
		Server: Server{Port: 8080, FrontendOrigins: []string{"http://localhost:3000"}},
		Database: Database{
			Host: "localhost",
			Name: "littlecafe_identity",
			Port: 5432,
		},
		Jwt:       Jwt{Algorithm: "ES256"},
		Outbox:    OutboxConfig{BatchSize: 10},
		RateLimit: RateLimit{TokenLimit: 10, TokensPerPeriod: 10, ReplenishmentSeconds: 10},
	}
}

func TestValidateConfigAcceptsValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig: %v", err)
	}
	if cfg.Idempotency.TTLHours != 24 {
		t.Fatalf("TTLHours = %d, want the default", cfg.Idempotency.TTLHours)
	}
	if cfg.I18n.DefaultLocale != "vi" || cfg.I18n.SupportedLocales != "vi,en" {
		t.Fatalf("i18n defaults = %+v", cfg.I18n)
	}
}

func TestValidateConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{"mode", func(c *Config) { c.App.Mode = "invalid" }, "app.mode"},
		{"server port", func(c *Config) { c.Server.Port = 0 }, "server.port"},
		{"origins empty", func(c *Config) { c.Server.FrontendOrigins = nil }, "frontend_origins"},
		{"origin invalid", func(c *Config) { c.Server.FrontendOrigins = []string{"ftp://x"} }, "invalid origin"},
		{"database host", func(c *Config) { c.Database.Host = "" }, "database.host"},
		{"database name", func(c *Config) { c.Database.Name = "" }, "database.name"},
		{"database port", func(c *Config) { c.Database.Port = 70000 }, "database.port"},
		{"jwt algorithm", func(c *Config) { c.Jwt.Algorithm = "HS256" }, "jwt.algorithm"},
		{"jwt key production", func(c *Config) {
			c.App.Env = "production"
			c.Jwt.PublicKeyFile = ""
		}, "public_key_file"},
		{"redis port", func(c *Config) { c.Redis.Host = "localhost"; c.Redis.Port = 0 }, "redis.port"},
		{"rabbit port", func(c *Config) { c.RabbitMQ.Host = "localhost"; c.RabbitMQ.Port = 0 }, "rabbitmq.port"},
		{"outbox batch", func(c *Config) { c.Outbox.BatchSize = 0 }, "outbox.batch_size"},
		{"token limit", func(c *Config) { c.RateLimit.TokenLimit = 0 }, "token_limit"},
		{"tokens per period", func(c *Config) { c.RateLimit.TokensPerPeriod = 0 }, "tokens_per_period"},
		{"replenishment", func(c *Config) { c.RateLimit.ReplenishmentSeconds = 0 }, "replenishment_seconds"},
		{"idempotency negative", func(c *Config) { c.Idempotency.TTLHours = -1 }, "ttl_hours"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.mutate(cfg)
			err := validateConfig(cfg)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestValidateConfigAllowsWildcardOutsideProduction(t *testing.T) {
	cfg := validConfig()
	cfg.Server.FrontendOrigins = []string{"*"}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig: %v", err)
	}
}

func TestValidateI18n(t *testing.T) {
	cfg := validConfig()
	cfg.I18n.SupportedLocales = "vi,,en"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("empty entries must fail")
	}

	cfg = validConfig()
	cfg.I18n.SupportedLocales = "vi,fr"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("unsupported locales must fail")
	}

	cfg = validConfig()
	cfg.I18n.DefaultLocale = "fr"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("unsupported default locale must fail")
	}

	cfg = validConfig()
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.SupportedLocales = "vi"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("default locale must be part of the supported list")
	}

	cfg = validConfig()
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.SupportedLocales = "vi,en"
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig: %v", err)
	}
	if cfg.I18n.DefaultLocale != "en" || cfg.I18n.SupportedLocales != "vi,en" {
		t.Fatalf("i18n = %+v", cfg.I18n)
	}
}

func TestI18nAccessors(t *testing.T) {
	empty := I18n{}
	if empty.Default() != i18n.DefaultLocale {
		t.Fatalf("Default = %s", empty.Default())
	}
	if got := empty.Supported(); len(got) != len(i18n.SupportedLocales) {
		t.Fatalf("Supported = %v", got)
	}

	custom := I18n{DefaultLocale: "en", SupportedLocales: "en,vi,en,fr"}
	if custom.Default() != i18n.EN {
		t.Fatalf("Default = %s", custom.Default())
	}
	if got := custom.Supported(); len(got) != 2 || got[0] != i18n.EN || got[1] != i18n.VI {
		t.Fatalf("Supported = %v", got)
	}

	unsupported := I18n{SupportedLocales: "fr"}
	if got := unsupported.Supported(); len(got) != len(i18n.SupportedLocales) {
		t.Fatalf("fallback Supported = %v", got)
	}
}

func TestIsProductionEnv(t *testing.T) {
	for _, env := range []string{"production", "PROD", " Prod "} {
		if !IsProductionEnv(env) {
			t.Fatalf("IsProductionEnv(%q) = false", env)
		}
	}
	for _, env := range []string{"", "development", "staging"} {
		if IsProductionEnv(env) {
			t.Fatalf("IsProductionEnv(%q) = true", env)
		}
	}
}

func TestExpandWithDefaults(t *testing.T) {
	t.Setenv("CONFIG_TEST_VALUE", "value")
	got := expandWithDefaults("${CONFIG_TEST_VALUE}|${CONFIG_TEST_MISSING:-fallback}|${CONFIG_TEST_MISSING}")
	if got != "value|fallback|" {
		t.Fatalf("expandWithDefaults = %q", got)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("APP_MODE", "worker")
	t.Setenv("SERVER_ENV", "staging")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("GRPC_PORT", "9001")
	t.Setenv("DB_MIGRATE_ON_START", "true")
	t.Setenv("I18N_DEFAULT_LOCALE", "en")
	t.Setenv("I18N_SUPPORTED_LOCALES", "en,vi")
	t.Setenv("SERVER_PORT", "9000")

	cfg := &Config{}
	applyEnvOverrides(cfg)
	if cfg.App.Mode != "worker" || cfg.App.Env != "staging" || cfg.Server.Port != 9000 ||
		cfg.Server.GRPCPort != 9001 || !cfg.Database.MigrateOnStart ||
		cfg.I18n.DefaultLocale != "en" || cfg.I18n.SupportedLocales != "en,vi" {
		t.Fatalf("cfg = %+v", cfg)
	}

	clearEnvOverrides(t)
	t.Setenv("SERVER_PORT", "not-a-number")
	t.Setenv("GRPC_PORT", "not-a-number")
	t.Setenv("DB_MIGRATE_ON_START", "not-a-bool")
	cfg = &Config{Server: Server{Port: 1, GRPCPort: 2}}
	applyEnvOverrides(cfg)
	if cfg.Server.Port != 1 || cfg.Server.GRPCPort != 2 {
		t.Fatalf("invalid values must be ignored: %+v", cfg.Server)
	}
}

func TestLoadMissingBaseFile(t *testing.T) {
	clearEnvOverrides(t)
	t.Setenv("CONFIG_BASE", filepath.Join(t.TempDir(), "missing.yaml"))
	if err := Load("order", &Config{}); err == nil {
		t.Fatal("missing base config must fail")
	}
}

func TestLoadReadsNearestDotEnv(t *testing.T) {
	clearEnvOverrides(t)
	dir := t.TempDir()
	base := writeTempConfig(t, dir, "base.yaml", strings.ReplaceAll(validBaseConfig, "port: 8080", "port: ${TEST_DOTENV_PORT:-8080}"))
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("TEST_DOTENV_PORT=8091\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("CONFIG_BASE", base)
	cfg := &Config{}
	if err := Load("order", cfg); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 8091 {
		t.Fatalf("Server.Port = %d, want 8091", cfg.Server.Port)
	}
	_ = os.Unsetenv("TEST_DOTENV_PORT")
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("CONFIG_TEST_OR_DEFAULT", "  ")
	if got := envOrDefault("CONFIG_TEST_OR_DEFAULT", "fallback"); got != "fallback" {
		t.Fatalf("envOrDefault = %q", got)
	}
	t.Setenv("CONFIG_TEST_OR_DEFAULT", " value ")
	if got := envOrDefault("CONFIG_TEST_OR_DEFAULT", "fallback"); got != "value" {
		t.Fatalf("envOrDefault = %q", got)
	}
}
