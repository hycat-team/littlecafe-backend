package middleware

import (
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/utils/errorutils"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GlobalErrorHandler(log logger.Interface, appEnv string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rawStack := string(debug.Stack())

				log.Error("CRITICAL PANIC RECOVERED",
					zap.Any("error", r),
					zap.String("source", errorutils.PrimaryStackFrame(rawStack)),
					zap.String("path", c.Request.URL.Path),
				)

				if c.Writer.Written() {
					c.Abort()
					return
				}

				res := apperror.Error{
					Type:     "about:blank",
					Status:   http.StatusInternalServerError,
					Title:    "Lỗi hệ thống",
					Detail:   apperror.PanicRecoveryMessage,
					Message:  apperror.PanicRecoveryMessage,
					Instance: c.Request.URL.Path,
					TraceID:  c.GetHeader("X-Trace-ID"),
				}

				// Debug details are gated strictly on the configured env, not
				// gin.Mode() — staging/unset envs must not leak internals.
				if !config.IsProductionEnv(appEnv) {
					res.StackTrace = errorutils.FilterStackTraceArray(rawStack)
				}

				c.Header("Content-Type", "application/problem+json")
				c.JSON(http.StatusInternalServerError, res)
				c.Errors = nil
				c.Abort()
			}
		}()

		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		rawErr := c.Errors.Last().Err
		appErr := errorutils.ToAppError(rawErr)
		if appErr == nil {
			appErr = apperror.NewInternalError(apperror.GenericInternalMessage)
		}

		fields := []zap.Field{
			zap.Int("status", appErr.Status),
			zap.String("title", appErr.Title),
			zap.String("message", appErr.Message),
			zap.String("path", c.Request.URL.Path),
		}
		if rawErr != nil && rawErr.Error() != appErr.Message {
			fields = append(fields, zap.String("error", rawErr.Error()))
		}
		if cause := appErr.Unwrap(); cause != nil {
			if cause.Error() != appErr.Message {
				fields = append(fields, zap.String("cause", cause.Error()))
			}
		}
		if source := errorutils.PrimaryStackFrame(appErr.Stack()); source != "" {
			fields = append(fields, zap.String("source", source))
		}

		log.Error("API Execution Error Logged", fields...)

		if c.Writer.Written() {
			c.Abort()
			return
		}

		res := *appErr
		if res.Type == "" {
			res.Type = "about:blank"
		}
		if res.Detail == "" {
			res.Detail = res.Message
		}
		res.Instance = c.Request.URL.Path
		res.TraceID = c.GetHeader("X-Trace-ID")
		res.StackTrace = nil
		res.DebugMessage = ""
		// Debug details are gated strictly on the configured env, not
		// gin.Mode() — staging/unset envs must not leak internals.
		if !config.IsProductionEnv(appEnv) {
			if rawErr != nil && rawErr.Error() != appErr.Message {
				res.DebugMessage = rawErr.Error()
			}
			res.StackTrace = errorutils.FilterStackTraceArray(appErr.Stack())
		}

		c.Header("Content-Type", "application/problem+json")
		c.JSON(appErr.Status, res)
		c.Errors = nil
		c.Abort()
	}
}
