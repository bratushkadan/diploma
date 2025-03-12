package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/internal/catalog/presentation"
	oapi_codegen "github.com/bratushkadan/floral/internal/catalog/presentation/generated"
	"github.com/bratushkadan/floral/internal/catalog/service"
	"github.com/bratushkadan/floral/internal/catalog/store"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	xgin "github.com/bratushkadan/floral/pkg/xhttp/gin"
	"github.com/getkin/kin-openapi/openapi3filter"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	middleware "github.com/oapi-codegen/gin-middleware"
	"go.uber.org/zap"
)

var (
	Port               = cfg.EnvDefault("PORT", "8080")
	OpenSearchEndoints = cfg.EnvDefault(setup.EnvKeyOpenSearchEndpoints, "https://localhost:9200")
)

func main() {
	env := cfg.AssertEnv(
		setup.EnvKeyOpenSearchUser,
		setup.EnvKeyOpenSearchPassword,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	client, err := store.NewOpenSearchClientBuilder().
		Username(env[setup.EnvKeyOpenSearchUser]).
		Password(env[setup.EnvKeyOpenSearchPassword]).
		Addresses(strings.Split(OpenSearchEndoints, ",")).
		Build()
	if err != nil {
		logger.Fatal("failed to setup OpenSearch client: %v", zap.Error(err))
	}

	store := store.NewStoreBuilder().
		OpenSearchClient(client).
		Logger(logger).
		Build()
	service := service.NewCatalog(store, logger)

	apiImpl := presentation.ApiImpl{Logger: logger, Service: service}

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

	r.POST("/api/internal/v1/sync-catalog", apiImpl.CatalogSync)

	swagger, err := oapi_codegen.GetSwagger()
	if err != nil {
		logger.Fatal("failed to setup swagger spec")
	}

	// TODO: determine why additionalProperties: false is not respected
	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		ErrorHandler: apiImpl.ErrorHandlerValidation,
		Options: openapi3filter.Options{
			// TODO: do some explorations
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}))

	oapi_codegen.RegisterHandlersWithOptions(r, apiImpl, oapi_codegen.GinServerOptions{
		ErrorHandler: apiImpl.ErrorHandler,
		Middlewares:  []oapi_codegen.MiddlewareFunc{},
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
