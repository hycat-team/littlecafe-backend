package bootstrap

import (
	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
	"github.com/hycat-team/littlecafe-backend/shared/presentation"
	"github.com/hycat-team/littlecafe-backend/shared/presentation/middleware"
)

// ModuleRouter attaches one business module's routes to the API group.
type ModuleRouter interface {
	Register(api *gin.RouterGroup)
}

// NewRouter builds the HTTP engine with shared middleware, the health probe
// and every business module router.
func NewRouter(cfg *config.Config, log logger.Interface, modules []ModuleRouter) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.GlobalErrorHandler(log, cfg.App.Env))

	api := engine.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		presentation.Success(c, "ok", gin.H{"status": "healthy"})
	})
	for _, m := range modules {
		m.Register(api)
	}
	return engine
}
