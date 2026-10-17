package apperror

import (
	"net/http"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func NewError(status int, titleKey, detailKey string, args ...any) *Error {
	return New(status, titleKey, detailKey, args...)
}

func NewBadRequest(messageKey string, args ...any) *Error {
	return New(http.StatusBadRequest, i18n.ErrorTitleBadRequest, messageKey, args...)
}

func NewNotFound(messageKey string, args ...any) *Error {
	return New(http.StatusNotFound, i18n.ErrorTitleNotFound, messageKey, args...)
}

func NewTooManyRequest(messageKey string, args ...any) *Error {
	return New(http.StatusTooManyRequests, i18n.ErrorTitleTooManyRequests, messageKey, args...)
}

func NewRequestEntityTooLarge(messageKey string, args ...any) *Error {
	return New(http.StatusRequestEntityTooLarge, i18n.ErrorTitleRequestEntityTooLarge, messageKey, args...)
}

func NewConflict(messageKey string, args ...any) *Error {
	return New(http.StatusConflict, i18n.ErrorTitleConflict, messageKey, args...)
}

func NewUnauthorized(messageKey string, args ...any) *Error {
	return New(http.StatusUnauthorized, i18n.ErrorTitleUnauthorized, messageKey, args...)
}

func NewForbidden(messageKey string, args ...any) *Error {
	return New(http.StatusForbidden, i18n.ErrorTitleForbidden, messageKey, args...)
}

func NewInternalError(messageKey string, args ...any) *Error {
	return New(http.StatusInternalServerError, i18n.ErrorTitleInternal, messageKey, args...)
}

func NewServiceUnavailable(messageKey string, args ...any) *Error {
	return New(http.StatusServiceUnavailable, i18n.ErrorTitleServiceUnavailable, messageKey, args...)
}

// NewUpstreamUnavailable maps a third-party upstream failure to a client-safe
// 503; the raw provider error travels only as Cause (never serialized) so
// server logs keep the detail.
func NewUpstreamUnavailable(cause error) *Error {
	e := NewServiceUnavailable(i18n.ErrorUpstreamUnavailable)
	e.Cause = cause
	return e
}
