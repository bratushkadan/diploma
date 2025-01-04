package domain

import (
	"context"
	"time"
)

// Also known as Subject.
type User struct {
	Id   string
	Name string
	Type string
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
	CreateUser(context.Context, UserProviderCreateUserReq) (User, error)
	FindUser(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	CheckUserCredentials(ctx context.Context, email string, password string) (*User, error)
}

type RefreshTokenPersisterProvider interface {
	Get(ctx context.Context, subjectId string) ([]RefreshToken, error)
	Add(context.Context, *RefreshToken) error
	Delete(ctx context.Context, tokenId string) error
}
