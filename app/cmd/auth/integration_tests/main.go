package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/infrastructure/authn"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/bratushkadan/floral/pkg/ymq"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/ydb-platform/ydb-go-sdk/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, "metadata")

	ydbFullEndpoint := cfg.MustEnv(setup.EnvKeyYdbEndpoint)

	sqsQueueUrl := cfg.MustEnv(setup.EnvKeySqsQueueUrlAccountCreations)
	sqsAccessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	sqsSecretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	logger, err := logging.NewZapConf("dev").Build()
	if err != nil {
		log.Fatal("Error setting up zap")
	}

	accountIdHasher, err := idhash.New(os.Getenv(setup.EnvKeyAccountIdHashSalt), idhash.WithPrefix("ie"))
	if err != nil {
		log.Fatal(err)
	}
	tokenIdHasher, err := idhash.New(os.Getenv(setup.EnvKeyTokenIdHashSalt), idhash.WithPrefix("rb"))
	if err != nil {
		log.Fatal(err)
	}
	passwordHasher, err := auth.NewPasswordHasher(os.Getenv(setup.EnvKeyPasswordHashSalt))
	if err != nil {
		logger.Fatal("failed to set up password hasher", zap.Error(err))
	}

	ctx := context.Background()

	logger.Debug("setup ydb")
	db, err := ydb.Open(ctx, ydbFullEndpoint, ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	logger.Debug("set up ydb")
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Print()
		}
	}()

	tokenProvider, err := authn.NewTokenProviderBuilder().
		PrivateKeyPath(os.Getenv(setup.EnvKeyAuthTokenPrivateKeyPath)).
		PublicKeyPath(os.Getenv(setup.EnvKeyAuthTokenPublicKeyPath)).
		Build()
	if err != nil {
		logger.Fatal("failed to setup token provider", zap.Error(err))
	}
	accountAdapter := ydb_adapter.NewAccount(ydb_adapter.AccountConf{
		DbDriver:       db,
		IdHasher:       accountIdHasher,
		PasswordHasher: passwordHasher,
		Logger:         logger,
	})
	refreshTokenAdapter := ydb_adapter.NewToken(ydb_adapter.TokenConf{
		DbDriver: db,
		IdHasher: tokenIdHasher,
		Logger:   logger,
	})

	sqsClient, err := ymq.New(ctx, sqsAccessKeyId, sqsSecretAccessKey, sqsQueueUrl, logger)
	if err != nil {
		logger.Fatal("failed to setup ymq")
	}
	accountCreationNotificationAdapter := ymq_adapter.AccountCreation{
		Sqs:         sqsClient,
		SqsQueueUrl: sqsQueueUrl,
	}

	authSvc, err := service.NewAuthBuilder().
		AccountProvider(accountAdapter).
		RefreshTokenProvider(refreshTokenAdapter).
		TokenProvider(tokenProvider).
		AccountCreationNotificationProvider(&accountCreationNotificationAdapter).
		Logger(zap.NewNop()).
		Build()
	if err != nil {
		logger.Fatal("could not build auth svc", zap.Error(err))
	}

	if err := runAccountTests(ctx, accountIdHasher, logger, authSvc, tokenProvider, refreshTokenAdapter, accountAdapter); err != nil {
		logger.Fatal("failed to run account service tests", zap.Error(err))
	}
}

