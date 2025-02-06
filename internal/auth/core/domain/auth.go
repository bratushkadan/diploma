package domain

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidEmail             = errors.New("invalid email address")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrInvalidAccessToken       = errors.New("invalid access token")
	ErrPermissionDenied         = errors.New("permission denied")
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailIsInUse             = errors.New("email is in use")
	ErrAccountEmailNotConfirmed = errors.New("account email is not confirmed")
)

// TODO: domain only errors

// Adapter errors
var (
	ErrSendAccountConfirmationFailed = errors.New("failed to send account confirmation email")
)

var (
	RegexEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// Also known as Subject.
type User struct {
	Id   string
	Name string
	Type string
}

type CreateUserReq struct {
	Name     string
	Password string
	Email    string
}
type CreateCustomerReq struct {
	CreateUserReq
}
type CreateSellerReq struct {
	CreateUserReq
}
type CreateAdminReq struct {
	CreateUserReq
	SecretToken string
}

func NewUser() User {
	return User{}
}

type UserAccount struct {
	name        string
	password    string
	email       string
	accountType string
}

func (a UserAccount) Name() string {
	return a.name
}
func (a UserAccount) Password() string {
	return a.password
}
func (a UserAccount) Email() string {
	return a.email
}
func (a UserAccount) Type() string {
	return a.accountType
}
func (a UserAccount) validateEmail() bool {
	return RegexEmail.MatchString(a.email)
}

func NewUserAccount(name, password, email, accountType string) (UserAccount, error) {
	acc := UserAccount{
		name:        name,
		password:    password,
		email:       email,
		accountType: accountType,
	}

	if !acc.validateEmail() {
		return acc, ErrInvalidEmail
	}

	return acc, nil
}

type RefreshToken struct {
	TokenId   string
	TokenType string
	SubjectId string
	ExpiresAt time.Time
}

type AccessToken struct {
	TokenType   string
	SubjectId   string
	SubjectType string
	ExpiresAt   time.Time
}

type RefreshTokenProvider interface {
	Create(subjectId string) (token *RefreshToken, tokenString string, err error)
	Decode(token string) (*RefreshToken, error)
}

type AccessTokenProvider interface {
	Create(subjectId string, subjectType string) (token *AccessToken, tokenString string, err error)
	Decode(token string) (*AccessToken, error)
}

type UserProviderCreateUserReq struct {
	Name     string
	Password string
	Email    string
	Type     string
}

type UserProvider interface {
	CreateUser(context.Context, UserProviderCreateUserReq) (*User, error)
	FindUser(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	CheckUserCredentials(ctx context.Context, email string, password string) (*User, error)
	AddEmailConfirmationId(ctx context.Context, email string) (string, error)
	GetIsUserConfirmedByEmail(ctx context.Context, email string) (bool, error)
	ConfirmEmailByConfirmationId(ctx context.Context, id string) error
}

type RefreshTokenPersisterProvider interface {
	Get(ctx context.Context, subjectId string) (tokenIds []string, err error)
	Add(context.Context, *RefreshToken) error
	Delete(ctx context.Context, tokenId string) error
}

type NotificationProvider interface {
	NotifyAccountCreated(ctx context.Context) error
}

type ConfirmationProvider interface {
	Send(ctx context.Context, emailAddr, confirmationId string) error
}
