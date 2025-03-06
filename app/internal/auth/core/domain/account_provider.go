package domain

import (
	"context"

	"github.com/google/uuid"
)

type AccountProvider interface {
	CreateAccount(context.Context, CreateAccountDTOInput) (CreateAccountDTOOutput, error)
	FindAccount(context.Context, FindAccountDTOInput) (*FindAccountDTOOutput, error)
	FindAccountByEmail(context.Context, FindAccountByEmailDTOInput) (*FindAccountByEmailDTOOutput, error)
	CheckAccountCredentials(context.Context, CheckAccountCredentialsDTOInput) (CheckAccountCredentialsDTOOutput, error)
	ActivateAccountsByEmail(context.Context, ActivateAccountsByEmailDTOInput) error
}

type CreateAccountDTOInput struct {
	Id       uuid.UUID
	Name     string
	Password string
	Email    string
	Type     string
}
type CreateAccountDTOOutput struct {
	Id    uuid.UUID
	Name  string
	Email string
	Type  string
}

type FindAccountDTOInput struct {
	Id uuid.UUID
}
type FindAccountDTOOutput struct {
	Name      string
	Email     string
	Type      string
	Activated bool
}

type FindAccountByEmailDTOInput struct {
	Email string
}
type FindAccountByEmailDTOOutput struct {
	Id        uuid.UUID
	Name      string
	Type      string
	Activated bool
}

type CheckAccountCredentialsDTOInput struct {
	Email    string
	Password string
}
type CheckAccountCredentialsDTOOutput struct {
	Ok        bool
	Activated bool
	AccountId uuid.UUID
}

type ActivateAccountsByEmailDTOInput struct {
	Emails []string
}
