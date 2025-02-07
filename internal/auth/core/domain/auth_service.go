package domain

import "context"

type AuthServiceV2 interface {
	CreateAccount(context.Context, CreateAccountReq) (CreateAccountRes, error)
	// ActivateAccounts(context.Context, ActivateAccountsReq) (ActivateAccountsRes, error)

	// Authenticate(context.Context, AuthenticateReq) (AuthenticateRes, error)
	// RenewRefreshToken(context.Context, RenewRefreshTokenReq) (RenewRefreshTokenRes, error)
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
}

type RenewRefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}
type RenewRefreshTokenRes struct {
	RefreshToken string `json:"refresh_token"`
}
