package main

import (
	"context"
	"fmt"
	"log"

	account_creation_daemon_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/account-creation/daemon"
	ydb_dynamodb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/dynamodb"
	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/pkg/ymq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	ydbDocApiEndpoint := cfg.MustEnv(setup.EnvKeyYdbDocApiEndpoint)

	accessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	secretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	sqsQueueUrlAccountCreations := cfg.MustEnv(setup.EnvKeySqsQueueUrlAccountCreations)

	senderEmail := cfg.MustEnv(setup.EnvKeySenderEmail)
	senderPassword := cfg.MustEnv(setup.EnvKeySenderPassword)

	emailConfirmationApiEndpoint := cfg.EnvDefault(setup.EnvKeyEmailConfirmationApiEndpoint, "/api/v1/auth/confirm-email")

	logger, err := logging.NewZapConf("dev").Build()
	if err != nil {
		log.Fatal("Error setting up zap")
	}

	ctx := context.Background()

	tokens, err := ydb_dynamodb_adapter.NewEmailConfirmationTokens(ctx, accessKeyId, secretAccessKey, ydbDocApiEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to setup email confirmation tokens ydb dynamodb", zap.Error(err))
	}

	sender, err := email_confirmer.NewBuilder().
		SenderEmail(senderEmail).
		SenderPassword(senderPassword).
		StaticConfirmationUrl(fmt.Sprintf("http://localhost:8080%s", emailConfirmationApiEndpoint)).
		Build()
	if err != nil {
		logger.Fatal("failed to setup email confirmations sender", zap.Error(err))
	}

	accountCreationSqsEndpoint := sqsQueueUrlAccountCreations
	accountCreationSqsClient, err := ymq.New(ctx, accessKeyId, secretAccessKey, accountCreationSqsEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to build new ymq", zap.Error(err))
	}

	svc, err := service.NewEmailConfirmationBuilder().
		Logger(logger).
		Sender(sender).
		Tokens(tokens).
		Build()
	if err != nil {
		logger.Fatal("failed to build new email confirmation service", zap.Error(err))
	}

	daemon, err := account_creation_daemon_adapter.NewBuilder().
		Logger(logger).
		Service(svc).
		SqsClient(accountCreationSqsClient).
		SqsQueueUrl(accountCreationSqsEndpoint).
		Build()
	if err != nil {
		logger.Fatal("failed to build account confirmation sqs daemon adapter", zap.Error(err))
	}

	if err := daemon.ReceiveProcessAccountCreationMessages(ctx); err != nil {
		logger.Fatal("error running process account confirmation daemon", zap.Error(err))
	}
}
