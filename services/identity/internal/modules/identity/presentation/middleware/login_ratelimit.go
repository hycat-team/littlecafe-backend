// Package middleware holds Identity presentation middleware.
package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	shared_caching "github.com/hycat-team/littlecafe-backend/shared/application/interface/caching"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

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
			c.Error(apperror.NewTooManyRequest(i18n.ErrorLoginRateLimit))
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
