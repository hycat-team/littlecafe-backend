package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var corsTestOrigins = []string{
	"https://littlecafe.kids",
	"https://www.littlecafe.kids",
	" http://localhost:3000 ",
}

func newCORSTestRouter(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware(origins))
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func performCORSRequest(r *gin.Engine, method, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/health", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCORSAllowedOriginIsEchoed(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodGet, "https://littlecafe.kids")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://littlecafe.kids" {
		t.Fatalf("allow origin = %q", got)
	}
	if !strings.Contains(w.Header().Get("Vary"), "Origin") {
		t.Fatalf("Vary = %q, want Origin", w.Header().Get("Vary"))
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q", got)
	}
}

func TestCORSSecondOriginDespiteSpaces(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodGet, "http://localhost:3000")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSTrailingSlashIsNormalized(t *testing.T) {
	r := newCORSTestRouter([]string{"https://littlecafe.kids/"})
	w := performCORSRequest(r, http.MethodGet, "https://littlecafe.kids")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://littlecafe.kids" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSUnknownOriginGetsNoHeaders(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodGet, "https://evil.example.com")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q, want empty", got)
	}
}

func TestCORSNoOriginHeaderSetsNothing(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodGet, "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q, want empty", got)
	}
}

func TestCORSPreflightAllowed(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodOptions, "https://www.littlecafe.kids")

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://www.littlecafe.kids" {
		t.Fatalf("allow origin = %q", got)
	}
	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key") {
		t.Fatalf("allow headers = %q", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if !strings.Contains(w.Header().Get("Access-Control-Expose-Headers"), "X-Trace-ID") {
		t.Fatalf("expose headers = %q", w.Header().Get("Access-Control-Expose-Headers"))
	}
}

func TestCORSPreflightDenied(t *testing.T) {
	r := newCORSTestRouter(corsTestOrigins)
	w := performCORSRequest(r, http.MethodOptions, "https://evil.example.com")

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q, want empty", got)
	}
}

func TestCORSWildcardEchoesOrigin(t *testing.T) {
	r := newCORSTestRouter([]string{"*"})
	w := performCORSRequest(r, http.MethodGet, "https://preview.example.com")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://preview.example.com" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSToleratesEmptyEntries(t *testing.T) {
	r := newCORSTestRouter([]string{"", " ", "https://littlecafe.kids"})
	w := performCORSRequest(r, http.MethodGet, "https://littlecafe.kids")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://littlecafe.kids" {
		t.Fatalf("allow origin = %q", got)
	}
}
