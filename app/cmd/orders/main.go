package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/internal/orders/presentation"
	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/service"
	"github.com/bratushkadan/floral/internal/orders/store"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	xgin "github.com/bratushkadan/floral/pkg/xhttp/gin"
	"github.com/bratushkadan/floral/pkg/xhttp/gin/middleware/auth"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	ginzap "github.com/gin-contrib/zap"
	middleware "github.com/oapi-codegen/gin-middleware"
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

	env := cfg.AssertEnv(
		setup.EnvKeyYdbEndpoint,
		setup.EnvKeyAuthTokenPublicKey,
	)

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, ydbpkg.YdbAuthMethodMetadata)
	db, err := ydb.Open(ctx, env[setup.EnvKeyYdbEndpoint], ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			logger.Error("close ydb", zap.Error(err))
		}
	}()

	store, err := store.NewBuilder().
		Logger(logger).
		Ydb(db).
		Build()
	if err != nil {
		logger.Fatal("new orders store", zap.Error(err))
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	gz := ginzap.Ginzap(logger, time.RFC3339, true)
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/ready" || c.Request.URL.Path == "/health" {
			c.Next()
		} else {
			gz(c)
		}
	})
	r.Use(ginzap.RecoveryWithZap(logger, true))

	readinessHandler := xgin.HandleReadiness(ctx)
	r.GET("/ready", readinessHandler)
	r.GET("/health", readinessHandler)

	svc, err := service.NewBuilder().
		Logger(logger).
		Store(store).
		Build()
	if err != nil {
		logger.Fatal("new cart service", zap.Error(err))
	}

	apiImpl := &presentation.ApiImpl{Logger: logger, Service: svc}

	bearerAuthenticator, err := auth.NewJwtBearerAuthenticator(env[setup.EnvKeyAuthTokenPublicKey])
	if err != nil {
		logger.Fatal("failed to setup jwt bearer authenticator", zap.Error(err))
	}

	swagger, err := oapi_codegen.GetSwagger()
	if err != nil {
		logger.Fatal("failed to setup swagger spec")
	}

	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		ErrorHandler: apiImpl.ErrorHandlerValidation,
		Options: openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}))

	authMiddleware, err := auth.NewBuilder().
		Authenticator(bearerAuthenticator).
		Routes(
			auth.NewRequiredRoute(
				oapi_codegen.OrdersGetOperationMethod,
				oapi_codegen.OrdersGetOperationPath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.OrdersGetOrderMethod,
				oapi_codegen.OrdersGetOrderPath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.OrdersUpdateOrderMethod,
				oapi_codegen.OrdersUpdateOrderPath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.OrdersListOrdersMethod,
				oapi_codegen.OrdersListOrdersPath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.OrdersCreateOrderMethod,
				oapi_codegen.OrdersCreateOrderPath,
			),
		).
		Build()
	if err != nil {
		logger.Fatal("build auth middleware", zap.Error(err))
	}

	oapi_codegen.RegisterHandlersWithOptions(r, apiImpl, oapi_codegen.GinServerOptions{
		ErrorHandler: apiImpl.ErrorHandler,
		Middlewares:  []oapi_codegen.MiddlewareFunc{authMiddleware},
	})

	r.NoRoute(xgin.HandleNotFound())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", Port),
		Handler: r.Handler(),
	}

	go func() {
		<-ctx.Done()

		logger.Info("got shutdown signal")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("failed to shut down http listener", zap.Error(err))
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("listen", zap.Error(err))
		}
	}
	logger.Info("shutdown server")
}
