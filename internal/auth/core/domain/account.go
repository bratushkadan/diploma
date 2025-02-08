package domain

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailIsInUse        = errors.New("email is in use")
	ErrAccountNotActivated = errors.New("account not activated")
)

var (
	regexDomain = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

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
	return regexDomain.MatchString(a.email)
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
