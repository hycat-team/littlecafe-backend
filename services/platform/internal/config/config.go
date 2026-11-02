// Package config holds the Platform service configuration: the shared
// sections plus the outbound email settings owned by this service.
package config

import (
	"log"

	sharedconfig "github.com/hycat-team/littlecafe-backend/shared/config"
)

// Config is the Platform service configuration.
type Config struct {
	sharedconfig.Config `mapstructure:",squash"`

	Email Email `mapstructure:"email"`
}

// Email holds outbound Brevo settings.
type Email struct {
	Enabled     bool   `mapstructure:"enabled"`
	Provider    string `mapstructure:"provider"`
	APIKey      string `mapstructure:"api_key"`
	Endpoint    string `mapstructure:"endpoint"`
	SenderName  string `mapstructure:"sender_name"`
	SenderEmail string `mapstructure:"sender_email"`
	ReplyTo     string `mapstructure:"reply_to"`
}

// Load reads configs/base.yaml merged with configs/platform.yaml.
func Load() *Config {
	cfg := &Config{}
	if err := sharedconfig.Load("platform", cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
