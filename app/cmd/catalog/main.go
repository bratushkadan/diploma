package main

import (
	"context"
	"io"
	"log"
	"os/signal"
	"syscall"
	"time"

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

func main() {
	productsCdcTopic := "products-cdc-target"
	productsCdcTopicConsumer := "catalog"

	env := cfg.AssertEnv()
	_ = env
	// env := cfg.AssertEnv(
	// 	setup.EnvKeyYdbEndpoint,
	// )

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	client, err := store.NewOpenSearchClientBuilder().
		Username("admin").
		Password("iAdnWfymi1(").
		Addresses([]string{"https://localhost:9200"}).
		Build()
	if err != nil {
		logger.Fatal("failed to setup OpenSearch client: %v", zap.Error(err))
	}

	store := store.NewStoreBuilder().
		OpenSearchClient(client).
		Build()
	service := service.NewCatalog(store, logger)

	impl := presentation.ApiImpl{Logger: logger, Service: service}

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

	swagger, err := oapi_codegen.GetSwagger()
	if err != nil {
		logger.Fatal("failed to setup swagger spec")
	}

	// TODO: determine why additionalProperties: false is not respected
	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		ErrorHandler: impl.ErrorHandlerValidation,
		Options: openapi3filter.Options{
			// TODO: do some explorations
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}))

}
