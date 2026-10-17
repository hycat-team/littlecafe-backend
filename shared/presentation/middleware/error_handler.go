package middleware

import (
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
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
					Title:    i18n.T(c.Request.Context(), i18n.ErrorTitleInternal),
					Detail:   i18n.T(c.Request.Context(), i18n.ErrorInternalPanic),
					Message:  i18n.T(c.Request.Context(), i18n.ErrorInternalPanic),
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

		localized := appErr.Localize(i18n.FromContext(c.Request.Context()))

		fields := []zap.Field{
			zap.Int("status", localized.Status),
			zap.String("title", localized.Title),
			zap.String("message", localized.Message),
			zap.String("path", c.Request.URL.Path),
		}
		if localized.MessageKey != "" {
			fields = append(fields, zap.String("messageKey", localized.MessageKey))
		}
		if rawErr != nil && rawErr.Error() != localized.Message {
			fields = append(fields, zap.String("error", rawErr.Error()))
		}
		if cause := localized.Unwrap(); cause != nil {
			if cause.Error() != localized.Message {
				fields = append(fields, zap.String("cause", cause.Error()))
			}
		}
		if source := errorutils.PrimaryStackFrame(localized.Stack()); source != "" {
			fields = append(fields, zap.String("source", source))
		}

		log.Error("API Execution Error Logged", fields...)

		if c.Writer.Written() {
			c.Abort()
			return
		}

		res := *localized
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
			if rawErr != nil && rawErr.Error() != localized.Message {
				res.DebugMessage = rawErr.Error()
			}
			res.StackTrace = errorutils.FilterStackTraceArray(localized.Stack())
		}

		c.Header("Content-Type", "application/problem+json")
		c.JSON(localized.Status, res)
		c.Errors = nil
		c.Abort()
	}
}
