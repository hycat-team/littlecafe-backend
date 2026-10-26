package brevo

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"

	notificationconfig "github.com/hycat-team/littlecafe-backend/services/notification/internal/config"
	"github.com/hycat-team/littlecafe-backend/services/notification/internal/modules/notification/application/ports"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type captureLogger struct {
	infos  int
	warns  int
	errors int
}

func (c *captureLogger) Debug(string, ...zap.Field) { c.infos++ }
func (c *captureLogger) Info(string, ...zap.Field)  { c.infos++ }
func (c *captureLogger) Warn(string, ...zap.Field)  { c.warns++ }
func (c *captureLogger) Error(string, ...zap.Field) { c.errors++ }
func (c *captureLogger) Fatal(string, ...zap.Field) {}

func newConfig(enabled bool) *notificationconfig.Config {
	cfg := &notificationconfig.Config{}
	cfg.Email.Enabled = enabled
	cfg.Email.APIKey = "key"
	cfg.Email.Endpoint = "https://brevo.test/send"
	cfg.Email.SenderName = "LittleCafe"
	cfg.Email.SenderEmail = "no-reply@littlecafe.test"
	return cfg
}

func TestSendDisabled(t *testing.T) {
	log := &captureLogger{}
	sender := NewSender(newConfig(false), log)
	if err := sender.Send(context.Background(), ports.EmailMessage{Subject: "hello"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if log.infos != 1 {
		t.Fatalf("infos = %d", log.infos)
	}
}

func TestSendSuccess(t *testing.T) {
	var body string
	sender := NewSender(newConfig(true), &captureLogger{})
	sender.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("api-key") != "key" {
			return nil, errors.New("missing api key")
		}
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(bytes.NewBuffer(nil)), Header: make(http.Header)}, nil
	})}

	err := sender.Send(context.Background(), ports.EmailMessage{To: []string{"a@example.com", "b@example.com"}, Subject: "s", HTMLBody: "<p>x</p>"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	for _, want := range []string{`"subject":"s"`, `"htmlContent":"\u003cp\u003ex\u003c/p\u003e"`, `"email":"a@example.com"`, `"email":"b@example.com"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, missing %q", body, want)
		}
	}
}

func TestSendFailures(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		err      error
	}{
		{name: "transport", err: errors.New("network down")},
		{name: "status", response: &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBuffer(nil)), Header: make(http.Header)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sender := NewSender(newConfig(true), &captureLogger{})
			sender.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return tc.response, tc.err
			})}
			if err := sender.Send(context.Background(), ports.EmailMessage{To: []string{"a@example.com"}}); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestSendInvalidEndpoint(t *testing.T) {
	cfg := newConfig(true)
	cfg.Email.Endpoint = "://bad"
	sender := NewSender(cfg, &captureLogger{})
	if err := sender.Send(context.Background(), ports.EmailMessage{}); err == nil {
		t.Fatal("invalid endpoint must fail")
	}
}
