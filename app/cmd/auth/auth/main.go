package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, "metadata")

	ydbFullEndpoint := cfg.MustEnv(setup.EnvKeyYdbEndpoint)

	sqsQueueUrl := cfg.MustEnv(setup.EnvKeySqsQueueUrlAccountCreations)
	sqsAccessKeyId := cfg.MustEnv(setup.EnvKeyAwsAccessKeyId)
	sqsSecretAccessKey := cfg.MustEnv(setup.EnvKeyAwsSecretAccessKey)

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	accountIdHasher, err := idhash.New(os.Getenv(setup.EnvKeyAccountIdHashSalt), idhash.WithPrefix("ie"))
	if err != nil {
		logger.Fatal("failed to set up account id hasher")
	}
	tokenIdHasher, err := idhash.New(os.Getenv(setup.EnvKeyTokenIdHashSalt), idhash.WithPrefix("rb"))
	if err != nil {
		logger.Fatal("failed to set up token id hasher")
	}
	passwordHasher, err := auth.NewPasswordHasher(os.Getenv(setup.EnvKeyPasswordHashSalt))
	if err != nil {
		logger.Fatal("failed to set up password hasher", zap.Error(err))
	}

	logger.Debug("setup ydb")
	db, err := ydb.Open(ctx, ydbFullEndpoint, ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	logger.Debug("set up ydb")
	defer func() {
		if err := db.Close(ctx); err != nil {
			logger.Error("failed to close ydb", zap.Error(err))
		}
	}()

	tokenProvider, err := authn.NewTokenProviderBuilder().
		PrivateKeyPath(os.Getenv(setup.EnvKeyAuthTokenPrivateKeyPath)).
		PublicKeyPath(os.Getenv(setup.EnvKeyAuthTokenPublicKeyPath)).
		Build()
	if err != nil {
		logger.Fatal("failed to setup token provider", zap.Error(err))
	}
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

	sqsClient, err := ymq.New(ctx, sqsAccessKeyId, sqsSecretAccessKey, sqsQueueUrl, logger)
	if err != nil {
		logger.Fatal("failed to setup ymq")
	}
	accountCreationNotificationAdapter := ymq_adapter.AccountCreation{
		Sqs:         sqsClient,
		SqsQueueUrl: sqsQueueUrl,
	}

	svc, err := service.NewAuthBuilder().
		AccountProvider(accountAdapter).
		RefreshTokenProvider(refreshTokenAdapter).
		TokenProvider(tokenProvider).
		AccountCreationNotificationProvider(&accountCreationNotificationAdapter).
		Logger(zap.NewNop()).
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

	rApi := chi.NewRouter()
	rV1 := chi.NewRouter()
	rUsers := chi.NewRouter()

	r.Mount("/api", rApi)
	rApi.Mount("/v1", rV1)
	rV1.Mount("/users", rUsers)

	rUsers.Post("/:register", http.HandlerFunc(httpAdapter.RegisterUserHandler))
	rUsers.Post("/:registerSeller", http.HandlerFunc(httpAdapter.RegisterSellerHandler))
	rUsers.Post("/:registerAdmin", http.HandlerFunc(httpAdapter.RegisterAdminHandler))
	rUsers.Post("/:authenticate", http.HandlerFunc(httpAdapter.AuthenticateHandler))
	rUsers.Post("/:renewRefreshToken", http.HandlerFunc(httpAdapter.RenewRefreshTokenHandler))
	rUsers.Post("/:createAccessToken", http.HandlerFunc(httpAdapter.CreateAccessToken))

	http.ListenAndServe(":8080", r)

}
