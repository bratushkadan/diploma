package domain

import (
	"context"
	"time"
)

type EmailConfirmationRecord struct {
	Email     string    `dynamodbav:"email" json:"email"`
	Token     string    `dynamodbav:"token" json:"token"`
	ExpiresAt time.Time `dynamodbav:"expires_at" json:"expires_at"`
}

type EmailConfirmatorTokenRepo interface {
	InsertToken(ctx context.Context, email, token string) error
	ListTokensEmail(context context.Context, email string) ([]EmailConfirmationRecord, error)
	FindTokenRecord(context context.Context, token string) (*EmailConfirmationRecord, error)
}
