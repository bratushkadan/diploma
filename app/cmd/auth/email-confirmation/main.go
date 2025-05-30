package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	email_confirmation_http_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/email-confirmation/http"
	ydb_dynamodb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/dynamodb"
	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/bratushkadan/floral/pkg/ymq"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	Port = cfg.EnvDefault("PORT", "8080")
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	ymqTriggerEndpointsEnabled, err := strconv.ParseBool(cfg.EnvDefault(setup.EnvKeyYmqTriggerHttpEndpointsEnabled, "0"))
	if err != nil {
		logger.Fatal("failed to parse ymq trigger http endpoints enabled from env", zap.String("env_key", setup.EnvKeyYmqTriggerHttpEndpointsEnabled), zap.Error(err))
	}

	ydbDocApiEndpoint := cfg.MustEnv(setup.EnvKeyYdbDocApiEndpoint)

	accessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	secretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	tokens, err := ydb_dynamodb_adapter.NewEmailConfirmationTokens(ctx, accessKeyId, secretAccessKey, ydbDocApiEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to setup email confirmation tokens ydb dynamodb", zap.Error(err))
	}

	b := service.
		NewEmailConfirmationBuilder().
		Logger(logger).
		Tokens(tokens)

	if _, ok := os.LookupEnv(setup.EnvKeySqsQueueUrlEmailConfirmations); ok {
		sqsQueueUrl := cfg.MustEnv(setup.EnvKeySqsQueueUrlEmailConfirmations)
		sqsClient, err := ymq.New(ctx, accessKeyId, secretAccessKey, sqsQueueUrl, logger)
		if err != nil {
			logger.Fatal("failed to setup ymq sqs client for publishing email confirmation messages", zap.Error(err))
		}

		notifications := &ymq_adapter.EmailConfirmation{
			Sqs:         sqsClient,
			SqsQueueUrl: sqsQueueUrl,
		}

		b = b.Notifications(notifications)
	}

	if _, ok := os.LookupEnv(setup.EnvKeySenderEmail); ok {
		senderEmail := cfg.MustEnv(setup.EnvKeySenderEmail)
		senderPassword := cfg.MustEnv(setup.EnvKeySenderPassword)
		endpoint := cfg.MustEnv(setup.EnvKeyEmailConfirmationApiEndpoint)
		origin := os.Getenv(setup.EnvKeyEmailConfirmationOrigin)

		senderB := email_confirmer.NewBuilder()

		if origin == "" {
			senderB = senderB.ConfirmationHostCtxResolver(endpoint)
		} else {
			senderB = senderB.StaticConfirmationUrl(fmt.Sprintf("%s/%s", origin, strings.TrimPrefix(endpoint, "/")))
		}

		sender, err := senderB.
			SenderEmail(senderEmail).
			SenderPassword(senderPassword).
			Build()
		if err != nil {
			logger.Fatal("failed to setup email confirmations sender", zap.Error(err))
		}

		b = b.Sender(sender)
	}

	svc, err := b.Build()
	if err != nil {
		logger.Fatal("failed to setup auth service", zap.Error(err))
	}

	httpAdapter := email_confirmation_http_adapter.New(svc, logger)

	r := chi.NewRouter()
	// r.Use(middleware.Logger)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r)
		})
	})

	r.Get("/ready", xhttp.HandleReadiness(ctx))
	r.Get("/health", xhttp.HandleReadiness(ctx))

	v1ApiRouter := chi.NewRouter()

	apiRouter := chi.NewRouter()
	apiRouter.Mount("/v1", v1ApiRouter)
	r.Mount("/api", apiRouter)

	// FIXME: delete after tests
	v1ApiRouter.Get("/auth/confirmEmail", func(w http.ResponseWriter, _ *http.Request) {
		w.Write(confirmEmailGetHtmlPage)
	})
	v1ApiRouter.Post("/auth/confirmEmail", httpAdapter.HandleConfirmEmail)
	v1ApiRouter.Post("/auth/send-confirmation-email", httpAdapter.HandleSendConfirmation)

	if ymqTriggerEndpointsEnabled {
		logger.Debug("Yandex Cloud YMQ Trigger endpoints enabled")
		v1ApiRouter.Post("/auth:send-confirmation-email-trigger", httpAdapter.HandleSendConfirmationYmqTrigger)
	}

	r.NotFound(xhttp.HandleNotFound())

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", Port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      r,
	}
	server.RegisterOnShutdown(doCleanup)

	go func() {
		<-ctx.Done()

		logger.Info("got shutdown signal")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("error while stopping http listener", zap.Error(err))
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(http.ErrServerClosed, err) {
			logger.Fatal("failed to listen and serve", zap.Error(err))
		}
	}
}

func doCleanup() {}

var confirmEmailGetHtmlPage = []byte(`<!DOCTYPE html>
<html>
  <head>
    <link rel="icon" href="data:,">
  </head>
  <body>
    <script defer>
      function report(message) {
        window.document.body.innerHTML = '<div>' + message + '</div>'
      }
      function reportProblem(message) {
        report('Error: ' + message)
      }
      function reportSuccess(message) {
        report(message)
      }

      document.addEventListener("DOMContentLoaded", main)

      async function main() {
        if (new URLSearchParams(window.location.search).get("token") === null) {
          reportProblem("Failed to confirm email: token query parameter must be set")
        }
        try {
          const response = await fetch(window.location.origin + window.location.pathname, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({token: new URLSearchParams(window.location.search).get("token")}),
          })
          if (response.ok) {
            reportSuccess("Successfuly confirmed email.")
            const responseBody = await response.json()
            console.log(responseBody)
            return
          }

          const responseBody = await response.json()
          if ('errors' in responseBody) {
            reportProblem(JSON.stringify(responseBody.errors, null, 2))
            return
          }

          console.error("Unknown upstream server error format", data)
          throw new Error("Unknown upstream server error format")
        } catch (err) {
            console.error(err)
            reportProblem("Could not confirm email (server internal error).")
        }
      }
    </script>
  </body>
</html>`)
