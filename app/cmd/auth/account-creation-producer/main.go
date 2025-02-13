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

const (
	// Email address to send confirmation to
	EnvKeyTargetEmail = "TARGET_EMAIL"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	ydbFullEndpoint := cfg.MustEnv(setup.EnvKeyYdbEndpoint)

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, "metadata")

	sqsQueueUrl := cfg.MustEnv(setup.EnvKeySqsQueueUrlAccountCreations)
	sqsAccessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	sqsSecretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	targetEmail := cfg.MustEnv(EnvKeyTargetEmail)

	logger, err := logging.NewZapConf("dev").Build()
	if err != nil {
		log.Fatal("Error setting up zap")
	}

	accountIdHasher, err := idhash.New(os.Getenv(setup.EnvKeyAccountIdHashSalt), idhash.WithPrefix("ie"))
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

	if err := runCreateAccount(ctx, authSvc, logger, targetEmail); err != nil {
		logger.Fatal("failed to run create account", zap.Error(err))
	}
}

func runCreateAccount(ctx context.Context, svc domain.AuthService, logger *zap.Logger, targetEmail string) error {
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
