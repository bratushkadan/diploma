package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"go.uber.org/zap"
)

type Auth struct {
	accProv                     domain.AccountProvider
	accCreationNotificationProv domain.AccountCreationNotifications
	refreshTokenProv            domain.RefreshTokenProvider
	tokenProv                   domain.TokenProvider

	// token TTL for rows is also applied to the provider's refresh_tokens YDB table
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration

	l *zap.Logger
}

var _ domain.AuthService = (*Auth)(nil)

type AuthBuilder struct {
	auth *Auth
}

func (b *AuthBuilder) AccountProvider(prov domain.AccountProvider) *AuthBuilder {
	b.auth.accProv = prov
	return b
}
func (b *AuthBuilder) AccountCreationNotificationProvider(prov domain.AccountCreationNotifications) *AuthBuilder {
	b.auth.accCreationNotificationProv = prov
	return b
}
func (b *AuthBuilder) RefreshTokenProvider(prov domain.RefreshTokenProvider) *AuthBuilder {
	b.auth.refreshTokenProv = prov
	return b
}
func (b *AuthBuilder) TokenProvider(prov domain.TokenProvider) *AuthBuilder {
	b.auth.tokenProv = prov
	return b
}

func (b *AuthBuilder) RefreshTokenDuration(dur time.Duration) *AuthBuilder {
	b.auth.refreshTokenDuration = dur
	return b
}
func (b *AuthBuilder) AccessTokenDuration(dur time.Duration) *AuthBuilder {
	b.auth.accessTokenDuration = dur
	return b
}

func (b *AuthBuilder) Logger(l *zap.Logger) *AuthBuilder {
	b.auth.l = l
	return b
}

func (b *AuthBuilder) Build() (*Auth, error) {
	return b.auth, nil
}

func NewAuthBuilder() *AuthBuilder {
	auth := Auth{
		refreshTokenDuration: 30 * 24 * time.Hour,
		accessTokenDuration:  30 * time.Minute,
	}
	return &AuthBuilder{auth: &auth}
}

type createAccountReq struct {
	domain.CreateUserReq
	Name     string
	Email    string
	Password string
	Type     domain.AccountType
}

func (svc *Auth) createAccount(ctx context.Context, req createAccountReq) (domain.CreateUserRes, error) {
	acc, err := domain.NewAccount(req.Name, req.Password, req.Email, req.Type)
	if err != nil {
		svc.l.Info("failed to create new account from provided input", zap.Error(err))
		return domain.CreateUserRes{}, err
	}

	out, err := svc.accProv.CreateAccount(ctx, domain.CreateAccountDTOInput{
		Name:     acc.Name(),
		Password: acc.Password(),
		Email:    acc.Email(),
		Type:     acc.Type(),
	})
	if err != nil {
		svc.l.Error("failed to create account via account provider", zap.Error(err))
		return domain.CreateUserRes{}, err
	}

	accountRes := domain.CreateUserRes{
		Id:    out.Id,
		Name:  out.Name,
		Email: out.Email,
	}

	_, err = svc.accCreationNotificationProv.Send(ctx, domain.SendAccountCreationNotificationDTOInput{
		Email: accountRes.Email,
	})
	if err != nil {
		err = fmt.Errorf("%w: %v", domain.ErrSendAccountConfirmationFailed, err)
		svc.l.Error("failed to send account email confirmation message", zap.Error(err))
		return domain.CreateUserRes{}, err
	}

	return accountRes, nil
}

func (svc *Auth) CreateUser(ctx context.Context, req domain.CreateUserReq) (domain.CreateUserRes, error) {
	res, err := svc.createAccount(ctx, createAccountReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Type:     domain.AccountTypeUser,
	})
	if err != nil {
		return domain.CreateUserRes{}, err
	}
	return domain.CreateUserRes{
		Name:  res.Name,
		Email: res.Email,
		Id:    res.Id,
	}, nil
}

func (svc *Auth) CreateSeller(ctx context.Context, req domain.CreateSellerReq) (domain.CreateSellerRes, error) {
	token, err := svc.tokenProv.DecodeAccess(req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRefreshToken):
			svc.l.Info("invalid refresh token", zap.Error(err))
			return domain.CreateSellerRes{}, err
		case errors.Is(err, domain.ErrTokenExpired):
			svc.l.Info("refresh token expired", zap.Any("token", token))
			return domain.CreateSellerRes{}, err
		case errors.Is(err, domain.ErrTokenParseFailed):
		default:
			svc.l.Error("failed to decode refresh token: %w", zap.Error(err))
			return domain.CreateSellerRes{}, err
		}
	}

	res, err := svc.createAccount(ctx, createAccountReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Type:     domain.AccountTypeSeller,
	})
	if err != nil {
		return domain.CreateSellerRes{}, err
	}
	return domain.CreateSellerRes{
		Name:  res.Name,
		Email: res.Email,
		Id:    res.Id,
	}, nil
}

