package email_confirmation_daemon_adapter

import "github.com/bratushkadan/floral/internal/auth/core/domain"

type EmailConfirmation struct {
	svc domain.AuthService
}
