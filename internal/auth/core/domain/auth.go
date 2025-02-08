package domain

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailIsInUse     = errors.New("email is in use")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidTokenType    = errors.New("invalid token type")
	ErrTokenExpired        = errors.New("token expired")
)

var (
	ErrAccountEmailNotConfirmed = errors.New("account email is not confirmed")
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

type Account struct {
	name        string
	password    string
	email       string
	accountType string
}

func (a Account) Name() string {
	return a.name
}
func (a Account) Password() string {
	return a.password
}
func (a Account) Email() string {
	return a.email
}
func (a Account) Type() string {
	return a.accountType
}
func (a Account) validateEmail() bool {
	return RegexEmail.MatchString(a.email)
}

func NewAccount(name, password, email, accountType string) (Account, error) {
	acc := Account{
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

type TokenType string

var (
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeAccess  TokenType = "access"
)

type RefreshToken struct {
	Id        string
	SubjectId string
	ExpiresAt time.Time
}

type AccessToken struct {
	SubjectId   string
	SubjectType string
	ExpiresAt   time.Time
}

type TokenProvider interface {
	EncodeRefresh(token RefreshToken) (tokenString string, err error)
	DecodeRefresh(token string) (RefreshToken, error)
	EncodeAccess(token AccessToken) (tokenString string, err error)
	DecodeAccess(token string) (AccessToken, error)
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
	Add(context.Context, RefreshToken) error
	Delete(ctx context.Context, tokenId string) error
}

type NotificationProvider interface {
	NotifyAccountCreated(ctx context.Context) error
}

type ConfirmationProvider interface {
	Send(ctx context.Context, emailAddr, confirmationId string) error
}
