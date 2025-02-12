package domain

import "context"

type AccountEmailConfirmation interface {
	Confirm(ctx context.Context, token string) error
	Send(ctx context.Context, email string) error
}
