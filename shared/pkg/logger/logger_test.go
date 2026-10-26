package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestNewConsoleLogger(t *testing.T) {
	for _, env := range []string{"Development", "dev", "production"} {
		log := New(env, "", "invalid-level", false)
		if log == nil {
			t.Fatalf("New(%q) returned nil", env)
		}
		log.Debug("debug", zap.String("k", "v"))
		log.Info("info")
		log.Warn("warn")
		log.Error("error")
	}
}

func TestNewFileLogger(t *testing.T) {
	dir, err := os.MkdirTemp("", "logger-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "logs", "app.log")
	log := New("production", path, "info", true)
	log.Info("written to file")

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log file was not created: %v", err)
	}
}
