package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"go.uber.org/zap"
)

type AuthV2 struct {
	accProv             domain.AccountProvider
	accConfirmationProv domain.AccountConfirmationProvider
	refreshTokenProv    domain.RefreshTokenProvider
	tokenProv           domain.TokenProvider

	// token TTL for rows is also applied to the provider's refresh_tokens YDB table
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration

	l *zap.Logger
}

var _ domain.AuthServiceV2 = (*AuthV2)(nil)

type AuthV2Builder struct {
	auth *AuthV2
}

func (b *AuthV2Builder) TokenProvider(prov domain.TokenProvider) *AuthV2Builder {
	b.auth.tokenProv = prov
	return b
}
func (b *AuthV2Builder) AccountProvider(prov domain.AccountProvider) *AuthV2Builder {
	b.auth.accProv = prov
	return b
}
func (b *AuthV2Builder) AccountConfirmationProvider(prov domain.AccountConfirmationProvider) *AuthV2Builder {
	b.auth.accConfirmationProv = prov
	return b
}

func (b *AuthV2Builder) RefreshTokenDuration(dur time.Duration) *AuthV2Builder {
	b.auth.refreshTokenDuration = dur
	return b
}
func (b *AuthV2Builder) AccessTokenDuration(dur time.Duration) *AuthV2Builder {
	b.auth.accessTokenDuration = dur
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
	auth := AuthV2{
		refreshTokenDuration: 30 * 24 * time.Hour,
		accessTokenDuration:  30 * time.Minute,
	}
	return &AuthV2Builder{auth: &auth}
}

func (svc *AuthV2) CreateAccount(ctx context.Context, req domain.CreateAccountReq) (domain.CreateAccountRes, error) {
	acc, err := domain.NewAccount(req.Name, req.Password, req.Email, req.Type)
	if err != nil {
		svc.l.Error("failed to create new account from provided input", zap.Error(err))
		return domain.CreateAccountRes{}, err
	}

	out, err := svc.accProv.CreateAccount(ctx, domain.CreateAccountDTOInput{
		Name:     acc.Name(),
		Password: acc.Password(),
		Email:    acc.Email(),
		Type:     acc.Type(),
	})
	if err != nil {
		svc.l.Error("failed to create account via account provider", zap.Error(err))
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
		err = fmt.Errorf("%w: %v", domain.ErrSendAccountConfirmationFailed, err)
		svc.l.Error("failed to send account email confirmation message", zap.Error(err))
		return domain.CreateAccountRes{}, err
	}

	return accountRes, nil
}

func (svc *AuthV2) ActivateAccounts(ctx context.Context, req domain.ActivateAccountsReq) (domain.ActivateAccountsRes, error) {
	if err := svc.accProv.ActivateAccountsByEmail(
		ctx, domain.ActivateAccountsByEmailDTOInput{Emails: req.Emails},
	); err != nil {
		svc.l.Error("failed to activate accounts by email", zap.Any("emails", req.Emails), zap.Error(err))
		return domain.ActivateAccountsRes{}, err
	}

	return domain.ActivateAccountsRes{}, nil
}

func (svc *AuthV2) Authenticate(ctx context.Context, req domain.AuthenticateReq) (domain.AuthenticateRes, error) {
	out, err := svc.accProv.CheckAccountCredentials(ctx, domain.CheckAccountCredentialsDTOInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		svc.l.Error("failed to authenticate account", zap.String("email", req.Email), zap.Error(err))
		return domain.AuthenticateRes{}, err
	}

	if !out.Ok {
		return domain.AuthenticateRes{}, domain.ErrInvalidCredentials
	}

	token := domain.RefreshToken{
		SubjectId: out.AccountId,
	}

	// FIXME: clean Go transactions
	outToken, err := svc.refreshTokenProv.Add(ctx, domain.RefreshTokenAddDTOInput{
		AccountId: out.AccountId,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(svc.refreshTokenDuration),
	})
	if err != nil {
		svc.l.Error("failed to add data on refresh token", zap.Error(err))
		return domain.AuthenticateRes{}, err
	}
	token.Id = outToken.Id
	token.ExpiresAt = outToken.ExpiresAt

	tokenStr, err := svc.tokenProv.EncodeRefresh(token)
	if err != nil {
		svc.l.Error("failed to encode refresh token", zap.Any("token_to_encode", token), zap.Any("refresh_token_adapter_output", outToken))
		return domain.AuthenticateRes{}, err
	}

	return domain.AuthenticateRes{
		RefreshToken: tokenStr,
		ExpiresAt:    token.ExpiresAt,
	}, nil
}

