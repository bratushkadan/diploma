package domain

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

var (
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrPasswordTooLong     = errors.New("password too long")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailIsInUse        = errors.New("email is in use")
	ErrAccountNotActivated = errors.New("account not activated")
	ErrPermissionDenied    = errors.New("permission denied")
)

var (
	regexDomain = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type AccountType = string

const (
	AccountTypeUser   AccountType = "user"
	AccountTypeSeller AccountType = "seller"
	AccountTypeAdmin  AccountType = "admin"
)

type Account struct {
	name        string
	password    string
	email       string
	accountType AccountType
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
func (a Account) Type() AccountType {
	return a.accountType
}

func (a Account) validateEmail() error {
	if regexDomain.MatchString(a.email) {
		return nil
	}
	return ErrInvalidEmail
}

func (a Account) validatePassword() error {
	if utf8.RuneCountInString(a.password) > 24 {
		// Hashed sequences of byte length > 72 by bcrypt are not valid.
		return ErrPasswordTooLong
	}

	return nil
}

func NewAccount(name, password, email, accountType string) (Account, error) {
	acc := Account{
		name:        name,
		password:    password,
		email:       email,
		accountType: accountType,
	}

	var errs []error

	if err := acc.validateEmail(); err != nil {
		errs = append(errs, err)
	}

	if err := acc.validatePassword(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return acc, errors.Join(errs...)
	}

	return acc, nil
}
