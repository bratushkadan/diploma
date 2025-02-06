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
	// FindUser(ctx context.Context, id string) (*User, error)
	// FindUserByEmail(ctx context.Context, email string) (*User, error)
	// CheckUserCredentials(ctx context.Context, email string, password string) (*User, error)
	// AddEmailConfirmationId(ctx context.Context, email string) (string, error)
	// GetIsAccountConfirmed(ctx context.Context, email string) (bool, error)
	// ConfirmAccountByEmail(ctx context.Context, email string) error
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
