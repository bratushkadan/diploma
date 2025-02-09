package domain

import "context"

type CheckUserCredentialsDTOOutput struct {
	UserId   string
	UserName string
	UserType string
}

type AccountProvider interface {
	CreateAccount(context.Context, CreateAccountDTOInput) (CreateAccountDTOOutput, error)
	FindAccount(context.Context, FindAccountDTOInput) (*FindAccountDTOOutput, error)
	FindAccountByEmail(context.Context, FindAccountByEmailDTOInput) (*FindAccountByEmailDTOOutput, error)
	CheckAccountCredentials(context.Context, CheckAccountCredentialsDTOInput) (CheckAccountCredentialsDTOOutput, error)
	ActivateAccountsByEmail(context.Context, ActivateAccountsByEmailDTOInput) error
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
	Name      string
	Email     string
	Type      string
	Activated bool
}

type FindAccountByEmailDTOInput struct {
	Email string
}
type FindAccountByEmailDTOOutput struct {
	Id        string
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
	AccountId string
}

type ActivateAccountsByEmailDTOInput struct {
	Emails []string
}
