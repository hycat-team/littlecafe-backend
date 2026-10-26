package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	shared_caching "github.com/hycat-team/littlecafe-backend/shared/application/interface/caching"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"

	"github.com/gin-gonic/gin"
)

type fakeTokenBucket struct {
	state shared_caching.TokenBucketState
	err   error
}

func (f fakeTokenBucket) Allow(context.Context, string, int, float64) (shared_caching.TokenBucketState, error) {
	return f.state, f.err
}

func newRateLimitTestRouter(limiter shared_caching.ITokenBucketLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{RateLimit: config.RateLimit{
		TokenLimit:           10,
		TokensPerPeriod:      10,
		ReplenishmentSeconds: 60,
	}}
	mw := NewRateLimitMiddleware(cfg, limiter)

	r := gin.New()
	r.Use(GlobalErrorHandler(logger.New("test", "", "error", false), "production"))
	r.Use(mw.Handle())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRateLimitSetsHeadersWhenAllowed(t *testing.T) {
	reset := time.Unix(1_760_000_000, 0)
	r := newRateLimitTestRouter(fakeTokenBucket{
		state: shared_caching.TokenBucketState{Allowed: true, Remaining: 9, Reset: reset},
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("X-RateLimit-Limit"); got != "10" {
		t.Fatalf("limit = %q", got)
	}
	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "9" {
		t.Fatalf("remaining = %q", got)
	}
	if got := rec.Header().Get("X-RateLimit-Reset"); got != "1760000000" {
		t.Fatalf("reset = %q", got)
	}
	if got := rec.Header().Get("Retry-After"); got != "" {
		t.Fatalf("retry-after = %q, want empty", got)
	}
}

func TestRateLimitSetsRetryAfterWhenDenied(t *testing.T) {
	r := newRateLimitTestRouter(fakeTokenBucket{
		state: shared_caching.TokenBucketState{
			Allowed:    false,
			Remaining:  0,
			Reset:      time.Unix(1_760_000_030, 0),
			RetryAfter: 1500 * time.Millisecond,
		},
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "2" {
		t.Fatalf("retry-after = %q, want 2", got)
	}
	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("remaining = %q", got)
	}
}

func TestRateLimitFailsOpenOnLimiterError(t *testing.T) {
	r := newRateLimitTestRouter(fakeTokenBucket{err: context.DeadlineExceeded})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Fatalf("limit = %q, want empty", got)
	}
}