func (svc *AuthV2) ReplaceRefreshToken(ctx context.Context, req domain.ReplaceRefreshTokenReq) (domain.ReplaceRefreshTokenRes, error) {
	token, err := svc.tokenProv.DecodeRefresh(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRefreshToken):
			svc.l.Info("invalid refresh token", zap.Error(err))
			return domain.ReplaceRefreshTokenRes{}, err
		case errors.Is(err, domain.ErrTokenExpired):
			svc.l.Info("refresh token expired", zap.Any("token", token))
			return domain.ReplaceRefreshTokenRes{}, err
		default:
			svc.l.Error("failed to decode refresh token: %w", zap.Error(err))
			return domain.ReplaceRefreshTokenRes{}, err
		}
	}

	out, err := svc.refreshTokenProv.Replace(ctx, domain.RefreshTokenReplaceDTOInput{
		Id:        token.Id,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(svc.refreshTokenDuration),
	})
	if err != nil {
		svc.l.Error("failed to replace refresh token: %w", zap.Error(err))
		return domain.ReplaceRefreshTokenRes{}, err
	}

	newToken := domain.RefreshToken{
		Id:        out.Id,
		SubjectId: token.SubjectId,
		ExpiresAt: out.ExpiresAt,
	}

	newTokenEncoded, err := svc.tokenProv.EncodeRefresh(newToken)
	if err != nil {
		svc.l.Error("failed to encode refresh token", zap.Error(err))
		return domain.ReplaceRefreshTokenRes{}, err
	}

	return domain.ReplaceRefreshTokenRes{
		RefreshToken: newTokenEncoded,
		ExpiresAt:    newToken.ExpiresAt,
	}, nil
}

// -----------------------------

