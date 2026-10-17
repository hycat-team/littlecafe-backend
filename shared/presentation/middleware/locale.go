package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

// AcceptLanguageHeader is the request header used to negotiate the client-facing
// language.
const AcceptLanguageHeader = "Accept-Language"

// LocaleMiddleware negotiates the request locale from Accept-Language and
// stores it on both the Gin context and the request context. Error and success
// messages are rendered by the presentation layer using that locale.
type LocaleMiddleware struct {
	negotiator *i18n.Negotiator
}

// NewLocaleMiddleware builds the middleware from the i18n configuration.
func NewLocaleMiddleware(cfg *config.Config) *LocaleMiddleware {
	// Loading the catalog here validates bundle parity at boot, before the first
	// request can hit a missing translation.
	i18n.Default()

	defaultLocale := i18n.DefaultLocale
	supported := i18n.SupportedLocales
	if cfg != nil {
		defaultLocale = cfg.I18n.Default()
		supported = cfg.I18n.Supported()
	}
	return &LocaleMiddleware{
		negotiator: i18n.NewNegotiator(defaultLocale, supported),
	}
}

// Handle resolves and propagates the request locale.
func (m *LocaleMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale, _ := m.negotiator.Negotiate(c.GetHeader(AcceptLanguageHeader))
		locale = i18n.Normalize(locale)

		c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), locale))

		c.Header("Content-Language", string(locale))
		c.Writer.Header().Add("Vary", AcceptLanguageHeader)

		c.Next()
	}
}
