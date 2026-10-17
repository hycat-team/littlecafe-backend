package apperror

import "github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"

func ErrInvalidToken() *Error {
	return NewUnauthorized(i18n.ErrorAuthInvalidToken)
}

func ErrInvalidAccessToken() *Error {
	return NewUnauthorized(i18n.ErrorAuthSessionExpired)
}

func ErrUnauthorized() *Error {
	return NewUnauthorized(i18n.ErrorAuthRequired)
}

func ErrForbidden() *Error {
	return NewForbidden(i18n.ErrorForbiddenAccess)
}

func ErrSearchIndexNotFound() *Error {
	return NewNotFound(i18n.ErrorSearchIndexNotFound)
}

func ErrUnexpectedSigningToken() *Error {
	return NewInternalError(i18n.ErrorAuthInvalidToken)
}
