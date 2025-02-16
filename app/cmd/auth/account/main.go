package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	http_adapter "github.com/bratushkadan/floral/internal/auth/adapters/primary/auth/http"
	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	ymq_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ymq"
	"github.com/bratushkadan/floral/internal/auth/infrastructure/authn"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	"github.com/bratushkadan/floral/pkg/xhttp"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/bratushkadan/floral/pkg/ymq"
	"github.com/go-chi/chi/v5"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/zap"
)

var (
	Port = cfg.EnvDefault("PORT", "8080")
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, ydbpkg.YdbAuthMethodMetadata)

	env := cfg.AssertEnv(
		setup.EnvKeyYdbEndpoint,
		setup.EnvKeySqsQueueUrlAccountCreations,
		setup.EnvKeyAwsAccessKeyId,
		setup.EnvKeyAwsSecretAccessKey,
		setup.EnvKeyAccountIdHashSalt,
		setup.EnvKeyTokenIdHashSalt,
		setup.EnvKeyPasswordHashSalt,
		setup.EnvKeyAuthTokenPublicKey,
		setup.EnvKeyAuthTokenPrivateKey,
	)

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	accountIdHasher, err := idhash.New(env[setup.EnvKeyAccountIdHashSalt], idhash.WithPrefix("ie"))
	if err != nil {
		logger.Fatal("failed to set up account id hasher")
	}
	tokenIdHasher, err := idhash.New(env[setup.EnvKeyTokenIdHashSalt])
	if err != nil {
		logger.Fatal("failed to set up token id hasher")
	}
	passwordHasher, err := auth.NewPasswordHasher((env[setup.EnvKeyPasswordHashSalt]))
	if err != nil {
		logger.Fatal("failed to set up password hasher", zap.Error(err))
	}

	tokenProvider, err := authn.NewTokenProviderBuilder().
		PublicKey([]byte(env[setup.EnvKeyAuthTokenPublicKey])).
		PrivateKey([]byte(env[setup.EnvKeyAuthTokenPrivateKey])).
		Build()
	if err != nil {
		logger.Fatal("failed to setup token provider", zap.Error(err))
	}

	logger.Debug("setup ydb")
	db, err := ydb.Open(ctx, env[setup.EnvKeyYdbEndpoint], ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	logger.Debug("set up ydb")
	defer func() {
		if err := db.Close(ctx); err != nil {
			logger.Error("failed to close ydb", zap.Error(err))
		}
	}()

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

	sqsClient, err := ymq.New(
		ctx,
		env[setup.EnvKeyAwsAccessKeyId],
		env[setup.EnvKeyAwsSecretAccessKey],
		env[setup.EnvKeySqsQueueUrlAccountCreations],
		logger,
	)
	if err != nil {
		logger.Fatal("failed to setup ymq", zap.Error(err))
	}
	accountCreationNotificationAdapter := ymq_adapter.AccountCreation{
		Sqs:         sqsClient,
		SqsQueueUrl: env[setup.EnvKeySqsQueueUrlAccountCreations],
	}

	svc, err := service.NewAuthBuilder().
		AccountProvider(accountAdapter).
		RefreshTokenProvider(refreshTokenAdapter).
		TokenProvider(tokenProvider).
		AccountCreationNotificationProvider(&accountCreationNotificationAdapter).
		Logger(logger).
		Build()
	if err != nil {
		logger.Fatal("failed to setup auth service", zap.Error(err))
	}

	httpAdapter, err := http_adapter.NewBuilder().
		Logger(logger).
		Svc(svc).
		Build()
	if err != nil {
		logger.Fatal("failed to setup auth http adapter", zap.Error(err))
	}

	r := chi.NewRouter()
	rUsers := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r)
		})
	})

	rUsers.Post("/:createAccount", http.HandlerFunc(httpAdapter.RegisterUserHandler))
	rUsers.Post("/:createSellerAccount", http.HandlerFunc(httpAdapter.RegisterSellerHandler))
	// Expose this endpoint ONLY internally
	rUsers.Post("/:createAdminAccount", http.HandlerFunc(httpAdapter.RegisterAdminHandler))
	// Expose this endpoint ONLY internally
	rUsers.Post("/:activateAccounts", http.HandlerFunc(httpAdapter.ActivateAccountsHandler))
	rUsers.Post("/:authenticate", http.HandlerFunc(httpAdapter.AuthenticateHandler))
	rUsers.Post("/:replaceRefreshToken", http.HandlerFunc(httpAdapter.ReplaceRefreshTokenHandler))
	rUsers.Post("/:createAccessToken", http.HandlerFunc(httpAdapter.CreateAccessToken))

	// Get
	// rUsers.Get("/{id}")
	// List
	// rUsers.Get("/")

	r.Get("/ready", xhttp.HandleReadiness(ctx))
	r.Get("/health", xhttp.HandleReadiness(ctx))
	r.NotFound(xhttp.HandleNotFound())

	rApi := chi.NewRouter()
	rV1 := chi.NewRouter()
	r.Mount("/api", rApi)
	rApi.Mount("/v1", rV1)
	rV1.Mount("/users", rUsers)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", Port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      r,
	}
	server.RegisterOnShutdown(func() {})

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
