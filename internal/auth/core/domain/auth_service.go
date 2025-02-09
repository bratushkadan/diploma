package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSendAccountConfirmationFailed = errors.New("failed to send account confirmation email")
	ErrInvalidCredentials            = errors.New("invalid credentials")

	ErrInvalidRefreshToken           = errors.New("invalid refresh token")
	ErrInvalidAccessToken            = errors.New("invalid access token")
	ErrInvalidTokenType              = errors.New("invalid token type")
	ErrTokenParseFailed              = errors.New("token parse failed")
	ErrTokenExpired                  = errors.New("token expired")
	ErrTokenRevoked                  = errors.New("token revoked")
	ErrRefreshTokenToReplaceNotFound = errors.New("refresh token to replace not found")
)

type AuthService interface {
	CreateUser(context.Context, CreateUserReq) (CreateUserRes, error)
	CreateSeller(context.Context, CreateSellerReq) (CreateSellerRes, error)
	// DO NOT expose this method externally.
	CreateAdmin(context.Context, CreateAdminReq) (CreateAdminRes, error)
	ActivateAccounts(context.Context, ActivateAccountsReq) (ActivateAccountsRes, error)

	Authenticate(context.Context, AuthenticateReq) (AuthenticateRes, error)
	ReplaceRefreshToken(context.Context, ReplaceRefreshTokenReq) (ReplaceRefreshTokenRes, error)

	CreateAccessToken(context.Context, CreateAccessTokenReq) (CreateAccessTokenRes, error)
}

type CreateUserReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type CreateUserRes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateSellerReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	// Access token that belongs to the admin.
	AccessToken string `json:"access_token"`
}
type CreateSellerRes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateAdminReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type CreateAdminRes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
