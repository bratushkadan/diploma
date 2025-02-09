package domain

import (
	"time"
)

type TokenType string

var (
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeAccess  TokenType = "access"
)

type RefreshToken struct {
	Id        string    `json:"id"`
	SubjectId string    `json:"subject_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AccessToken struct {
	SubjectId   string    `json:"subject_id"`
	SubjectType string    `json:"subject_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
