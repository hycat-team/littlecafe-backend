package contextutils

import (
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/domain/constants/roleslug"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserId returns the authenticated user id stored on the context.
func GetUserId(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(CtxUserId)
	if !exists {
		return uuid.Nil, apperror.ErrUnauthorized()
	}

	uid, ok := val.(uuid.UUID)
	if !ok {
		strVal, ok := val.(string)
		if !ok {
			return uuid.Nil, apperror.ErrUnauthorized()
		}
		parsed, err := uuid.Parse(strVal)
		if err != nil {
			return uuid.Nil, apperror.ErrUnauthorized()
		}
		return parsed, nil
	}

	return uid, nil
}

// GetUserIdOptional returns uuid.Nil when no authenticated user is present.
func GetUserIdOptional(c *gin.Context) uuid.UUID {
	val, exists := c.Get(CtxUserId)
	if !exists {
		return uuid.Nil
	}
	switch v := val.(type) {
	case uuid.UUID:
		return v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil
		}
		return parsed
	default:
		return uuid.Nil
	}
}

// GetRoleSlug returns the verified role stored on the context.
func GetRoleSlug(c *gin.Context) (roleslug.RoleSlug, error) {
	val, exists := c.Get(CtxRoleSlug)
	if !exists {
		return "", apperror.ErrUnauthorized()
	}

	role, ok := val.(string)
	if !ok {
		return "", apperror.ErrUnauthorized()
	}

	return roleslug.RoleSlug(role), nil
}

// GetAccessToken returns the raw access token stored on the context.
func GetAccessToken(c *gin.Context) (string, error) {
	val, exists := c.Get(CtxAccessToken)
	if !exists {
		return "", apperror.ErrUnauthorized()
	}

	token, ok := val.(string)
	if !ok {
		return "", apperror.ErrUnauthorized()
	}

	return token, nil
}
