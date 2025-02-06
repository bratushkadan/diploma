package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/auth"
	"go.uber.org/zap"
)

type AuthV2 struct {
	accProv             domain.AccountProvider
	accConfirmationProv domain.AccountConfirmationProvider

	l *zap.Logger
}

var _ domain.AuthServiceV2 = (*AuthV2)(nil)

type AuthV2Builder struct {
	auth *AuthV2
}

func (b *AuthV2Builder) AccountProvider(prov domain.AccountProvider) *AuthV2Builder {
	b.auth.accProv = prov
	return b
}
func (b *AuthV2Builder) AccountConfirmationProvider(prov domain.AccountConfirmationProvider) *AuthV2Builder {
	b.auth.accConfirmationProv = prov
	return b
}

func (b *AuthV2Builder) Logger(l *zap.Logger) *AuthV2Builder {
	b.auth.l = l
	return b
}

func (b *AuthV2Builder) Build() (*AuthV2, error) {
	return b.auth, nil
}

func NewAuthBuilder() *AuthV2Builder {
	return &AuthV2Builder{auth: &AuthV2{}}
}

func (svc *AuthV2) CreateAccount(ctx context.Context, req domain.CreateAccountReq) (domain.CreateAccountRes, error) {
	acc, err := domain.NewUserAccount(req.Name, req.Password, req.Email, req.Type)
	if err != nil {
		return domain.CreateAccountRes{}, err
	}

	out, err := svc.accProv.CreateAccount(ctx, domain.CreateAccountDTOInput{
		Name:     acc.Name(),
		Password: acc.Password(),
		Email:    acc.Email(),
		Type:     acc.Type(),
	})
	if err != nil {
		svc.l.Error("failed to create account", zap.Error(err))
		return domain.CreateAccountRes{}, err
	}

	accountRes := domain.CreateAccountRes{
		Id:    out.Id,
		Name:  out.Name,
		Email: out.Email,
		Type:  out.Type,
	}

	_, err = svc.accConfirmationProv.Send(ctx, domain.SendAccountConfirmationDTOInput{
		Name:  accountRes.Name,
		Email: accountRes.Email,
	})
	if err != nil {
		svc.l.Error("failed to send email confirmation message", zap.Error(err))
		err = fmt.Errorf("%w: %v", domain.ErrSendAccountConfirmationFailed, err)
		return domain.CreateAccountRes{}, err
	}

	return accountRes, nil
}

// -----------------------------

type AuthConf struct {
	RefreshTokenProvider domain.RefreshTokenProvider
	AccessTokenProvider  domain.AccessTokenProvider

	ConfirmationProvider          domain.ConfirmationProvider
	UserProvider                  domain.UserProvider
	RefreshTokenPersisterProvider domain.RefreshTokenPersisterProvider

	// "Break the glass" token for managing admin accounts.
	SecretToken string
}

type Auth struct {
	rtProv domain.RefreshTokenProvider
	atProv domain.AccessTokenProvider

	confirmationProv domain.ConfirmationProvider
	userProv         domain.UserProvider
	rtPerProv        domain.RefreshTokenPersisterProvider

	secretToken string
}

func NewAuth(conf *AuthConf) *Auth {
	return &Auth{
		rtProv: conf.RefreshTokenProvider,
		atProv: conf.AccessTokenProvider,

		confirmationProv: conf.ConfirmationProvider,
		userProv:         conf.UserProvider,
		rtPerProv:        conf.RefreshTokenPersisterProvider,

		secretToken: conf.SecretToken,
	}
}

// FIXME: add way to re-send confirmation link
func (s *Auth) createUserAccount(ctx context.Context, req domain.CreateUserReq, userType string) (*domain.User, error) {
	acc, err := domain.NewUserAccount(req.Name, req.Password, req.Email, userType)
	if err != nil {
		return nil, err
	}

	user, err := s.userProv.CreateUser(ctx, domain.UserProviderCreateUserReq{
		Name:     acc.Name(),
		Password: acc.Password(),
		Email:    acc.Email(),
		Type:     acc.Type(),
	})
	if err != nil {
		return nil, err
	}

	confirmationId, err := s.userProv.AddEmailConfirmationId(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := s.confirmationProv.Send(ctx, req.Email, confirmationId); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Auth) createPersistToken(ctx context.Context, subjectId string) (string, error) {
	token, tokenString, err := s.rtProv.Create(subjectId)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := s.rtPerProv.Add(ctx, token); err != nil {
		return "", fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return tokenString, nil
}

func (s *Auth) lookupRefreshToken(ctx context.Context, refreshTokenString string) (*domain.RefreshToken, error) {
	token, err := s.rtProv.Decode(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode refresh token: %w", domain.ErrInvalidRefreshToken, err)
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

	return nil, fmt.Errorf("failed to lookup refresh token: %w", domain.ErrInvalidRefreshToken)
}

func (s *Auth) CreateCustomer(ctx context.Context, req domain.CreateCustomerReq) (*domain.User, error) {
	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeCustomer)
}

func (s *Auth) CreateSeller(ctx context.Context, req domain.CreateSellerReq, accessTokenString string) (*domain.User, error) {
	accessToken, err := s.atProv.Decode(accessTokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidAccessToken, err)
	}

	if accessToken.SubjectType != auth.UserTypeAdmin {
		return nil, fmt.Errorf("only admins can create seller accounts: %w", domain.ErrPermissionDenied)
	}

	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeSeller)
}

// Expose this method carefully.
func (s *Auth) CreateAdmin(ctx context.Context, req domain.CreateAdminReq) (*domain.User, error) {
	if req.SecretToken == "" {
		return nil, errors.New("admin account creation is disabled: no secret token provided")
	}
	if req.SecretToken != s.secretToken {
		return nil, errors.New("failed to create admin account: incorrect secret token provided")
	}
	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeAdmin)
}

// Returns token if the provided credentials are correct
func (s *Auth) Authenticate(ctx context.Context, email, password string) (string, error) {
	user, err := s.userProv.CheckUserCredentials(ctx, email, password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return "", err
		}
		return "", fmt.Errorf("failed to check user credentials: %w", err)
	}

	ok, err := s.userProv.GetIsUserConfirmedByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", domain.ErrAccountEmailNotConfirmed
	}

	return s.createPersistToken(ctx, user.Id)
}

func (s *Auth) RenewRefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
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

func (s *Auth) GetAccessToken(ctx context.Context, refreshTokenString string) (*domain.AccessToken, string, error) {
	token, err := s.lookupRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, "", err
	}

	user, err := s.userProv.FindUser(ctx, token.SubjectId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to request user: %w", err)
	}

	accessToken, tokenStr, err := s.atProv.Create(user.Id, user.Type)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create access token: %w", err)
	}

	return accessToken, tokenStr, nil
}

func (s *Auth) ConfirmEmail(ctx context.Context, confirmationId string) error {
	return s.userProv.ConfirmEmailByConfirmationId(ctx, confirmationId)
}
