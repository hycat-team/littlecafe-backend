package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowHeaders  = "Authorization, Content-Type, Accept, Accept-Language, Idempotency-Key, X-Trace-ID"
	corsAllowMethods  = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsExposeHeaders = "Content-Language, X-Trace-ID"
)

func normalizeOrigin(o string) string {
	return strings.TrimRight(strings.TrimSpace(o), "/")
}

// CORSMiddleware enforces CORS at the application layer. It echoes the exact
// request origin (never "*" when credentials are enabled), supports multiple
// origins and the "*" wildcard outside production, and answers preflight
// requests before routing.
func CORSMiddleware(allowOrigins []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool, len(allowOrigins))
	for _, o := range allowOrigins {
		if normalized := normalizeOrigin(o); normalized != "" {
			allowedMap[normalized] = true
		}
	}
	allowAll := allowedMap["*"]

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader("Origin"))
		allowed := origin != "" && (allowAll || allowedMap[origin])

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Add("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", corsAllowHeaders)
			c.Header("Access-Control-Allow-Methods", corsAllowMethods)
			c.Header("Access-Control-Expose-Headers", corsExposeHeaders)
			c.Header("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			if !allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
