package config

import (
	"bytes"
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

// LoadConfig loads configuration or exits: invalid configuration at startup
// must never be silently ignored.
func LoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}

func loadConfig() (*Config, error) {
	configPath := envOrDefault("CONFIG_FILE", filepath.Join("configs", "config.yaml"))
	loadLocalDotEnv(configPath)

	rawBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", configPath, err)
	}
	expanded := expandWithDefaults(string(rawBytes))

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return nil, fmt.Errorf("read config file %q: %w", configPath, err)
	}

	var cfg Config
	if err := v.UnmarshalExact(&cfg); err != nil {
		return nil, fmt.Errorf("decode config file %q: %w", configPath, err)
	}
	applyEnvOverrides(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
