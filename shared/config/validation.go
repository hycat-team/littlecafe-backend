package config

import (
	"fmt"
	"strings"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func validateConfig(cfg *Config) error {
	validModes := map[string]bool{"http": true, "worker": true, "all": true}
	if cfg.App.Mode == "" {
		cfg.App.Mode = "all"
	}
	if !validModes[cfg.App.Mode] {
		return fmt.Errorf("app.mode must be one of http, worker, all (got %q)", cfg.App.Mode)
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be within 1-65535 (got %d)", cfg.Server.Port)
	}
	if len(cfg.Server.FrontendOrigins) == 0 {
		return fmt.Errorf("server.frontend_origins must contain at least one origin")
	}
	for _, origin := range cfg.Server.FrontendOrigins {
		if origin == "*" {
			if IsProductionEnv(cfg.App.Env) {
				return fmt.Errorf("server.frontend_origins must not contain * in production")
			}
			continue
		}
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("server.frontend_origins contains invalid origin %q", origin)
		}
	}
	if strings.TrimSpace(cfg.Database.Host) == "" {
		return fmt.Errorf("database.host is required")
	}
	if strings.TrimSpace(cfg.Database.Name) == "" {
		return fmt.Errorf("database.name is required")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("database.port must be within 1-65535 (got %d)", cfg.Database.Port)
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.Jwt.Algorithm), "ES256") {
		return fmt.Errorf("jwt.algorithm must be ES256")
	}
	if IsProductionEnv(cfg.App.Env) && strings.TrimSpace(cfg.Jwt.PublicKeyFile) == "" {
		return fmt.Errorf("jwt.public_key_file is required in production")
	}
	if cfg.Redis.Host != "" && (cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535) {
		return fmt.Errorf("redis.port must be within 1-65535 (got %d)", cfg.Redis.Port)
	}
	if cfg.RabbitMQ.Host != "" && (cfg.RabbitMQ.Port <= 0 || cfg.RabbitMQ.Port > 65535) {
		return fmt.Errorf("rabbitmq.port must be within 1-65535 (got %d)", cfg.RabbitMQ.Port)
	}
	if cfg.Outbox.BatchSize <= 0 {
		return fmt.Errorf("outbox.batch_size must be positive (got %d)", cfg.Outbox.BatchSize)
	}
	if cfg.RateLimit.TokenLimit <= 0 {
		return fmt.Errorf("rate_limit.token_limit must be positive (got %d)", cfg.RateLimit.TokenLimit)
	}
	if cfg.RateLimit.TokensPerPeriod <= 0 {
		return fmt.Errorf("rate_limit.tokens_per_period must be positive (got %d)", cfg.RateLimit.TokensPerPeriod)
	}
	if cfg.RateLimit.ReplenishmentSeconds <= 0 {
		return fmt.Errorf("rate_limit.replenishment_seconds must be positive (got %d)", cfg.RateLimit.ReplenishmentSeconds)
	}
	if cfg.Idempotency.TTLHours == 0 {
		cfg.Idempotency.TTLHours = 24
	}
	if cfg.Idempotency.TTLHours < 0 {
		return fmt.Errorf("idempotency.ttl_hours must be positive (got %d)", cfg.Idempotency.TTLHours)
	}
	if err := validateI18n(cfg); err != nil {
		return err
	}
	return nil
}

func validateI18n(cfg *Config) error {
	rawSupported := strings.TrimSpace(cfg.I18n.SupportedLocales)
	if rawSupported == "" {
		rawSupported = joinLocales(i18n.SupportedLocales)
	}

	supported := make([]i18n.Locale, 0, 2)
	for _, part := range strings.Split(rawSupported, ",") {
		loc := i18n.Locale(strings.TrimSpace(part))
		if loc == "" {
			return fmt.Errorf("i18n.supported_locales must not contain empty entries")
		}
		if !i18n.IsSupported(loc) {
			return fmt.Errorf("i18n.supported_locales contains unsupported locale %q", loc)
		}
		supported = append(supported, loc)
	}

	defaultLocale := i18n.Locale(strings.TrimSpace(cfg.I18n.DefaultLocale))
	if defaultLocale == "" {
		defaultLocale = i18n.DefaultLocale
	}
	if !i18n.IsSupported(defaultLocale) {
		return fmt.Errorf("i18n.default_locale %q is not supported", defaultLocale)
	}
	if !containsLocale(supported, defaultLocale) {
		return fmt.Errorf("i18n.default_locale %q must be listed in i18n.supported_locales", defaultLocale)
	}

	cfg.I18n.SupportedLocales = rawSupported
	cfg.I18n.DefaultLocale = string(defaultLocale)
	return nil
}

func joinLocales(locales []i18n.Locale) string {
	values := make([]string, 0, len(locales))
	for _, loc := range locales {
		values = append(values, string(loc))
	}
	return strings.Join(values, ",")
}

func containsLocale(locales []i18n.Locale, target i18n.Locale) bool {
	for _, loc := range locales {
		if loc == target {
			return true
		}
	}
	return false
}
