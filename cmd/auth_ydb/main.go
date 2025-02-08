package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/infrastructure/authn"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ydb-platform/ydb-go-sdk/v3"
)

var (
	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")
	authMethod      = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")
)

const (
	EnvKeyAccountIdHashSalt       = "APP_ID_ACCOUNT_HASH_SALT"
	EnvKeyTokenIdHashSalt         = "APP_ID_TOKEN_HASH_SALT"
	EnvKeyPasswordHashSalt        = "APP_PASSWORD_HASH_SALT"
	EnvKeyAuthTokenPrivateKeyPath = "APP_AUTH_TOKEN_PRIVATE_KEY_PATH"
	EnvKeyAuthTokenPublicKeyPath  = "APP_AUTH_TOKEN_PUBLIC_KEY_PATH"
)

type DummyAccountConfirmationProvider struct {
}

func (p DummyAccountConfirmationProvider) Send(_ context.Context, _ domain.SendAccountConfirmationDTOInput) (domain.SendAccountConfirmationDTOOutput, error) {
	return domain.SendAccountConfirmationDTOOutput{}, nil
}

var _ domain.AccountConfirmationProvider = (*DummyAccountConfirmationProvider)(nil)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	// conf := zap.NewProductionConfig()
	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := conf.Build()
	if err != nil {
		log.Fatal("Error setting up zap")
	}

	accountIdHasher, err := idhash.New(os.Getenv(EnvKeyAccountIdHashSalt), idhash.WithPrefix("ie"))
	if err != nil {
		log.Fatal(err)
	}
	tokenIdHasher, err := idhash.New(os.Getenv(EnvKeyTokenIdHashSalt), idhash.WithPrefix("rb"))
	if err != nil {
		log.Fatal(err)
	}
	passwordHasher, err := auth.NewPasswordHasher(os.Getenv(EnvKeyPasswordHashSalt))
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
		PrivateKeyPath(os.Getenv(EnvKeyAuthTokenPrivateKeyPath)).
		PublicKeyPath(os.Getenv(EnvKeyAuthTokenPublicKeyPath)).
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

	authSvc, err := service.NewAuthBuilder().
		AccountProvider(accountAdapter).
		RefreshTokenProvider(refreshTokenAdapter).
		TokenProvider(tokenProvider).
		AccountConfirmationProvider(DummyAccountConfirmationProvider{}).
		Logger(zap.NewNop()).
		Build()
	if err != nil {
		logger.Fatal("could not build auth svc", zap.Error(err))
	}

	if err := runAccountTests(ctx, accountIdHasher, logger, authSvc, accountAdapter); err != nil {
		logger.Fatal("failed to run account service tests", zap.Error(err))
	}
}

func runAccountTests(ctx context.Context, accountIdHasher idhash.IdHasher, logger *zap.Logger, svc domain.AuthService, accAdapter domain.AccountProvider) error {
	logger.Info("create account")
	email := fmt.Sprintf(`someemail-%d@gmail.com`, time.Now().UnixMilli())
	password := "ooga"
	resp, err := svc.CreateAccount(ctx, domain.CreateAccountReq{
		Name:     "Danila",
		Email:    email,
		Password: password,
		Type:     "user",
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

	logger.Info("find account")
	acc, err := accAdapter.FindAccount(ctx, domain.FindAccountDTOInput{Id: resp.Id})
	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}
	if acc != nil {
		logger.Info("found account", zap.Any("account", acc))
	} else {
		logger.Info("account not found", zap.String("id", resp.Id))
	}

	logger.Info("find account by email")
	accByEmail, err := accAdapter.FindAccountByEmail(ctx, domain.FindAccountByEmailDTOInput{Email: "someemail-1738903445714@gmail.com"})
	if err != nil {
		return fmt.Errorf("failed to find account by email: %w", err)
	}
	if acc != nil {
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

	return nil
}
