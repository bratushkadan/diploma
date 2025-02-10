package main

import (
	"context"
	"log"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ydb-platform/ydb-go-sdk/v3"

	account_creation_daemon_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/account-creation/daemon"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/bratushkadan/floral/pkg/ymq"
)

var (
	ydbFullEndpoint string
	authMethod      string

	sqsQueueUrl        string
	sqsAccessKeyId     string
	sqsSecretAccessKey string
)

const (
	EnvKeySqsQueueUrl        = "SQS_QUEUE_URL"
	EnvKeySqsAccessKeyId     = "SQS_ACCESS_KEY_ID"
	EnvKeySqsSecretAccessKey = "SQS_SECRET_ACCESS_KEY"
)

type DummyAccountCreationNotificationProvider struct {
}

func (p DummyAccountCreationNotificationProvider) Send(_ context.Context, _ domain.SendAccountCreationNotificationDTOInput) (domain.SendAccountCreationNotificationDTOOutput, error) {
	return domain.SendAccountCreationNotificationDTOOutput{}, nil
}

var _ domain.AccountCreationNotificationProvider = (*DummyAccountCreationNotificationProvider)(nil)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}
}

func init() {
	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")
	authMethod = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")

	sqsQueueUrl = cfg.MustEnv(EnvKeySqsQueueUrl)
	sqsAccessKeyId = cfg.MustEnv(EnvKeySqsAccessKeyId)
	sqsSecretAccessKey = cfg.MustEnv(EnvKeySqsSecretAccessKey)
}

func main() {
	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := conf.Build()
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

	sqsClient, err := ymq.New(ctx, sqsAccessKeyId, sqsSecretAccessKey, sqsQueueUrl, logger)
	if err != nil {
		logger.Fatal("failed to build new ymq", zap.Error(err))
	}

	svc, err := service.NewAuthBuilder().
		Logger(logger).
		AccountProvider(accountProvider).
		Build()
	if err != nil {
		logger.Fatal("failed to build new auth", zap.Error(err))
	}

	daemon, err := account_creation_daemon_adapter.NewBuilder().
		AuthService(svc).
		Logger(logger).
		SqsClient(sqsClient).
		SqsQueueUrl(sqsQueueUrl).
		Build()
	if err != nil {
		logger.Fatal("failed to build account creation sqs daemon adapter", zap.Error(err))
	}

	if err := daemon.ReceiveProcessAccountCreationMessages(ctx); err != nil {
		logger.Fatal("error running process account creation daemon", zap.Error(err))
	}
}
