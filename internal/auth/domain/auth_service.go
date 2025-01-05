package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/bratushkadan/floral/pkg/auth"
)

type AuthServiceConf struct {
	RefreshTokenProvider RefreshTokenProvider
	AccessTokenProvider  AccessTokenProvider

	UserProvider                  UserProvider
	RefreshTokenPersisterProvider RefreshTokenPersisterProvider

	// "Break the glass" token for managing admin accounts.
	SecretToken string
}

type AuthService struct {
	rtProv RefreshTokenProvider
	atProv AccessTokenProvider

	userProv  UserProvider
	rtPerProv RefreshTokenPersisterProvider

	secretToken string
}

func NewAuthService(conf *AuthServiceConf) *AuthService {
	return &AuthService{
		rtProv:      conf.RefreshTokenProvider,
		atProv:      conf.AccessTokenProvider,
		userProv:    conf.UserProvider,
		rtPerProv:   conf.RefreshTokenPersisterProvider,
		secretToken: conf.SecretToken,
	}
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrPermissionDenied    = errors.New("permission denied")
)

var (
	RegexEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

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

func (s *AuthService) validateEmail(email string) bool {
	return RegexEmail.MatchString(email)
}

func (s *AuthService) createUser(ctx context.Context, req CreateUserReq, userType string) (*User, error) {
	if ok := s.validateEmail(req.Email); !ok {
		return nil, ErrInvalidEmail
	}

	user, err := s.userProv.CreateUser(ctx, UserProviderCreateUserReq{
		Name:     req.Name,
		Password: req.Password,
		Email:    req.Email,
		Type:     userType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
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

func (s *AuthService) lookupRefreshToken(ctx context.Context, refreshTokenString string) (*RefreshToken, error) {
	token, err := s.rtProv.Decode(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode refresh token: %w", ErrInvalidRefreshToken, err)
	}

	tokenIds, err := s.rtPerProv.Get(ctx, token.SubjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve refresh tokens for subject: %w", err)
	}

	for _, id := range tokenIds {
		if id == token.TokenId {
			return token, nil
		}
	}

	return nil, fmt.Errorf("failed to lookup refresh token: %w", ErrInvalidRefreshToken)
}

func (s *AuthService) CreateCustomer(ctx context.Context, req CreateCustomerReq) (*User, error) {
	return s.createUser(ctx, req.CreateUserReq, auth.UserTypeCustomer)
}

func (s *AuthService) CreateSeller(ctx context.Context, req CreateSellerReq, accessTokenString string) (*User, error) {
	accessToken, err := s.atProv.Decode(accessTokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode access token: %w", err)
	}

	if accessToken.SubjectType != auth.UserTypeAdmin {
		return nil, fmt.Errorf("only admins can create seller accounts: %w", ErrPermissionDenied)
	}

	return s.createUser(ctx, req.CreateUserReq, auth.UserTypeSeller)
}

// Expose this method carefully.
func (s *AuthService) CreateAdmin(ctx context.Context, req CreateAdminReq) (*User, error) {
	if req.SecretToken == "" {
		return nil, errors.New("admin account creation is disabled: no secret token provided")
	}
	if req.SecretToken != s.secretToken {
		return nil, errors.New("failed to create admin account: incorrect secret token provided")
	}
	return s.createUser(ctx, req.CreateUserReq, auth.UserTypeAdmin)
}

// Returns token if the provided credentials are correct
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	user, err := s.userProv.CheckUserCredentials(ctx, email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return "", err
		}
		return "", fmt.Errorf("failed to check user credentials: %w", err)
	}

	return s.createPersistToken(ctx, user.Id)
}

func (s *AuthService) RenewRefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	token, err := s.lookupRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", err
	}

	tokenStr, err := s.createPersistToken(ctx, token.SubjectId)
	if err != nil {
		return "", err
	}

	if err := s.rtPerProv.Delete(ctx, token.TokenId); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *AuthService) GetAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
	token, err := s.lookupRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", err
	}

	user, err := s.userProv.FindUser(ctx, token.SubjectId)
	if err != nil {
		return "", fmt.Errorf("failed to request user: %w", err)
	}

	_, tokenStr, err := s.atProv.Create(user.Id, user.Type)
	if err != nil {
		return "", fmt.Errorf("failed to create access token: %w", err)
	}

	return tokenStr, nil
}
