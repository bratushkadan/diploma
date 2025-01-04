package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/bratushkadan/floral/pkg/auth"
)

type AuthService struct {
	rtProv RefreshTokenProvider
	atProv AccessTokenProvider

	userProv  UserProvider
	rtPerProv RefreshTokenPersisterProvider
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

var (
	RegexEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type CreateCustomerReq struct {
	Name     string
	Password string
	Email    string
}

func (s *AuthService) validateEmail(email string) bool {
	return RegexEmail.MatchString(email)
}

func (s *AuthService) CreateCustomer(ctx context.Context, req CreateCustomerReq) (*User, error) {
	if ok := s.validateEmail(req.Email); !ok {
		return nil, ErrInvalidEmail
	}

	user, err := s.userProv.CreateUser(ctx, UserProviderCreateUserReq{
		Name:     req.Name,
		Password: req.Password,
		Email:    req.Email,
		Type:     auth.UserTypeCustomer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Returns token if the provided credentials are correct
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	user, err := s.userProv.CheckUserCredentials(ctx, email, password)
	if err != nil {
		return "", fmt.Errorf("failed to check user credentials: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	token, tokenString, err := s.rtProv.Create(user.Id)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := s.rtPerProv.Add(ctx, token); err != nil {
		return "", fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) createPersistToken(ctx context.Context, subjectId string) (string, error) {
	token, tokenString, err := s.rtProv.Create(subjectId)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := s.rtPerProv.Add(ctx, token); err != nil {
		return "", fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) RenewRefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	token, err := s.rtProv.Decode(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("%w: failed to decode refresh token: %w", ErrInvalidRefreshToken, err)
	}

	if token.ExpiresAt.After(time.Now()) {
		return "", fmt.Errorf("%w: token is expired", ErrInvalidRefreshToken)
	}

	tokens, err := s.rtPerProv.Get(ctx, token.SubjectId)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve refresh tokens for subject: %w", err)
	}

	for _, v := range tokens {
		if v.TokenId == token.TokenId {
			newToken, err := s.createPersistToken(ctx, token.SubjectId)
			if err != nil {
				return "", fmt.Errorf("failed to create and persist new refresh token: %w", err)
			}

			return newToken, nil
		}
	}

	return "", fmt.Errorf("failed to lookup refresh token: %w", ErrInvalidRefreshToken)
}

func (s *AuthService) GetAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
}
