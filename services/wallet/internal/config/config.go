// Package config holds the Wallet service configuration: the shared sections
// plus the payment provider settings owned by this service.
package config

import (
	"log"

	sharedconfig "github.com/hycat-team/littlecafe-backend/shared/config"
)

// Config is the Wallet service configuration.
type Config struct {
	sharedconfig.Config `mapstructure:",squash"`

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

// Load reads configs/base.yaml merged with configs/wallet.yaml.
func Load() *Config {
	cfg := &Config{}
	if err := sharedconfig.Load("wallet", cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
