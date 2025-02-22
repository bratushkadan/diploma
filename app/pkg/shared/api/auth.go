package api

import (
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenJwtClaims struct {
	TokenType   domain.TokenType `json:"token_type"`
	SubjectId   string           `json:"subject_id"`
	SubjectType string           `json:"subject_type"`
	jwt.RegisteredClaims
}
