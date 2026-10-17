package config

import (
	"strings"
	"time"
)

// Config is the root configuration model.
type Config struct {
	App        App          `mapstructure:"app"`
	Server     Server       `mapstructure:"server"`
	Database   Database     `mapstructure:"database"`
	Jwt        Jwt          `mapstructure:"jwt"`
	Logger     Logger       `mapstructure:"logger"`
	Startup    Startup      `mapstructure:"startup"`
	Redis      Redis        `mapstructure:"redis"`
	RateLimit  RateLimit    `mapstructure:"rate_limit"`
	RabbitMQ   RabbitMQ     `mapstructure:"rabbitmq"`
	Outbox     OutboxConfig `mapstructure:"outbox"`
	Cloudinary Cloudinary   `mapstructure:"cloudinary"`
	Payment    Payment      `mapstructure:"payment"`
	Google     Google       `mapstructure:"google"`
	Email      Email        `mapstructure:"email"`
	Admin      AdminConfig  `mapstructure:"admin"`
}

// Redis holds cache/rate-limit connection settings.
type Redis struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	Db         int    `mapstructure:"db"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
}

// RateLimit configures the Redis token-bucket limiter.
type RateLimit struct {
	TokenLimit           int `mapstructure:"token_limit"`
	TokensPerPeriod      int `mapstructure:"tokens_per_period"`
	ReplenishmentSeconds int `mapstructure:"replenishment_seconds"`
}

// RabbitMQ holds broker connection settings.
type RabbitMQ struct {
	Host                 string `mapstructure:"host"`
	Port                 int    `mapstructure:"port"`
	ManagementPort       int    `mapstructure:"management_port"`
	User                 string `mapstructure:"user"`
	Password             string `mapstructure:"password"`
	Vhost                string `mapstructure:"vhost"`
	Prefetch             int    `mapstructure:"prefetch"`
	TLSEnabled           bool   `mapstructure:"tls_enabled"`
	ManagementTLSEnabled bool   `mapstructure:"management_tls_enabled"`
}

// OutboxConfig tunes the transactional outbox publisher.
type OutboxConfig struct {
	RetryMaxAttempts       int  `mapstructure:"retry_max_attempts"`
	RetryIntervalMinutes   int  `mapstructure:"retry_interval_minutes"`
	CleanupRetentionDays   int  `mapstructure:"cleanup_retention_days"`
	CleanupIntervalMinutes int  `mapstructure:"cleanup_interval_minutes"`
	BatchSize              int  `mapstructure:"batch_size"`
	ListenerEnabled        bool `mapstructure:"listener_enabled"`
}

// App selects the runtime topology of a service binary.
type App struct {
	Mode string `mapstructure:"mode"` // http | worker | all
	Env  string `mapstructure:"env"`  // development | staging | production
}

// Server holds HTTP and gRPC server settings.
type Server struct {
	Port     int `mapstructure:"port"`
	GRPCPort int `mapstructure:"grpc_port"`
	SSE      SSE `mapstructure:"sse"`
}

// ManagedEnvs lists server.env values that run on managed platforms, where
// managed-only requirements (infrastructure ID tokens, migrate-on-start) are
// enforced.
var ManagedEnvs = []string{"managed", "managed-drill"}

// IsManagedEnv reports whether env denotes a managed platform environment.
func IsManagedEnv(env string) bool {
	for _, e := range ManagedEnvs {
		if strings.EqualFold(strings.TrimSpace(env), e) {
			return true
		}
	}
	return false
}

// Database holds Postgres connection settings.
type Database struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SslMode         string        `mapstructure:"sslmode"`
	TimeZone        string        `mapstructure:"timezone"`
	MigrateOnStart  bool          `mapstructure:"migrate_on_start"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// Jwt holds the confirmed ES256 token settings.
type Jwt struct {
	Algorithm      string `mapstructure:"algorithm"`
	PrivateKeyFile string `mapstructure:"private_key_file"`
	PublicKeyFile  string `mapstructure:"public_key_file"`
}

// Logger holds logging settings.
type Logger struct {
	LogLevel  string `mapstructure:"log_level"`
	FilePath  string `mapstructure:"file_path"`
	LogToFile bool   `mapstructure:"log_to_file"`
}

// Startup holds the dependency connection retry plan (seconds per attempt).
type Startup struct {
	RetryAttempt1Seconds int `mapstructure:"retry_attempt_1_seconds"`
	RetryAttempt2Seconds int `mapstructure:"retry_attempt_2_seconds"`
	RetryAttempt3Seconds int `mapstructure:"retry_attempt_3_seconds"`
}

// IsProductionEnv reports whether env string denotes production semantics.
func IsProductionEnv(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return true
	default:
		return false
	}
}

// SSE holds server-sent events tuning.
type SSE struct {
	HeartbeatSeconds       int `mapstructure:"heartbeat_seconds"`
	ReplayBufferSize       int `mapstructure:"replay_buffer_size"`
	ReplayBufferTTLSeconds int `mapstructure:"replay_buffer_ttl_seconds"`
}

// Cloudinary holds media storage settings.
type Cloudinary struct {
	CloudName string `mapstructure:"cloud_name"`
	ApiKey    string `mapstructure:"api_key"`
	ApiSecret string `mapstructure:"api_secret"`
	APIUrl    string `mapstructure:"api_url"`
	Folder    string `mapstructure:"folder"`
}

// Payment holds the payment provider adapter settings.
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

// Google holds OAuth client settings.
type Google struct {
	WebClientID     string `mapstructure:"web_client_id"`
	WebClientSecret string `mapstructure:"web_client_secret"`
	RedirectURI     string `mapstructure:"redirect_uri"`
	StateTTLMinutes int    `mapstructure:"state_ttl_minutes"`
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

// AdminConfig holds admin-plane settings.
type AdminConfig struct {
	AuditRetentionMonths int `mapstructure:"audit_retention_months"`
}
