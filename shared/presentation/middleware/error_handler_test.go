package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

type stubLogger struct{}

func (stubLogger) Debug(string, ...zap.Field) {}
func (stubLogger) Info(string, ...zap.Field)  {}
func (stubLogger) Warn(string, ...zap.Field)  {}
func (stubLogger) Error(string, ...zap.Field) {}
func (stubLogger) Fatal(string, ...zap.Field) {}

func newTestEngine(env string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(GlobalErrorHandler(stubLogger{}, env))
	engine.GET("/test", handler)
	return engine
}

func performRequest(t *testing.T, engine *gin.Engine, traceID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (body=%q)", err, rec.Body.String())
	}
	return body
}

func TestGlobalErrorHandlerSerializesProblemDetails(t *testing.T) {
	engine := newTestEngine("development", func(c *gin.Context) {
		_ = c.Error(apperror.NewNotFound("Không tìm thấy đơn hàng"))
		c.Abort()
	})

	rec := performRequest(t, engine, "trace-123")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}

	body := decodeBody(t, rec)
	if body["type"] != "about:blank" {
		t.Fatalf("type = %v, want about:blank", body["type"])
	}
	if body["title"] != "Không tìm thấy dữ liệu" {
		t.Fatalf("title = %v", body["title"])
	}
	if body["detail"] != "Không tìm thấy đơn hàng" {
		t.Fatalf("detail = %v", body["detail"])
	}
	if body["instance"] != "/test" {
		t.Fatalf("instance = %v, want /test", body["instance"])
	}
	if body["traceId"] != "trace-123" {
		t.Fatalf("traceId = %v, want trace-123", body["traceId"])
	}
}

func TestGlobalErrorHandlerProductionHidesDebugFields(t *testing.T) {
	engine := newTestEngine("production", func(c *gin.Context) {
		_ = c.Error(errors.New("pq: duplicate key value violates unique constraint"))
		c.Abort()
	})

	rec := performRequest(t, engine, "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	body := decodeBody(t, rec)
	if body["detail"] != i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalGeneric) {
		t.Fatalf("detail = %v, want generic message", body["detail"])
	}
	if _, exists := body["debugMessage"]; exists {
		t.Fatal("production must not expose debugMessage")
	}
	if _, exists := body["stackTrace"]; exists {
		t.Fatal("production must not expose stackTrace")
	}
	if strings.Contains(rec.Body.String(), "duplicate key") {
		t.Fatalf("raw error leaked in production: %s", rec.Body.String())
	}
}

func TestGlobalErrorHandlerDevelopmentExposesDebugFields(t *testing.T) {
	engine := newTestEngine("development", func(c *gin.Context) {
		_ = c.Error(errors.New("pq: duplicate key value violates unique constraint"))
		c.Abort()
	})

	rec := performRequest(t, engine, "")

	body := decodeBody(t, rec)
	if body["debugMessage"] != "pq: duplicate key value violates unique constraint" {
		t.Fatalf("debugMessage = %v", body["debugMessage"])
	}
	if _, exists := body["stackTrace"]; !exists {
		t.Fatal("development must expose stackTrace")
	}
}

func TestGlobalErrorHandlerRecoversPanic(t *testing.T) {
	engine := newTestEngine("development", func(*gin.Context) {
		panic("boom")
	})

	rec := performRequest(t, engine, "trace-panic")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q", got)
	}

	body := decodeBody(t, rec)
	if body["detail"] != i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalPanic) {
		t.Fatalf("detail = %v, want panic recovery message", body["detail"])
	}
	if body["traceId"] != "trace-panic" {
		t.Fatalf("traceId = %v, want trace-panic", body["traceId"])
	}
	if _, exists := body["stackTrace"]; !exists {
		t.Fatal("development panic response must expose stackTrace")
	}
}

func TestGlobalErrorHandlerPanicHidesStackInProduction(t *testing.T) {
	engine := newTestEngine("production", func(*gin.Context) {
		panic("boom")
	})

	rec := performRequest(t, engine, "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := decodeBody(t, rec)
	if _, exists := body["stackTrace"]; exists {
		t.Fatal("production panic response must not expose stackTrace")
	}
}

func TestGlobalErrorHandlerLocalizesByRequestLocale(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.ContextWithFallback = true
	engine.Use(GlobalErrorHandler(stubLogger{}, "production"))
	engine.Use(NewLocaleMiddleware(&config.Config{}).Handle())
	engine.GET("/test", func(c *gin.Context) {
		_ = c.Error(apperror.NewNotFound(i18n.ErrorSearchIndexNotFound))
		c.Abort()
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Language", "en")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Language"); got != string(i18n.EN) {
		t.Fatalf("Content-Language = %q, want en", got)
	}

	body := decodeBody(t, rec)
	if body["title"] != i18n.Default().T(i18n.EN, i18n.ErrorTitleNotFound) {
		t.Fatalf("title = %v", body["title"])
	}
	if body["detail"] != i18n.Default().T(i18n.EN, i18n.ErrorSearchIndexNotFound) {
		t.Fatalf("detail = %v", body["detail"])
	}
}

func TestGlobalErrorHandlerDoesNotOverwriteWrittenResponse(t *testing.T) {
	engine := newTestEngine("development", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		_ = c.Error(apperror.NewBadRequest("late error"))
		c.Abort()
	})

	rec := performRequest(t, engine, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want the already-written 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "late error") {
		t.Fatalf("written response must not be overwritten: %s", rec.Body.String())
	}
}
