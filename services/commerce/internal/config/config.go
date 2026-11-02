// Package config holds the Commerce service configuration: the shared sections
// plus the Google OAuth client and the ES256 signing key owned by this service.
package config

import (
	"fmt"
	"log"
	"strings"

	sharedconfig "github.com/hycat-team/littlecafe-backend/shared/config"
)

// Config is the Commerce service configuration.
type Config struct {
	sharedconfig.Config `mapstructure:",squash"`

	Google  Google  `mapstructure:"google"`
	Signing Signing `mapstructure:"signing"`
	Payment Payment `mapstructure:"payment"`
}

// Payment holds the payment provider adapter settings (payOS collection and
// the feature-gated payout path).
type Payment struct {
	Provider           string `mapstructure:"provider"`
	ClientID           string `mapstructure:"client_id"`
	APIKey             string `mapstructure:"api_key"`
	ChecksumKey        string `mapstructure:"checksum_key"`
	ReturnURL          string `mapstructure:"return_url"`
	CancelURL          string `mapstructure:"cancel_url"`
	WebhookURL         string `mapstructure:"webhook_url"`
	PayoutEnabled      bool   `mapstructure:"payout_enabled"`
	ExpiredMinutes     int    `mapstructure:"expired_minutes"`
	ReconciliationCron string `mapstructure:"reconciliation_cron"`
}

// Google holds the Google OAuth/OIDC client settings.
type Google struct {
	WebClientID     string `mapstructure:"web_client_id"`
	WebClientSecret string `mapstructure:"web_client_secret"`
	RedirectURI     string `mapstructure:"redirect_uri"`
	StateTTLMinutes int    `mapstructure:"state_ttl_minutes"`
}

// Signing holds the ES256 private key owned by this service.
type Signing struct {
	PrivateKeyFile string `mapstructure:"private_key_file"`
}

// Validate enforces the service-owned requirements that the shared config
// cannot know about.
func (c *Config) Validate() error {
	if sharedconfig.IsProductionEnv(c.App.Env) && strings.TrimSpace(c.Signing.PrivateKeyFile) == "" {
		return fmt.Errorf("signing.private_key_file is required in production")
	}
	return nil
}

// Load reads configs/base.yaml merged with configs/commerce.yaml.
func Load() *Config {
	cfg := &Config{}
	if err := sharedconfig.Load("commerce", cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
