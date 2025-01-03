package domain

import "time"

type RefreshToken struct {
	TokenType string // 'refresh'
	SubjectId string
	ExpiresAt time.Time
}

// Short-lived token that expires after 5 minutes
type AccessToken struct {
	TokenType   string // 'access'
	SubjectId   string
	SubjectType string
	ExpiresAt   time.Time
}

type AuthService struct {
}

type RefreshTokenProvider interface {
	Create(subjectId string) (*RefreshToken, error)
	Decode(token string) (*RefreshToken, error)
}

type AccessTokenProvider interface {
	Create(subjectId string, subjectType string) (string, error)
	Decode(token string) (*AccessToken, error)
}
