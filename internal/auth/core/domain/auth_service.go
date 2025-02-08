package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSendAccountConfirmationFailed = errors.New("failed to send account confirmation email")
	ErrInvalidCredentials            = errors.New("invalid credentials")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidTokenType    = errors.New("invalid token type")
	ErrTokenParseFailed    = errors.New("token parse failed")
	ErrTokenExpired        = errors.New("token expired")
)

type AuthService interface {
	CreateAccount(context.Context, CreateAccountReq) (CreateAccountRes, error)
	ActivateAccounts(context.Context, ActivateAccountsReq) (ActivateAccountsRes, error)

	Authenticate(context.Context, AuthenticateReq) (AuthenticateRes, error)
	ReplaceRefreshToken(context.Context, ReplaceRefreshTokenReq) (ReplaceRefreshTokenRes, error)

	CreateAccessToken(context.Context, CreateAccessTokenReq) (CreateAccessTokenRes, error)
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

type CreateAccessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}
type CreateAccessTokenRes struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
