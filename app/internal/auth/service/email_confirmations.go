package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/entity"

	"go.uber.org/zap"
)

var (
	ErrInvalidConfirmationToken = errors.New("invalid confirmation token")
	ErrConfirmationTokenExpired = errors.New("confirmation token expired")
)

type EmailConfirmationBuilder struct {
	YdbDocApiAwsAccessKeyId     string
	YdbDocApiAwsSecretAccessKey string
	YdbDocApiEndpoint           string

	SqsAwsAccessKeyId     string
	SqsAwsSecretAccessKey string
	SqsQueueUrl           string

	SenderEmail                          string
	SenderPassword                       string
	EmailConfirmationApiEndpointResolver email_confirmer.ConfirmationUrlResolver

	emailConfirmationSendTimeout *time.Duration

	ec *EmailConfirmation
}

func NewEmailConfirmationBuilder() *EmailConfirmationBuilder {
	return &EmailConfirmationBuilder{
		ec: &EmailConfirmation{},
	}
}

func (b *EmailConfirmationBuilder) Tokens(a domain.EmailConfirmationTokens) *EmailConfirmationBuilder {
	b.ec.confirmationTokens = a
	return b
}
func (b *EmailConfirmationBuilder) Sender(a domain.EmailConfirmationSender) *EmailConfirmationBuilder {
	b.ec.confirmationSender = a
	return b
}
func (b *EmailConfirmationBuilder) Notifications(a domain.EmailConfirmationNotifications) *EmailConfirmationBuilder {
	b.ec.emailConfirmationNotifications = a
	return b
}
func (b *EmailConfirmationBuilder) Logger(l *zap.Logger) *EmailConfirmationBuilder {
	b.ec.l = l
	return b
}

func (b *EmailConfirmationBuilder) Build() (*EmailConfirmation, error) {
	if b.ec.l == nil {
		b.ec.l = zap.NewNop()
	}

	return b.ec, nil
}

type EmailConfirmation struct {
	confirmationTokens             domain.EmailConfirmationTokens
	emailConfirmationNotifications domain.EmailConfirmationNotifications
	confirmationSender             domain.EmailConfirmationSender

	l *zap.Logger
}

var _ domain.AccountEmailConfirmation = (*EmailConfirmation)(nil)

func (c *EmailConfirmation) Confirm(ctx context.Context, token string) error {
	c.l.Info("confirm email")
	c.l.Info("retrieve confirmation token records")
	record, err := c.confirmationTokens.FindTokenRecord(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to retrieve tokens: %v", err)
	}
	if record == nil {
		c.l.Info("invalid email confirmation token record")
		return ErrInvalidConfirmationToken
	}
	c.l.Info("retrieved email confirmation token record", zap.String("email", record.Email))
	if time.Now().After(record.ExpiresAt) {
		return ErrConfirmationTokenExpired
	}
	c.l.Info("validated email confirmation token record", zap.String("email", record.Email))

	if _, err := c.emailConfirmationNotifications.Send(ctx, domain.SendEmailConfirmationNotificationsDTOInput{Email: record.Email}); err != nil {
		return fmt.Errorf("failed to produce email confirmation message: %v", err)
	}
	c.l.Info("produced email confirmation message", zap.String("email", record.Email))

	c.l.Info("confirmed email", zap.String("email", record.Email))
	return nil
}

func (c *EmailConfirmation) Send(ctx context.Context, email string) error {
	c.l.Info("create confirmation token and send email", zap.String("email", email))
	tokenString := entity.Id(64)
	err := c.confirmationTokens.InsertToken(ctx, email, tokenString)
	if err != nil {
		return fmt.Errorf("failed to insert confirmation token: %v", err)
	}
	c.l.Info("inserted confirmation token", zap.String("email", email))

	if err := c.confirmationSender.Send(ctx, domain.EmailConfirmationSenderSendDTOInput{
		RecipientEmail:    email,
		ConfirmationToken: tokenString,
	}); err != nil {
		return fmt.Errorf("failed to send confirmation email: %v", err)
	}
	c.l.Info("sent confirmation email")

	return nil
}
