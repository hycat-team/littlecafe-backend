// Package brevo implements the ports.EmailSender contract against the Brevo
// transactional email API.
package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	notificationconfig "github.com/hycat-team/littlecafe-backend/services/notification/internal/config"
	"github.com/hycat-team/littlecafe-backend/services/notification/internal/modules/notification/application/ports"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
	"go.uber.org/zap"
)

type Sender struct {
	cfg    *notificationconfig.Config
	log    logger.Interface
	client *http.Client
}

func NewSender(cfg *notificationconfig.Config, log logger.Interface) *Sender {
	return &Sender{cfg: cfg, log: log, client: http.DefaultClient}
}

func (s *Sender) Send(ctx context.Context, message ports.EmailMessage) error {
	if !s.cfg.Email.Enabled {
		s.log.Info("email sending disabled", zap.String("subject", message.Subject))
		return nil
	}
	recipients := make([]map[string]string, 0, len(message.To))
	for _, address := range message.To {
		recipients = append(recipients, map[string]string{"email": address})
	}
	payload := map[string]any{"sender": map[string]string{"name": s.cfg.Email.SenderName, "email": s.cfg.Email.SenderEmail}, "to": recipients, "subject": message.Subject, "htmlContent": message.HTMLBody}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.Email.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", s.cfg.Email.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("brevo send: status %d", resp.StatusCode)
	}
	return nil
}

var _ ports.EmailSender = (*Sender)(nil)
