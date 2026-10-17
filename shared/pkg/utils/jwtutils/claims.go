package jwtutils

import (
	"crypto/ecdsa"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/application/constants/jwttype"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

type CustomClaims struct {
	Email    string          `json:"email"`
	RoleSlug string          `json:"role_slug"`
	Type     jwttype.JwtType `json:"type"`
	jwt.RegisteredClaims
}

func ValidateTokenES256(tokenString string, expectedTokenType jwttype.JwtType, publicKey *ecdsa.PublicKey) (*CustomClaims, error) {
	if publicKey == nil {
		return nil, apperror.NewInternalError(i18n.ErrorJWTMissingPublicKey)
	}
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, apperror.ErrUnexpectedSigningToken()
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid || claims.Type != expectedTokenType {
		return nil, apperror.ErrInvalidAccessToken()
	}
	return claims, nil
}
