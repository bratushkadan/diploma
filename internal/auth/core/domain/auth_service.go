package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSendAccountConfirmationFailed = errors.New("failed to send account confirmation email")
	ErrInvalidCredentials            = errors.New("invalid credentials")
)

type AuthServiceV2 interface {
	CreateAccount(context.Context, CreateAccountReq) (CreateAccountRes, error)
	ActivateAccounts(context.Context, ActivateAccountsReq) (ActivateAccountsRes, error)

	Authenticate(context.Context, AuthenticateReq) (AuthenticateRes, error)
	ReplaceRefreshToken(context.Context, ReplaceRefreshTokenReq) (ReplaceRefreshTokenRes, error)
}

type CreateAccountReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Type     string `json:"type"`
}
type CreateAccountRes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

type ActivateAccountsReq struct {
	Emails []string `json:"emails"`
}
type ActivateAccountsRes struct {
}

type AuthenticateReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthenticateRes struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    time.Time
}

type ReplaceRefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}
type ReplaceRefreshTokenRes struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    time.Time
}