func (svc *Auth) CreateAdmin(ctx context.Context, req domain.CreateAdminReq) (domain.CreateAdminRes, error) {
	res, err := svc.createAccount(ctx, createAccountReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Type:     domain.AccountTypeAdmin,
	})
	if err != nil {
		return domain.CreateAdminRes{}, err
	}
	return domain.CreateAdminRes{
		Name:  res.Name,
		Email: res.Email,
		Id:    res.Id,
	}, nil
}

func (svc *Auth) ActivateAccounts(ctx context.Context, req domain.ActivateAccountsReq) (domain.ActivateAccountsRes, error) {
	if err := svc.accProv.ActivateAccountsByEmail(
		ctx, domain.ActivateAccountsByEmailDTOInput{Emails: req.Emails},
	); err != nil {
		svc.l.Error("failed to activate accounts by email", zap.Any("emails", req.Emails), zap.Error(err))
		return domain.ActivateAccountsRes{}, err
	}

	return domain.ActivateAccountsRes{}, nil
}

func (svc *Auth) Authenticate(ctx context.Context, req domain.AuthenticateReq) (domain.AuthenticateRes, error) {
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

	if !out.Activated {
		svc.l.Info("rejected creating refresh token for account that has not been activated")
		return domain.AuthenticateRes{}, domain.ErrAccountNotActivated
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

func (svc *Auth) ReplaceRefreshToken(ctx context.Context, req domain.ReplaceRefreshTokenReq) (domain.ReplaceRefreshTokenRes, error) {
	token, err := svc.tokenProv.DecodeRefresh(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRefreshToken):
			svc.l.Info("invalid refresh token", zap.Error(err))
			return domain.ReplaceRefreshTokenRes{}, err
		case errors.Is(err, domain.ErrTokenExpired):
			svc.l.Info("refresh token expired", zap.Any("token", token))
			return domain.ReplaceRefreshTokenRes{}, err
		case errors.Is(err, domain.ErrTokenParseFailed):
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
	if out.Id == "" {
		return domain.ReplaceRefreshTokenRes{}, domain.ErrRefreshTokenToReplaceNotFound
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

func (svc *Auth) CreateAccessToken(ctx context.Context, req domain.CreateAccessTokenReq) (domain.CreateAccessTokenRes, error) {
	refreshToken, err := svc.tokenProv.DecodeRefresh(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRefreshToken):
			svc.l.Info("invalid refresh token", zap.Error(err))
			return domain.CreateAccessTokenRes{}, err
		case errors.Is(err, domain.ErrTokenExpired):
			svc.l.Info("refresh token expired", zap.Any("token", refreshToken))
			return domain.CreateAccessTokenRes{}, err
		case errors.Is(err, domain.ErrTokenParseFailed):
		default:
			svc.l.Error("failed to decode refresh token: %w", zap.Error(err))
			return domain.CreateAccessTokenRes{}, err
		}
	}

	out, err := svc.refreshTokenProv.List(ctx, domain.RefreshTokenListDTOInput{
		AccountId: refreshToken.SubjectId,
	})
	if err != nil {
		svc.l.Error("failed to list refresh tokens for creating access token", zap.Error(err))
		return domain.CreateAccessTokenRes{}, err
	}

	if tokenNotRevoked := slices.ContainsFunc(out.Tokens, func(v domain.RefreshTokenListDTOOutputToken) bool {
		return v.Id == refreshToken.Id
	}); !tokenNotRevoked {
		return domain.CreateAccessTokenRes{}, domain.ErrTokenRevoked
	}

	acc, err := svc.accProv.FindAccount(ctx, domain.FindAccountDTOInput{
		Id: refreshToken.SubjectId,
	})
	if err != nil {
		svc.l.Error("failed to find account for creating access token: %v", zap.Error(err))
		return domain.CreateAccessTokenRes{}, err
	}

	accessToken := domain.AccessToken{
		SubjectId:   refreshToken.SubjectId,
		SubjectType: acc.Type,
		ExpiresAt:   time.Now().Add(svc.accessTokenDuration),
	}
	token, err := svc.tokenProv.EncodeAccess(accessToken)
	if err != nil {
		svc.l.Error("failed to encode access token: %v", zap.Error(err))
		return domain.CreateAccessTokenRes{}, err
	}

	return domain.CreateAccessTokenRes{
		AccessToken: token,
		ExpiresAt:   accessToken.ExpiresAt,
	}, nil
}
