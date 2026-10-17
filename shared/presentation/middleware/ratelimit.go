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

// LoginRateLimitMiddleware is the stricter login/OTP shield keyed by client IP
// plus the attempted account identifier.
type LoginRateLimitMiddleware struct {
	limiter       shared_caching.IRateLimiter
	limit         int
	windowSeconds int
}

// NewLoginRateLimitMiddleware builds the login/OTP rule.
func NewLoginRateLimitMiddleware(cfg *config.Config, limiter shared_caching.IRateLimiter) *LoginRateLimitMiddleware {
	return &LoginRateLimitMiddleware{
		limiter:       limiter,
		limit:         cfg.RateLimit.TokenLimit,
		windowSeconds: cfg.RateLimit.ReplenishmentSeconds,
	}
}

// Handle enforces the login rule.
func (m *LoginRateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rl:login:" + c.ClientIP() + ":" + loginIdentifier(c)
		allowed, err := m.limiter.Allow(c.Request.Context(), key, m.limit, m.windowSeconds)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			c.Error(apperror.NewTooManyRequest("Bạn đã thử quá nhiều lần. Vui lòng thử lại sau ít phút."))
			c.Abort()
			return
		}
		c.Next()
	}
}

// loginIdentifier binds the rule to the attempting account as well as the IP.
func loginIdentifier(c *gin.Context) string {
	for _, field := range []string{"username", "email", "phone", "identifier", "userId"} {
		if v := c.Query(field); v != "" {
			return v
		}
		if v := c.PostForm(field); v != "" {
			return v
		}
	}
	if auth := c.GetHeader("Authorization"); len(auth) > 20 {
		return auth[len(auth)-12:]
	}
	return "anonymous"
}
