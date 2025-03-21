package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidConfirmationToken = errors.New("invalid confirmation token")
	ErrConfirmationTokenExpired = errors.New("confirmation token expired")
)

type EmailConfirmationRecord struct {
	Email     string    `dynamodbav:"email" json:"email"`
	Token     string    `dynamodbav:"token" json:"token"`
	ExpiresAt time.Time `dynamodbav:"expires_at" json:"expires_at"`
}

type EmailConfirmationTokens interface {
	InsertToken(ctx context.Context, email, token string) error
	ListTokensEmail(context context.Context, email string) ([]EmailConfirmationRecord, error)
	FindTokenRecord(context context.Context, token string) (*EmailConfirmationRecord, error)
}

type EmailConfirmationSender interface {
	Send(context.Context, EmailConfirmationSenderSendDTOInput) error
}

type EmailConfirmationSenderSendDTOInput struct {
	RecipientEmail    string
	ConfirmationToken string
}
