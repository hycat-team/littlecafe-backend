package presentation

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/shared/application/sse"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func newTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, rec
}

func TestResponseHelpers(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/")
	Created(c, "created", gin.H{"id": "1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("Created status = %d", rec.Code)
	}

	c, rec = newTestContext(http.MethodPost, "/")
	Accepted(c, "accepted", nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("Accepted status = %d", rec.Code)
	}

	c, rec = newTestContext(http.MethodPost, "/")
	CreatedWithLocation(c, "", "created", nil)
	if got := rec.Header().Get("Location"); got != "" {
		t.Fatalf("empty location must not set the header, got %q", got)
	}

	c, rec = newTestContext(http.MethodPost, "/")
	AcceptedWithLocation(c, "/jobs/1", "accepted", nil)
	if got := rec.Header().Get("Location"); got != "/jobs/1" {
		t.Fatalf("Location = %q", got)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("AcceptedWithLocation status = %d", rec.Code)
	}
}

func TestLocalizedHelpers(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/")
	c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), i18n.EN))

	LocalizedCreated(c, i18n.SuccessOK, gin.H{"id": "1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("LocalizedCreated status = %d", rec.Code)
	}

	c, rec = newTestContext(http.MethodPost, "/")
	c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), i18n.EN))
	LocalizedAccepted(c, i18n.SuccessOK, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("LocalizedAccepted status = %d", rec.Code)
	}

	c, rec = newTestContext(http.MethodPost, "/")
	c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), i18n.EN))
	LocalizedCreatedWithLocation(c, "/orders/1", i18n.SuccessOK, nil)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/orders/1" {
		t.Fatalf("LocalizedCreatedWithLocation = %d, %q", rec.Code, rec.Header().Get("Location"))
	}

	c, rec = newTestContext(http.MethodPost, "/")
	c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), i18n.EN))
	LocalizedAcceptedWithLocation(c, "/jobs/1", i18n.SuccessOK, nil)
	if rec.Code != http.StatusAccepted || rec.Header().Get("Location") != "/jobs/1" {
		t.Fatalf("LocalizedAcceptedWithLocation = %d, %q", rec.Code, rec.Header().Get("Location"))
	}

	var body APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Message != "ok" {
		t.Fatalf("message = %q", body.Message)
	}
}

func TestWrapHandler(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/")
	WrapHandler(func(c *gin.Context) error {
		Success(c, "ok", nil)
		return nil
	})(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("success status = %d", rec.Code)
	}

	c, _ = newTestContext(http.MethodGet, "/")
	called := false
	WrapHandler(func(c *gin.Context) error {
		called = true
		return errors.New("boom")
	})(c)
	if !called || !c.IsAborted() {
		t.Fatalf("handler error must abort the context, called = %v", called)
	}
	if len(c.Errors) != 1 {
		t.Fatalf("errors = %v", c.Errors)
	}
}

func TestSSEHelpers(t *testing.T) {
	if got := SanitizeSSEData("line\nbreak"); got != `"line\nbreak"` {
		t.Fatalf("SanitizeSSEData = %q", got)
	}

	ResetSSEDeadlines()
	ConfigureSSEDeadlines(0)
	if currentChunkDeadline() != DefaultSSEChunkDeadline {
		t.Fatalf("zero deadline = %v", currentChunkDeadline())
	}
	ConfigureSSEDeadlines(-time.Second)
	if currentChunkDeadline() != DefaultSSEChunkDeadline {
		t.Fatalf("negative deadline = %v", currentChunkDeadline())
	}
	ConfigureSSEDeadlines(5 * time.Second)
	if currentChunkDeadline() != 5*time.Second {
		t.Fatalf("custom deadline = %v", currentChunkDeadline())
	}
	ResetSSEDeadlines()

	c, _ := newTestContext(http.MethodGet, "/")
	c.Request.Header.Set("Last-Event-ID", "evt-9")
	if got := LastEventID(c); got != "evt-9" {
		t.Fatalf("LastEventID = %q", got)
	}

	flusher, err := InitSSE(c)
	if err != nil || flusher == nil {
		t.Fatalf("InitSSE = %v, %v", flusher, err)
	}
	if got := c.Writer.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestWriteSSEWithoutMetadataAndMultiline(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/")
	if err := WriteSSE(c, rec, sse.Event{Data: json.RawMessage("{\"a\":1}\n{\"b\":2}")}); err != nil {
		t.Fatalf("WriteSSE: %v", err)
	}
	body := rec.Body.String()
	if strings.Contains(body, "id:") || strings.Contains(body, "event:") {
		t.Fatalf("unexpected metadata in %q", body)
	}
	if strings.Count(body, "data: ") != 2 {
		t.Fatalf("multiline data was not split: %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Fatalf("event must end with a blank line: %q", body)
	}
}

func TestSSERelayWithoutRedis(t *testing.T) {
	relay := NewSSERelay(nil, nil, &noopLogger{})
	relay.Start(nil)
	relay.Stop()
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...zap.Field) {}
func (noopLogger) Info(string, ...zap.Field)  {}
func (noopLogger) Warn(string, ...zap.Field)  {}
func (noopLogger) Error(string, ...zap.Field) {}
func (noopLogger) Fatal(string, ...zap.Field) {}