// type AuthConf struct {
// 	RefreshTokenProvider domain.RefreshTokenProvider
// 	AccessTokenProvider  domain.AccessTokenProvider
//
// 	ConfirmationProvider          domain.ConfirmationProvider
// 	UserProvider                  domain.UserProvider
// 	RefreshTokenPersisterProvider domain.RefreshTokenPersisterProvider
//
// 	// "Break the glass" token for managing admin accounts.
// 	SecretToken string
// }
//
// type Auth struct {
// 	rtProv domain.RefreshTokenProvider
// 	atProv domain.AccessTokenProvider
//
// 	confirmationProv domain.ConfirmationProvider
// 	userProv         domain.UserProvider
// 	rtPerProv        domain.RefreshTokenPersisterProvider
//
// 	secretToken string
// }
//
// func NewAuth(conf *AuthConf) *Auth {
// 	return &Auth{
// 		rtProv: conf.RefreshTokenProvider,
// 		atProv: conf.AccessTokenProvider,
//
// 		confirmationProv: conf.ConfirmationProvider,
// 		userProv:         conf.UserProvider,
// 		rtPerProv:        conf.RefreshTokenPersisterProvider,
//
// 		secretToken: conf.SecretToken,
// 	}
// }
//
// // FIXME: add way to re-send confirmation link
// func (s *Auth) createUserAccount(ctx context.Context, req domain.CreateUserReq, userType string) (*domain.User, error) {
// 	acc, err := domain.NewAccount(req.Name, req.Password, req.Email, userType)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	user, err := s.userProv.CreateUser(ctx, domain.UserProviderCreateUserReq{
// 		Name:     acc.Name(),
// 		Password: acc.Password(),
// 		Email:    acc.Email(),
// 		Type:     acc.Type(),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	confirmationId, err := s.userProv.AddEmailConfirmationId(ctx, req.Email)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if err := s.confirmationProv.Send(ctx, req.Email, confirmationId); err != nil {
// 		return nil, err
// 	}
//
// 	return user, nil
// }
//
// func (s *Auth) createPersistToken(ctx context.Context, subjectId string) (string, error) {
// 	token, tokenString, err := s.rtProv.Create(subjectId)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create refresh token: %w", err)
// 	}
//
// 	if err := s.rtPerProv.Add(ctx, token); err != nil {
// 		return "", fmt.Errorf("failed to persist refresh token: %w", err)
// 	}
//
// 	return tokenString, nil
// }
//
// func (s *Auth) lookupRefreshToken(ctx context.Context, refreshTokenString string) (*domain.RefreshToken, error) {
// 	token, err := s.rtProv.Decode(refreshTokenString)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: failed to decode refresh token: %w", domain.ErrInvalidRefreshToken, err)
// 	}
//
// 	tokenIds, err := s.rtPerProv.Get(ctx, token.SubjectId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve refresh tokens for subject: %w", err)
// 	}
//
// 	for _, id := range tokenIds {
// 		if id == token.TokenId {
// 			return token, nil
// 		}
// 	}
//
// 	return nil, fmt.Errorf("failed to lookup refresh token: %w", domain.ErrInvalidRefreshToken)
// }
//
// func (s *Auth) CreateCustomer(ctx context.Context, req domain.CreateCustomerReq) (*domain.User, error) {
// 	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeCustomer)
// }
//
// func (s *Auth) CreateSeller(ctx context.Context, req domain.CreateSellerReq, accessTokenString string) (*domain.User, error) {
// 	accessToken, err := s.atProv.Decode(accessTokenString)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidAccessToken, err)
// 	}
//
// 	if accessToken.SubjectType != auth.UserTypeAdmin {
// 		return nil, fmt.Errorf("only admins can create seller accounts: %w", domain.ErrPermissionDenied)
// 	}
//
// 	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeSeller)
// }
//
// // Expose this method carefully.
// func (s *Auth) CreateAdmin(ctx context.Context, req domain.CreateAdminReq) (*domain.User, error) {
// 	if req.SecretToken == "" {
// 		return nil, errors.New("admin account creation is disabled: no secret token provided")
// 	}
// 	if req.SecretToken != s.secretToken {
// 		return nil, errors.New("failed to create admin account: incorrect secret token provided")
// 	}
// 	return s.createUserAccount(ctx, req.CreateUserReq, auth.UserTypeAdmin)
// }
//
// // Returns token if the provided credentials are correct
// func (s *Auth) Authenticate(ctx context.Context, email, password string) (string, error) {
// 	user, err := s.userProv.CheckUserCredentials(ctx, email, password)
// 	if err != nil {
// 		if errors.Is(err, domain.ErrInvalidCredentials) {
// 			return "", err
// 		}
// 		return "", fmt.Errorf("failed to check user credentials: %w", err)
// 	}
//
// 	ok, err := s.userProv.GetIsUserConfirmedByEmail(ctx, email)
// 	if err != nil {
// 		return "", err
// 	}
// 	if !ok {
// 		return "", domain.ErrAccountEmailNotConfirmed
// 	}
//
// 	return s.createPersistToken(ctx, user.Id)
// }
//
// func (s *Auth) ReplaceRefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
// 	token, err := s.lookupRefreshToken(ctx, refreshTokenString)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	tokenStr, err := s.createPersistToken(ctx, token.SubjectId)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	if err := s.rtPerProv.Delete(ctx, token.Id); err != nil {
// 		return "", err
// 	}
//
// 	return tokenStr, nil
// }
//
// func (s *Auth) GetAccessToken(ctx context.Context, refreshTokenString string) (*domain.AccessToken, string, error) {
// 	token, err := s.lookupRefreshToken(ctx, refreshTokenString)
// 	if err != nil {
// 		return nil, "", err
// 	}
//
// 	user, err := s.userProv.FindUser(ctx, token.SubjectId)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("failed to request user: %w", err)
// 	}
//
// 	accessToken, tokenStr, err := s.atProv.Create(user.Id, user.Type)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("failed to create access token: %w", err)
// 	}
//
// 	return accessToken, tokenStr, nil
// }
//
// func (s *Auth) ConfirmEmail(ctx context.Context, confirmationId string) error {
// 	return s.userProv.ConfirmEmailByConfirmationId(ctx, confirmationId)
// }
