// Package config holds the Identity service configuration: the shared sections
// plus the Google OAuth client owned by this service.
package config

import (
	"log"

	sharedconfig "github.com/hycat-team/littlecafe-backend/shared/config"
)

// Config is the Identity service configuration.
type Config struct {
	sharedconfig.Config `mapstructure:",squash"`

	Google Google `mapstructure:"google"`
}

// Google holds the Google OAuth/OIDC client settings.
type Google struct {
	WebClientID     string `mapstructure:"web_client_id"`
	WebClientSecret string `mapstructure:"web_client_secret"`
	RedirectURI     string `mapstructure:"redirect_uri"`
	StateTTLMinutes int    `mapstructure:"state_ttl_minutes"`
}

// Load reads configs/base.yaml merged with configs/identity.yaml.
func Load() *Config {
	cfg := &Config{}
	if err := sharedconfig.Load("identity", cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
