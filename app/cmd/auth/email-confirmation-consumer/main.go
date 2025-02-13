package main

import (
	"context"
	"log"

	email_confirmation_daemon_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/email-confirmation/daemon"
	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/joho/godotenv"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/pkg/ymq"
)

type DummyAccountCreationNotificationProvider struct {
}

func (p DummyAccountCreationNotificationProvider) Send(_ context.Context, _ domain.SendAccountCreationNotificationDTOInput) (domain.SendAccountCreationNotificationDTOOutput, error) {
	return domain.SendAccountCreationNotificationDTOOutput{}, nil
}

var _ domain.AccountCreationNotifications = (*DummyAccountCreationNotificationProvider)(nil)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	ydbFullEndpoint := cfg.MustEnv(setup.EnvKeyYdbEndpoint)
	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, "metadata")

	sqsQueueUrl := cfg.MustEnv(setup.EnvKeySqsQueueUrlEmailConfirmations)
	accessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	secretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	logger, err := logging.NewZapConf("dev").Build()
	if err != nil {
		log.Fatal("Error setting up zap")
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

	accountProvider := ydb_adapter.NewAccount(ydb_adapter.AccountConf{
		DbDriver: db,
		Logger:   logger,
	})

	sqsClient, err := ymq.New(ctx, accessKeyId, secretAccessKey, sqsQueueUrl, logger)
	if err != nil {
		logger.Fatal("failed to build new ymq", zap.Error(err))
	}

	svc, err := service.NewAuthBuilder().
		AccountProvider(accountProvider).
		Logger(logger).
		Build()
	if err != nil {
		logger.Fatal("failed to build new auth", zap.Error(err))
	}

	daemon, err := email_confirmation_daemon_adapter.NewBuilder().
		Service(svc).
		Logger(logger).
		SqsClient(sqsClient).
		SqsQueueUrl(sqsQueueUrl).
		Build()
	if err != nil {
		logger.Fatal("failed to build account creation sqs daemon adapter", zap.Error(err))
	}

	if err := daemon.ReceiveProcessEmailConfirmationMessages(ctx); err != nil {
		logger.Fatal("error running process account creation daemon", zap.Error(err))
	}
}
