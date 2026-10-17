package config

import (
	"fmt"
	"strings"
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
	return nil
}
