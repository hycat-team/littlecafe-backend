package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/hycat-team/littlecafe-backend/shared/config"
)

// DirectDSN builds the session-scoped DSN used for LISTEN/NOTIFY and advisory
// locks. Session-scoped work must never run through a transaction pooler,
// which does not preserve session state.
func DirectDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SslMode,
		cfg.Database.TimeZone,
	)
}

// OpenDirectDB opens a small pool on the direct DSN for session-scoped work.
// Callers must close the underlying sql.DB.
func OpenDirectDB(cfg *config.Config) (*gorm.DB, error) {
	opened, err := gorm.Open(postgres.Open(DirectDSN(cfg)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open direct connection: %w", err)
	}
	sqlDB, err := opened.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve direct connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(2)
	return opened, nil
}
