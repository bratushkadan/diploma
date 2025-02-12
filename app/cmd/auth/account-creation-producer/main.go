package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/bratushkadan/floral/pkg/ymq"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ydb-platform/ydb-go-sdk/v3"
)

var (
	ydbFullEndpoint string
	authMethod      = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")

	sqsQueueUrl        string
	sqsAccessKeyId     string
	sqsSecretAccessKey string

	targetEmail string
)

const (
	EnvKeyAccountIdHashSalt = "APP_ID_ACCOUNT_HASH_SALT"
	EnvKeyPasswordHashSalt  = "APP_PASSWORD_HASH_SALT"

	EnvKeySqsQueueUrl        = "SQS_QUEUE_URL"
	EnvKeySqsAccessKeyId     = "SQS_ACCESS_KEY_ID"
	EnvKeySqsSecretAccessKey = "SQS_SECRET_ACCESS_KEY"

	// Email address to send confirmation to
	EnvKeyTargetEmail = "TARGET_EMAIL"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")

	sqsQueueUrl = cfg.MustEnv(EnvKeySqsQueueUrl)
	sqsAccessKeyId = cfg.MustEnv(EnvKeySqsAccessKeyId)
	sqsSecretAccessKey = cfg.MustEnv(EnvKeySqsSecretAccessKey)

	targetEmail = cfg.MustEnv(EnvKeyTargetEmail)

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

	if err != nil {
		logger.Fatal("failed to setup token provider", zap.Error(err))
	}
	accountAdapter := ydb_adapter.NewAccount(ydb_adapter.AccountConf{
		DbDriver:       db,
		IdHasher:       accountIdHasher,
		PasswordHasher: passwordHasher,
		Logger:         logger,
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
		AccountCreationNotificationProvider(&accountCreationNotificationAdapter).
		Logger(zap.NewNop()).
		Build()
	if err != nil {
		logger.Fatal("could not build auth svc", zap.Error(err))
	}

	if err := runCreateAccount(ctx, authSvc, logger); err != nil {
		logger.Fatal("failed to run create account", zap.Error(err))
	}
}

func runCreateAccount(ctx context.Context, svc domain.AuthService, logger *zap.Logger) error {
	logger.Info("create account")
	name := uuid.New().String()[:24]
	password := uuid.New().String()[:24]
	_, err := svc.CreateUser(ctx, domain.CreateUserReq{
		Name:     name,
		Password: password,
		Email:    targetEmail,
	})
	if err != nil {
		return fmt.Errorf("failed to create account")
	}
	return nil
}
