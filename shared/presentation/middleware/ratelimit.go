package middleware

import (
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	shared_caching "github.com/hycat-team/littlecafe-backend/shared/application/interface/caching"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware is the coarse per-IP shield. It uses a Redis token bucket
// (GCRA) so throttling is smooth and exact across instances sharing the same
// Redis. A backend error fails open to preserve availability; the Nginx
// limit_req zone remains the outer flood shield.
//
// Business quotas ("N times per period" for a subject) use IRateLimiter with a
// fixed window instead — see the login/OTP middleware in the Identity service.
type RateLimitMiddleware struct {
	limiter      shared_caching.ITokenBucketLimiter
	burst        int
	refillPerSec float64
	keyPrefix    string
}

// NewRateLimitMiddleware builds the general API rule keyed by client IP.
// token_limit is the burst size; tokens_per_period / replenishment_seconds is
// the sustained refill rate.
func NewRateLimitMiddleware(cfg *config.Config, limiter shared_caching.ITokenBucketLimiter) *RateLimitMiddleware {
	refillPerSec := 0.0
	if cfg.RateLimit.ReplenishmentSeconds > 0 {
		refillPerSec = float64(cfg.RateLimit.TokensPerPeriod) / float64(cfg.RateLimit.ReplenishmentSeconds)
	}
	return &RateLimitMiddleware{
		limiter:      limiter,
		burst:        cfg.RateLimit.TokenLimit,
		refillPerSec: refillPerSec,
		keyPrefix:    "rl:api:",
	}
}

// Handle enforces the general API rule.
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := m.limiter.Allow(c.Request.Context(), m.keyPrefix+c.ClientIP(), m.burst, m.refillPerSec)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			c.Error(apperror.NewTooManyRequest(i18n.ErrorRateLimitExceeded))
			c.Abort()
			return
		}
		c.Next()
	}
}
