package main

import (
	"context"
	"log"
	"os"

	ydb_dynamodb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/dynamodb"
	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	email_confirmation_daemon_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/email-confirmations/daemon"
	"github.com/bratushkadan/floral/pkg/ymq"
)

var (
	ydbDocApiEndpoint string

	accessKeyId     string
	secretAccessKey string

	sqsQueueUrlEmailConfirmations string
	sqsQueueUrlAccountCreations   string

	senderEmail                  string
	senderPassword               string
	emailConfirmationApiEndpoint string
)

const (
	EnvKeyYdbDocApiEndpoint = "YDB_DOC_API_ENDPOINT"

	EnvKeyAwsAccessKeyId     = "AWS_ACCESS_KEY_ID"
	EnvKeyAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"

	EnvKeySqsQueueUrlEmailConfirmations = "SQS_QUEUE_URL_EMAIL_CONFIRMATIONS"
	EnvKeySqsQueueUrlAccountCreations   = "SQS_QUEUE_URL_ACCOUNT_CREATIONS"

	EnvKeySenderEmail    = "SENDER_EMAIL"
	EnvKeySenderPassword = "SENDER_PASSWORD"
	// EnvKeyEmailConfirmationApiEndpoint = "EMAIL_CONFIRMATION_API_ENDPOINT"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	ydbDocApiEndpoint = cfg.MustEnv(EnvKeyYdbDocApiEndpoint)

	accessKeyId = cfg.MustEnv(EnvKeyAwsAccessKeyId)
	secretAccessKey = cfg.MustEnv(EnvKeyAwsSecretAccessKey)

	sqsQueueUrlEmailConfirmations = cfg.MustEnv(EnvKeySqsQueueUrlEmailConfirmations)
	sqsQueueUrlAccountCreations = cfg.MustEnv(EnvKeySqsQueueUrlAccountCreations)

	senderEmail = cfg.MustEnv(EnvKeySenderEmail)
	senderPassword = cfg.MustEnv(EnvKeySenderPassword)
	// emailConfirmationApiEndpoint = cfg.MustEnv(EnvKeyEmailConfirmationApiEndpoint)

	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := conf.Build()
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
		StaticConfirmationUrl("http://localhost:8080/confirm").
		/*StaticConfirmationUrl(emailConfirmationApiEndpoint).*/
		Build()
	if err != nil {
		logger.Fatal("failed to setup email confirmations sender", zap.Error(err))
	}

	accountCreatedSqsEndpoint := os.Getenv(EnvKeySqsQueueUrlAccountCreations)
	accountCreatedSqsClient, err := ymq.New(ctx, accessKeyId, secretAccessKey, accountCreatedSqsEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to build new ymq", zap.Error(err))
	}

	emailConfirmationSqsEndpoint := os.Getenv(EnvKeySqsQueueUrlEmailConfirmations)
	emailConfirmationSqsClient, err := ymq.New(ctx, accessKeyId, secretAccessKey, emailConfirmationSqsEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to setup ymq sqs client for publishing email confirmation messages", zap.Error(err))
	}

	emailConfirmationNotifications := &ymq_adapter.EmailConfirmation{
		Sqs:         emailConfirmationSqsClient,
		SqsQueueUrl: emailConfirmationSqsEndpoint,
	}

	svc, err := service.NewEmailConfirmationBuilder().
		Logger(logger).
		Sender(sender).
		Tokens(tokens).
		Notifications(emailConfirmationNotifications).
		Build()
	if err != nil {
		logger.Fatal("failed to build new email confirmation service", zap.Error(err))
	}

	daemon, err := email_confirmation_daemon_adapter.NewBuilder().
		Service(svc).
		Logger(logger).
		SqsClient(accountCreatedSqsClient).
		SqsQueueUrl(accountCreatedSqsEndpoint).
		Build()
	if err != nil {
		logger.Fatal("failed to build account confirmation sqs daemon adapter", zap.Error(err))
	}

	if err := daemon.ReceiveProcessAccountCreationMessages(ctx); err != nil {
		logger.Fatal("error running process account confirmation daemon", zap.Error(err))
	}
}
