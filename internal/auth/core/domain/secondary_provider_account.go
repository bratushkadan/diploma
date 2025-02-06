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
	AuthenticateUser(ctx context.Context, req AuthenticateUserReq) (*AuthneticateUserRes, error)
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

// TODO: Delete "YDB" suffix and migrate
type AccountProviderYDB interface {
	CreateAccount(context.Context, CreateAccountDTOInput) (CreateAccountDTOOutput, error)
	// FindUser(ctx context.Context, id string) (*User, error)
	// FindUserByEmail(ctx context.Context, email string) (*User, error)
	// CheckUserCredentials(ctx context.Context, email string, password string) (*User, error)
	// AddEmailConfirmationId(ctx context.Context, email string) (string, error)
	// GetIsAccountConfirmed(ctx context.Context, email string) (bool, error)
	// ConfirmAccountByEmail(ctx context.Context, email string) error
}

type RefreshTokenPersisterProvider2 interface {
	Get(ctx context.Context, subjectId string) (tokenIds []string, err error)
	Add(context.Context, *RefreshToken) error
	Delete(ctx context.Context, tokenId string) error
}

type NotificationProvider2 interface {
	NotifyAccountCreated(ctx context.Context) error
}

type ConfirmationProvider2 interface {
	Send(ctx context.Context, emailAddr, confirmationId string) error
}
