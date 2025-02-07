package domain

import "context"

type UserRepo interface {
	CreateUser(context.Context, UserProviderCreateUserReq) (*User, error)
}

type CheckUserCredentialsDTOOutput struct {
	UserId   string
	UserName string
	UserType string
}

type AuthenticateUserReq struct{}
type AuthneticateUserRes struct{}

type AuthService interface {
	AuthenticateUser(context.Context, AuthenticateUserReq) (*AuthneticateUserRes, error)
}

type AuthServiceV2 interface {
	CreateAccount(context.Context, CreateAccountReq) (CreateAccountRes, error)
}

type CreateAccountReq struct {
	Name     string
	Password string
	Email    string
	Type     string
}
type CreateAccountRes struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

type AccountProvider interface {
	CreateAccount(context.Context, CreateAccountDTOInput) (CreateAccountDTOOutput, error)
	FindAccount(context.Context, FindAccountDTOInput) (*FindAccountDTOOutput, error)
	FindAccountByEmail(context.Context, FindAccountByEmailDTOInput) (*FindAccountByEmailDTOOutput, error)
	CheckAccountCredentials(context.Context, CheckAccountCredentialsDTOInput) (CheckAccountCredentialsDTOOutput, error)
	ConfirmAccountsByEmail(context.Context, ConfirmAccountsByEmailDTOInput) error
	// TODO: move this check to SQL query instead
	// GetIsAccountConfirmed(ctx context.Context, email string) (bool, error)
}

type CreateAccountDTOInput struct {
	Name     string
	Password string
	Email    string
	Type     string
}
type CreateAccountDTOOutput struct {
	Id    string
	Name  string
	Email string
	Type  string
}

type FindAccountDTOInput struct {
	Id string
}
type FindAccountDTOOutput struct {
	Name  string
	Email string
	Type  string
}

type FindAccountByEmailDTOInput struct {
	Email string
}
type FindAccountByEmailDTOOutput struct {
	Id   string
	Name string
	Type string
}

type CheckAccountCredentialsDTOInput struct {
	Email    string
	Password string
}
type CheckAccountCredentialsDTOOutput struct {
	Ok bool
}

type ConfirmAccountsByEmailDTOInput struct {
	Emails []string
}

type AccountConfirmationProvider interface {
	Send(context.Context, SendAccountConfirmationDTOInput) (SendAccountConfirmationDTOOutput, error)
}

type SendAccountConfirmationDTOInput struct {
	Name  string
	Email string
}
type SendAccountConfirmationDTOOutput struct {
}

type RefreshTokenProviderV2 interface {
	Get(ctx context.Context, subjectId string) (tokenIds []string, err error)
	Add(context.Context, *RefreshToken) error
	Delete(ctx context.Context, tokenId string) error
}
