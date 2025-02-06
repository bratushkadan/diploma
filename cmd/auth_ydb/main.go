package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
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
	EnvKeyIdHashSalt       = "APP_ID_HASH_SALT"
	EnvKeyPasswordHashSalt = "APP_PASSWORD_HASH_SALT"
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

	idHasher, err := idhash.New(os.Getenv(EnvKeyIdHashSalt), idhash.WithPrefix("ie"))
	if err != nil {
		log.Fatal(err)
	}
	passwordHasher, err := auth.NewPasswordHasher(os.Getenv(EnvKeyPasswordHashSalt))
	if err != nil {
		logger.Fatal("failed to set up password hasher", zap.Error(err))
	}

	ctx := context.Background()

	logger.Info("setup ydb")
	db, err := ydb.Open(ctx, ydbFullEndpoint, ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("set up ydb")
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Print()
		}
	}()
	accountAdapter := ydb_adapter.New(ydb_adapter.Conf{
		DbDriver:       db,
		IdHasher:       idHasher,
		PasswordHasher: passwordHasher,

		Logger: logger,
	})

	authSvc, err := service.NewAuthBuilder().
		AccountProvider(accountAdapter).
		AccountConfirmationProvider(DummyAccountConfirmationProvider{}).
		Build()
	if err != nil {
		logger.Fatal("could not build auth svc", zap.Error(err))
	}

	logger.Info("create account")
	resp, err := authSvc.CreateAccount(ctx, domain.CreateAccountReq{
		Name:     "Danila",
		Email:    fmt.Sprintf(`someemail-%d@gmail.com`, time.Now().UnixMilli()),
		Password: "ooga",
		Type:     "user",
	})
	if err != nil {
		logger.Error("error creating account", zap.Error(err))
	} else {
		logger.Info("response creating account", zap.Any("response", resp))

		idInt64, err := idHasher.DecodeInt64(resp.Id)
		if err != nil {
			logger.Fatal("failed to decode str id to in64", zap.String("str_id", resp.Id), zap.Error(err))
		}
		logger.Info("decoded string id to int64", zap.String("str_id", resp.Id), zap.Int64("id", idInt64))
	}
}
