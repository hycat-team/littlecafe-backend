package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/config"
)

type fakeLimiter struct {
	allowed bool
	err     error
	key     string
	limit   int
	window  int
}

func (f *fakeLimiter) Allow(_ context.Context, key string, limit int, windowSeconds int) (bool, error) {
	f.key = key
	f.limit = limit
	f.window = windowSeconds
	return f.allowed, f.err
}

func newMiddleware(limiter *fakeLimiter) *LoginRateLimitMiddleware {
	cfg := &config.Config{RateLimit: config.RateLimit{TokenLimit: 5, ReplenishmentSeconds: 60}}
	return NewLoginRateLimitMiddleware(cfg, limiter)
}

func runRequest(t *testing.T, limiter *fakeLimiter, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(newMiddleware(limiter).Handle())
	router.GET("/login", func(c *gin.Context) {
		c.Header("X-Handled", "true")
		c.Status(http.StatusNoContent)
	})
	router.POST("/login", func(c *gin.Context) {
		c.Header("X-Handled", "true")
		c.Status(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, request)
	return rec
}

func TestLoginRateLimitAllowsRequest(t *testing.T) {
	limiter := &fakeLimiter{allowed: true}
	rec := runRequest(t, limiter, httptest.NewRequest(http.MethodGet, "/login?username=alice", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if limiter.key != "rl:login:192.0.2.1:alice" || limiter.limit != 5 || limiter.window != 60 {
		t.Fatalf("limiter = %+v", limiter)
	}
}

func TestLoginRateLimitRejectsRequest(t *testing.T) {
	limiter := &fakeLimiter{allowed: false}
	rec := runRequest(t, limiter, httptest.NewRequest(http.MethodGet, "/login", nil))
	if rec.Header().Get("X-Handled") != "" {
		t.Fatal("handler must not run when the request is rejected")
	}
	if limiter.key == "" {
		t.Fatal("limiter must be consulted")
	}
}

func TestLoginRateLimitFailsOpen(t *testing.T) {
	limiter := &fakeLimiter{err: errors.New("redis down")}
	rec := runRequest(t, limiter, httptest.NewRequest(http.MethodGet, "/login", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLoginIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		query   string
		form    string
		auth    string
		want    string
		method  string
		hasBody bool
	}{
		{name: "username", query: "username=alice", want: "alice", method: http.MethodGet},
		{name: "email", query: "email=a@example.com", want: "a@example.com", method: http.MethodGet},
		{name: "phone", query: "phone=123", want: "123", method: http.MethodGet},
		{name: "identifier", query: "identifier=id-1", want: "id-1", method: http.MethodGet},
		{name: "userId", query: "userId=42", want: "42", method: http.MethodGet},
		{name: "form", form: "username=bob", want: "bob", method: http.MethodPost, hasBody: true},
		{name: "authorization", auth: "Bearer aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", want: "aaaaaaaaaaaa", method: http.MethodGet},
		{name: "anonymous", want: "anonymous", method: http.MethodGet},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := "/login"
			if tc.query != "" {
				target += "?" + tc.query
			}
			var request *http.Request
			if tc.hasBody {
				request = httptest.NewRequest(tc.method, target, strings.NewReader(tc.form))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				request = httptest.NewRequest(tc.method, target, nil)
			}
			if tc.auth != "" {
				request.Header.Set("Authorization", tc.auth)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = request
			if got := loginIdentifier(c); got != tc.want {
				t.Fatalf("loginIdentifier = %q, want %q", got, tc.want)
			}
		})
	}
}
