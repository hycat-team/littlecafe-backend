package middleware

import (
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	shared_caching "github.com/hycat-team/littlecafe-backend/shared/application/interface/caching"
	"github.com/hycat-team/littlecafe-backend/shared/config"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware is a Redis-backed fixed-window rate limiter. It is
// accurate across instances because all replicas share the same IRateLimiter
// backend. A backend error fails open to preserve availability.
type RateLimitMiddleware struct {
	limiter       shared_caching.IRateLimiter
	limit         int
	windowSeconds int
	keyPrefix     string
}

// NewRateLimitMiddleware builds the general API rule keyed by client IP.
func NewRateLimitMiddleware(cfg *config.Config, limiter shared_caching.IRateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:       limiter,
		limit:         cfg.RateLimit.TokenLimit,
		windowSeconds: cfg.RateLimit.ReplenishmentSeconds,
		keyPrefix:     "rl:api:",
	}
}

// Handle enforces the general API rule.
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := m.limiter.Allow(c.Request.Context(), m.keyPrefix+c.ClientIP(), m.limit, m.windowSeconds)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			c.Error(apperror.NewTooManyRequest("Bạn đang thao tác quá nhanh. Vui lòng thử lại sau ít phút."))
			c.Abort()
			return
		}
		c.Next()
	}
}
