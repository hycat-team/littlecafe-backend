package config

import (
	"strings"
	"time"
)

// Config is the shared configuration model for every service. Service-owned
// sections live in that service's config package and are merged from the
// service overlay (configs/<service>.yaml).
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
	Admin      AdminConfig  `mapstructure:"admin"`
	I18n       I18n         `mapstructure:"i18n"`
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
	Algorithm     string `mapstructure:"algorithm"`
	PublicKeyFile string `mapstructure:"public_key_file"`
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

// AdminConfig holds admin-plane settings.
type AdminConfig struct {
	AuditRetentionMonths int `mapstructure:"audit_retention_months"`
}
