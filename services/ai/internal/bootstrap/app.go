package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
)

// App owns the HTTP server lifecycle.
type App struct {
	cfg     *config.Config
	log     logger.Interface
	engine  *gin.Engine
	workers *Workers
}

// NewApp wires the application dependencies produced by Wire.
func NewApp(cfg *config.Config, log logger.Interface, engine *gin.Engine, workers *Workers) *App {
	return &App{cfg: cfg, log: log, engine: engine, workers: workers}
}

// Run starts the HTTP server and blocks until
// a shutdown signal arrives.
func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           a.engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()
	a.workers.Start(workerCtx)
	a.log.Info("background workers started")
	go func() {
		a.log.Info("HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-quit:
		a.log.Info("shutdown signal received")
		a.workers.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		a.log.Info("server stopped")
		return nil
	}
}
