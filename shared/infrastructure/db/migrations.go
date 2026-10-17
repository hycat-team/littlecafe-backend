package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
)

// RunMigrations applies the service's embedded Goose migrations on startup when
// database.migrate_on_start is true. It uses a direct (session-scoped)
// connection, never a transaction pooler.
func RunMigrations(cfg *config.Config, migrationsFS embed.FS, l logger.Interface) error {
	if !cfg.Database.MigrateOnStart {
		return nil
	}

	gdb, err := gorm.Open(postgres.Open(DirectDSN(cfg)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("migrations: open connection: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("migrations: resolve connection: %w", err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)

	ctx := context.Background()
	lockKey := "littlecafe:migrate:" + cfg.Database.Name
	if _, err := sqlDB.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", lockKey); err != nil {
		return fmt.Errorf("migrations: acquire advisory lock: %w", err)
	}
	defer func() {
		_, _ = sqlDB.ExecContext(ctx, "SELECT pg_advisory_unlock(hashtext($1))", lockKey)
	}()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrations: set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("migrations: up: %w", err)
	}
	l.Info("database migrations applied")
	return nil
}
