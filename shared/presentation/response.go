package presentation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type HandlerFuncErr func(*gin.Context) error

func WrapHandler(fn HandlerFuncErr) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			c.Error(err)
			c.Abort()
		}
	}
}

func Success(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Message: message,
		Data:    data,
	})
}

func Accepted(c *gin.Context, message string, data any) {
	c.JSON(http.StatusAccepted, APIResponse{
		Message: message,
		Data:    data,
	})
}

// NoContent renders a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// CreatedWithLocation renders a 201 response and sets the Location header.
func CreatedWithLocation(c *gin.Context, location, message string, data any) {
	if location != "" {
		c.Header("Location", location)
	}
	Created(c, message, data)
}

// AcceptedWithLocation renders a 202 response and sets the Location header.
func AcceptedWithLocation(c *gin.Context, location, message string, data any) {
	if location != "" {
		c.Header("Location", location)
	}
	Accepted(c, message, data)
}

// Message resolves a catalog key for the request locale.
func Message(c *gin.Context, key string, args ...any) string {
	return i18n.T(c.Request.Context(), key, args...)
}

// LocalizedSuccess renders Success with a message resolved from a catalog key.
func LocalizedSuccess(c *gin.Context, messageKey string, data any) {
	Success(c, Message(c, messageKey), data)
}

// LocalizedCreated renders Created with a message resolved from a catalog key.
func LocalizedCreated(c *gin.Context, messageKey string, data any) {
	Created(c, Message(c, messageKey), data)
}

// LocalizedCreatedWithLocation renders CreatedWithLocation with a resolved message.
func LocalizedCreatedWithLocation(c *gin.Context, location, messageKey string, data any) {
	CreatedWithLocation(c, location, Message(c, messageKey), data)
}

// LocalizedAccepted renders Accepted with a message resolved from a catalog key.
func LocalizedAccepted(c *gin.Context, messageKey string, data any) {
	Accepted(c, Message(c, messageKey), data)
}

// LocalizedAcceptedWithLocation renders AcceptedWithLocation with a resolved message.
func LocalizedAcceptedWithLocation(c *gin.Context, location, messageKey string, data any) {
	AcceptedWithLocation(c, location, Message(c, messageKey), data)
}
