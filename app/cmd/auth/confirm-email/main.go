package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	email_confirmation_http_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/email-confirmation/http"
	ydb_dynamodb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/dynamodb"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/httpmock"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/bratushkadan/floral/pkg/ymq"
	"go.uber.org/zap"
)

var (
	emailConfirmationSvc *email_confirmation_http_adapter.Adapter
	token                = os.Getenv("CONFIRMATION_TOKEN")
)

func main() {
	w := httpmock.NewMockResponseWriter()
	r := &http.Request{
		Body: &httpmock.ReadCloser{Reader: strings.NewReader(fmt.Sprintf(`{"token": "%s"}`, token))},
	}
	Handler(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	emailConfirmationSvc.HandleConfirmEmail(w, r)
}

func init() {
	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	ydbDocApiEndpoint := cfg.MustEnv(setup.EnvKeyYdbDocApiEndpoint)

	accessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	secretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	sqsQueueUrlEmailConfirmations := cfg.MustEnv(setup.EnvKeySqsQueueUrlEmailConfirmations)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tokens, err := ydb_dynamodb_adapter.NewEmailConfirmationTokens(ctx, accessKeyId, secretAccessKey, ydbDocApiEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to setup email confirmation tokens ydb dynamodb", zap.Error(err))
	}

	emailConfirmationSqsEndpoint := sqsQueueUrlEmailConfirmations
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
		Tokens(tokens).
		Notifications(emailConfirmationNotifications).
		Build()
	if err != nil {
		logger.Fatal("failed to setup auth service", zap.Error(err))
	}

	emailConfirmationSvc = email_confirmation_http_adapter.New(svc, logger)
}
