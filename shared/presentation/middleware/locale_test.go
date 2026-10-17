package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func newLocaleTestEngine(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.ContextWithFallback = true
	engine.Use(NewLocaleMiddleware(cfg).Handle())
	engine.GET("/locale", func(c *gin.Context) {
		c.String(http.StatusOK, string(i18n.FromContext(c)))
	})
	return engine
}

func performLocaleRequest(t *testing.T, engine *gin.Engine, acceptLanguage string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/locale", nil)
	if acceptLanguage != "" {
		req.Header.Set(AcceptLanguageHeader, acceptLanguage)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestLocaleMiddlewareNegotiates(t *testing.T) {
	engine := newLocaleTestEngine(&config.Config{})

	tests := []struct {
		header     string
		wantLocale string
	}{
		{header: "", wantLocale: "vi"},
		{header: "vi-VN", wantLocale: "vi"},
		{header: "en-US,en;q=0.9", wantLocale: "en"},
		{header: "fr-FR", wantLocale: "vi"},
	}

	for _, tc := range tests {
		rec := performLocaleRequest(t, engine, tc.header)
		if rec.Body.String() != tc.wantLocale {
			t.Fatalf("Accept-Language %q resolved to %q, want %q", tc.header, rec.Body.String(), tc.wantLocale)
		}
		if got := rec.Header().Get("Content-Language"); got != tc.wantLocale {
			t.Fatalf("Content-Language = %q, want %q", got, tc.wantLocale)
		}
		if got := rec.Header().Get("Vary"); got != AcceptLanguageHeader {
			t.Fatalf("Vary = %q, want %q", got, AcceptLanguageHeader)
		}
	}
}

func TestLocaleMiddlewareHonorsConfiguredDefault(t *testing.T) {
	cfg := &config.Config{}
	cfg.I18n.DefaultLocale = "en"
	cfg.I18n.SupportedLocales = "vi,en"

	engine := newLocaleTestEngine(cfg)

	rec := performLocaleRequest(t, engine, "")
	if rec.Body.String() != "en" {
		t.Fatalf("default locale = %q, want en", rec.Body.String())
	}
}