func runAccountTests(ctx context.Context, accountIdHasher idhash.IdHasher, logger *zap.Logger, svc domain.AuthService, tokenProvider domain.TokenProvider, refreshTokenAdapter domain.RefreshTokenProvider, accAdapter domain.AccountProvider) error {
	logger.Info("create account")
	email := fmt.Sprintf(`someemail-%d@gmail.com`, time.Now().UnixMilli())
	password := uuid.New().String()[:24]
	resp, err := svc.CreateUser(ctx, domain.CreateUserReq{
		Name:     "Danila",
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("error creating account: %w", err)
	} else {
		logger.Info("response creating account", zap.Any("response", resp))

		idInt64, err := accountIdHasher.DecodeInt64(resp.Id)
		if err != nil {
			return fmt.Errorf("failed to decode str id %s to in64: %w", resp.Id, err)
		}
		logger.Info("decoded string id to int64", zap.String("str_id", resp.Id), zap.Int64("id", idInt64))
	}
	accountId := resp.Id

	logger.Info("find account")
	acc, err := accAdapter.FindAccount(ctx, domain.FindAccountDTOInput{Id: accountId})
	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}
	if acc != nil {
		logger.Info("found account", zap.Any("account", acc))
	} else {
		logger.Info("account not found", zap.String("id", resp.Id))
	}

	logger.Info("find account by email")
	accByEmail, err := accAdapter.FindAccountByEmail(ctx, domain.FindAccountByEmailDTOInput{Email: email})
	if err != nil {
		return fmt.Errorf("failed to find account by email: %w", err)
	}
	if accByEmail != nil {
		logger.Info("found account by email", zap.Any("account", accByEmail))
	} else {
		logger.Info("account not found", zap.String("id", resp.Id))
	}

	logger.Info("check account credentials")
	if out, err := accAdapter.CheckAccountCredentials(ctx, domain.CheckAccountCredentialsDTOInput{
		Email:    email,
		Password: password,
	}); err != nil {
		return fmt.Errorf("failed to check account credentials: %w", err)
	} else if !out.Ok {
		return errors.New("wrong credentials")
	}
	logger.Info("you're logged in!")

	logger.Info("activate accounts by email")
	if err := accAdapter.ActivateAccountsByEmail(ctx, domain.ActivateAccountsByEmailDTOInput{
		Emails: []string{email},
	}); err != nil {
		return fmt.Errorf("failed to activate account with email=%s: %w", email, err)
	}
	logger.Info("account activated", zap.String("email", email))

	logger.Info("authenticate account")
	authenticateRes, err := svc.Authenticate(ctx, domain.AuthenticateReq{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate account: %w", err)
	}
	logger.Info(
		"authenticated",
		zap.String("refresh_token", authenticateRes.RefreshToken),
		zap.Any("expires_at", authenticateRes.ExpiresAt),
	)

	refreshToken := authenticateRes.RefreshToken

	logger.Info("create access token")
	createAccountTokenRes, err := svc.CreateAccessToken(ctx, domain.CreateAccessTokenReq{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to create access token: %w", err)
	}
	logger.Info(
		"created access token",
		zap.String("access_token", createAccountTokenRes.AccessToken),
		zap.Any("expires_at", createAccountTokenRes.ExpiresAt),
	)

	logger.Info("replace refresh token")
	replaceRefreshTokenRes, err := svc.ReplaceRefreshToken(ctx, domain.ReplaceRefreshTokenReq{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to replace refresh token: %w", err)
	}

	logger.Info("create access token with the replaced refresh token")
	createAccountTokenRes, err = svc.CreateAccessToken(ctx, domain.CreateAccessTokenReq{
		RefreshToken: refreshToken,
	})
	if err == nil {
		return fmt.Errorf("expected create access token with the replaced refresh token to fail")
	}
	logger.Info("failed to create access token with the replaced refresh token - as expected")

	refreshToken = replaceRefreshTokenRes.RefreshToken
	logger.Info("create access token with the replaced refresh token")
	createAccountTokenRes, err = svc.CreateAccessToken(ctx, domain.CreateAccessTokenReq{
		RefreshToken: replaceRefreshTokenRes.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to create access token with the replaced refresh token: %w", err)
	}
	logger.Info(
		"created access token with the replaced refresh token",
		zap.String("access_token", createAccountTokenRes.AccessToken),
		zap.Any("expires_at", createAccountTokenRes.ExpiresAt),
	)

	logger.Info("authenticate multiple times to issue multiple refresh tokens and check whether older tokens are deleted")
	n := 6
	authMultTimesResponses := make([]domain.AuthenticateRes, 0, n)
	for range n {
		res, err := svc.Authenticate(ctx, domain.AuthenticateReq{
			Email:    email,
			Password: password,
		})
		if err != nil {
			return fmt.Errorf("failed to authenticate account: %w", err)
		}

		authMultTimesResponses = append(authMultTimesResponses, res)
	}
	logger.Info(fmt.Sprintf("authenticated %d times", n))

	logger.Info("list account refresh tokens")
	accountRefreshTokensRes, err := refreshTokenAdapter.List(ctx, domain.RefreshTokenListDTOInput{
		AccountId: resp.Id,
	})

	for _, respToken := range authMultTimesResponses[1:] {
		refreshToken, err := tokenProvider.DecodeRefresh(respToken.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to decode refresh token from multiple auth: %v", err)
		}

		if !slices.ContainsFunc(accountRefreshTokensRes.Tokens, func(token domain.RefreshTokenListDTOOutputToken) bool {
			return refreshToken.Id == refreshToken.Id
		}) {
			return errors.New("list refresh tokens response doesn't contain refresh tokens that should not have been revoked")
		}
	}
	logger.Info("checked there's a correct amount of refresh tokens stored and the excess ones are revoked")

	logger.Info("delete refresh token")
	deletedRefToken, err := refreshTokenAdapter.Delete(ctx, domain.RefreshTokenDeleteDTOInput{
		Id: accountRefreshTokensRes.Tokens[1].Id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %v", err)
	}
	logger.Info("deleted refresh token")

	logger.Info("list refresh tokens")
	accountRefreshTokensRes, err = refreshTokenAdapter.List(ctx, domain.RefreshTokenListDTOInput{
		AccountId: resp.Id,
	})
	logger.Info("listed refresh tokens")

	if slices.ContainsFunc(accountRefreshTokensRes.Tokens, func(token domain.RefreshTokenListDTOOutputToken) bool {
		return token.Id == deletedRefToken.Id
	}) {
		return errors.New("refresh token must have been deleted, but it hasn't")
	}

	logger.Info("delete refresh tokens by account id")
	deleteAccountTokensRes, err := refreshTokenAdapter.DeleteByAccountId(ctx, domain.RefreshTokenDeleteByAccountIdDTOInput{
		Id: accByEmail.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete account by id: %v", err)
	}
	logger.Info("deleted refresh tokens by account id")

	if len(accountRefreshTokensRes.Tokens) != len(deleteAccountTokensRes.Ids) {
		return errors.New("amount of deleted account refresh tokens must match the amount of last refresh tokens listed the last time")
	}

	accountRefreshTokensRes, err = refreshTokenAdapter.List(ctx, domain.RefreshTokenListDTOInput{
		AccountId: accountId,
	})

	if len(accountRefreshTokensRes.Tokens) != 0 {
		return errors.New("there must be no refresh tokens for account")
	}

	adminEmail := fmt.Sprintf("admin-%d@admin.com", time.Now().Unix())
	adminPassword := uuid.New().String()[:24]
	logger.Info("create admin")
	createAdminRes, err := svc.CreateAdmin(ctx, domain.CreateAdminReq{
		Name:     fmt.Sprintf("Dan %d", time.Now().Unix()),
		Password: adminPassword,
		Email:    adminEmail,
	})
	if err != nil {
		return fmt.Errorf("failed to create admin account: %v", err)
	}
	logger.Info("created admin", zap.Any("admin", createAdminRes))

	logger.Info("activate admin account")
	_, err = svc.ActivateAccounts(ctx, domain.ActivateAccountsReq{Emails: []string{createAdminRes.Email}})
	if err != nil {
		return fmt.Errorf("failed to active admin account: %v", err)
	}
	logger.Info("activated admin account")

	logger.Info("authneticate admin")
	authneticateAdminRes, err := svc.Authenticate(ctx, domain.AuthenticateReq{
		Email:    adminEmail,
		Password: adminPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate admin: %v", err)
	}
	logger.Info("authneticated admin")

	logger.Info("create access token for admin")
	accessTokenRes, err := svc.CreateAccessToken(ctx, domain.CreateAccessTokenReq{
		RefreshToken: authneticateAdminRes.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to create access token for admin: %v", err)
	}
	logger.Info("created access token for admin")

	sellerEmail := fmt.Sprintf("seller-%d@seller.com", time.Now().Unix())
	sellerPassword := uuid.New().String()[:24]
	logger.Info("create seller using admin's access token")
	createSellerRes, err := svc.CreateSeller(ctx, domain.CreateSellerReq{
		Name:        "seller",
		Email:       sellerEmail,
		Password:    sellerPassword,
		AccessToken: accessTokenRes.AccessToken,
	})
	if err != nil {
		return fmt.Errorf("failed to create seller using admin's access token", err)
	}
	logger.Info("created seller using admin's access token", zap.Any("seller", createSellerRes))

	logger.Info("all tests passed")
	return nil
}
