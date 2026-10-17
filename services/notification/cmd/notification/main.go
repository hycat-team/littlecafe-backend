package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/services/notification/database/migrations"
	notificationconfig "github.com/hycat-team/littlecafe-backend/services/notification/internal/config"
	"github.com/hycat-team/littlecafe-backend/services/notification/internal/di"
	"github.com/hycat-team/littlecafe-backend/shared/infrastructure/db"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
)

// @title LittleCafe Backend - Notification API
// @version 1.0
// @description Backend platform for LittleCafe
// @BasePath /api/v1
func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local health endpoint and exit")
	migrateOnly := flag.Bool("migrate-only", false, "apply migrations and exit")
	flag.Parse()

	cfg := notificationconfig.Load()
	log := logger.New(cfg.App.Env, cfg.Logger.FilePath, cfg.Logger.LogLevel, cfg.Logger.LogToFile)

	if *healthcheck {
		os.Exit(probeHealth(cfg.Server.Port))
	}

	if err := db.RunMigrations(&cfg.Config, migrations.EmbedFS, log); err != nil {
		log.Fatal("failed to run database migrations", zap.Error(err))
	}
	if *migrateOnly {
		return
	}

	app, err := di.InitializeApp(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize application", zap.Error(err))
	}
	if err := app.Run(); err != nil {
		log.Fatal("application stopped with error", zap.Error(err))
	}
}

func probeHealth(port int) int {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", port)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "health returned %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
