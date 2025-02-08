package domain

import (
	"context"
	"time"
)

type RefreshTokenProvider interface {
	Get(context.Context, RefreshTokenGetDTOInput) (RefreshTokenGetDTOOutput, error)
	Add(context.Context, RefreshTokenAddDTOInput) (RefreshTokenAddDTOOutput, error)
	Replace(context.Context, RefreshTokenReplaceDTOInput) (RefreshTokenReplaceDTOOutput, error)
	Delete(context.Context, RefreshTokenDeleteDTOInput) (RefreshTokenDeleteDTOOutput, error)
	DeleteByAccountId(context.Context, RefreshTokenDeleteByAccountIdDTOInput) (RefreshTokenDeleteByAccountIdDTOOutput, error)
}

type RefreshTokenGetDTOInput struct {
	AccountId string
}
type RefreshTokenGetDTOOutput struct {
	Tokens []RefreshTokenGetDTOOutputToken `json:"tokens"`
}
type RefreshTokenGetDTOOutputToken struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshTokenAddDTOInput struct {
	AccountId string
	CreatedAt time.Time
	ExpiresAt time.Time
}
type RefreshTokenAddDTOOutput struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshTokenReplaceDTOInput struct {
	Id        string
	CreatedAt time.Time
	ExpiresAt time.Time
}
type RefreshTokenReplaceDTOOutput struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshTokenDeleteDTOInput struct {
	Id string
}
type RefreshTokenDeleteDTOOutput struct {
	Id string
}

type RefreshTokenDeleteByAccountIdDTOInput struct {
	Id string
}
type RefreshTokenDeleteByAccountIdDTOOutput struct {
	Ids []string
}
