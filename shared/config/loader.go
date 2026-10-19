package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`)

const (
	defaultBasePath = "configs/base.yaml"
	overlayDir      = "configs"
)

// ServiceConfig is implemented by per-service configuration structs. The shared
// sections live in an embedded Config; the service adds its own sections.
type ServiceConfig interface {
	Common() *Config
}

// Common returns the shared configuration, making *Config usable as a
// ServiceConfig for services that have no extra sections.
func (c *Config) Common() *Config { return c }

// Load reads configs/base.yaml, merges the service overlay, expands ${VAR}
// placeholders, applies env overrides, then validates.
//
// The overlay path comes from CONFIG_OVERLAY, or defaults to
// configs/<service>.yaml. A derived overlay that does not exist is skipped; an
// explicitly configured overlay must exist.
func Load(serviceName string, target ServiceConfig) error {
	basePath := envOrDefault("CONFIG_BASE", defaultBasePath)
	overlayPath := envOrDefault("CONFIG_OVERLAY", filepath.Join(overlayDir, serviceName+".yaml"))

	loadLocalDotEnv(basePath)

	v := viper.New()
	v.SetConfigType("yaml")

	baseBytes, err := os.ReadFile(basePath)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", basePath, err)
	}
	if err := v.ReadConfig(strings.NewReader(expandWithDefaults(string(baseBytes)))); err != nil {
		return fmt.Errorf("read config file %q: %w", basePath, err)
	}

	if serviceName != "" {
		overlayBytes, readErr := os.ReadFile(overlayPath)
		switch {
		case readErr == nil:
			if err := v.MergeConfig(strings.NewReader(expandWithDefaults(string(overlayBytes)))); err != nil {
				return fmt.Errorf("read config overlay %q: %w", overlayPath, err)
			}
		case os.IsNotExist(readErr) && os.Getenv("CONFIG_OVERLAY") == "":
			// The derived overlay is optional.
		default:
			return fmt.Errorf("read config overlay %q: %w", overlayPath, readErr)
		}
	}

	if err := v.UnmarshalExact(target); err != nil {
		return fmt.Errorf("decode configuration: %w", err)
	}

	common := target.Common()
	applyEnvOverrides(common)
	common.Server.FrontendOrigins = splitFrontendOrigins(common.Server.FrontendOriginsRaw)
	if err := validateConfig(common); err != nil {
		return err
	}
	if validator, ok := target.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MustLoad loads the common configuration for a service without extra
// sections, or exits.
func MustLoad(serviceName string) *Config {
	cfg := &Config{}
	if err := Load(serviceName, cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}

// expandWithDefaults expands ${VAR} and ${VAR:-default} using the environment.
func expandWithDefaults(s string) string {
	return envPattern.ReplaceAllStringFunc(s, func(match string) string {
		groups := envPattern.FindStringSubmatch(match)
		if v, ok := os.LookupEnv(groups[1]); ok && v != "" {
			return v
		}
		if len(groups) >= 4 {
			return groups[3]
		}
		return ""
	})
}

// loadLocalDotEnv loads the nearest .env without overriding existing variables.
func loadLocalDotEnv(configPath string) {
	candidates := []string{
		filepath.Join(filepath.Dir(configPath), ".env"),
		".env",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}

// splitFrontendOrigins parses the comma-separated FRONTEND_ORIGINS value into a
// normalized list (trimmed, trailing slash removed, empties dropped).
func splitFrontendOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := strings.TrimRight(strings.TrimSpace(part), "/")
		if normalized != "" {
			origins = append(origins, normalized)
		}
	}
	return origins
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APP_MODE"); v != "" {
		cfg.App.Mode = v
	}
	if v := os.Getenv("SERVER_ENV"); v != "" {
		cfg.App.Env = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.GRPCPort = n
		}
	}
	if v := os.Getenv("DB_MIGRATE_ON_START"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Database.MigrateOnStart = b
		}
	}
	if v := os.Getenv("I18N_DEFAULT_LOCALE"); v != "" {
		cfg.I18n.DefaultLocale = v
	}
	if v := os.Getenv("I18N_SUPPORTED_LOCALES"); v != "" {
		cfg.I18n.SupportedLocales = v
	}
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
