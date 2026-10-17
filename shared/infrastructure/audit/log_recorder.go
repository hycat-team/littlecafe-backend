// Package audit provides a logger-backed audit recorder for local development
// and tests. Replace it with a persistence adapter when audit retention is
// required.
package audit

import (
	"context"

	"github.com/hycat-team/littlecafe-backend/shared/application/audit"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"

	"go.uber.org/zap"
)

// LogRecorder writes audit entries to the structured log.
type LogRecorder struct {
	cfg *config.Config
	log logger.Interface
}

// NewLogRecorder builds the logger-backed recorder.
func NewLogRecorder(cfg *config.Config, log logger.Interface) *LogRecorder {
	return &LogRecorder{cfg: cfg, log: log}
}

// Record emits one audit entry.
func (r *LogRecorder) Record(_ context.Context, entry audit.Entry) error {
	r.log.Info("admin audit",
		zap.String("actorId", entry.ActorID),
		zap.String("action", entry.Action),
		zap.String("targetType", entry.TargetType),
		zap.String("targetId", entry.TargetID),
		zap.Int("retentionMonths", r.cfg.Admin.AuditRetentionMonths),
	)
	return nil
}

var _ audit.Recorder = (*LogRecorder)(nil)
