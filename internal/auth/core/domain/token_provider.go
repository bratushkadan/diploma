package domain

import (
	"context"
	"time"
)

type RefreshTokenProviderV2 interface {
	Get(context.Context, RefreshTokenGetDTOInput) (RefreshTokenGetDTOOutput, error)
	Add(context.Context, RefreshTokenAddDTOInput) (RefreshTokenAddDTOOutput, error)
	Replace(context.Context, RefreshTokenReplaceDTOOutput) (RefreshTokenReplaceDTOOutput, error)
	Delete(context.Context, RefreshTokenDeleteDTOInput) (RefreshTokenDeleteDTOOutput, error)
}

type RefreshTokenGetDTOInput struct {
	AccountId string
}
type RefreshTokenGetDTOOutput struct {
	Tokens []struct {
		Id string `json:"id"`
	} `json:"tokens"`
}

type RefreshTokenAddDTOInput struct {
	AccountId string
	Type      string
	ExpiresAt time.Time
}
type RefreshTokenAddDTOOutput struct {
	Id string
}

type RefreshTokenReplaceDTOInput struct {
	Id string
}
type RefreshTokenReplaceDTOOutput struct {
	Id string
}

type RefreshTokenDeleteDTOInput struct {
	Id string
}
type RefreshTokenDeleteDTOOutput struct {
	Id string
}
