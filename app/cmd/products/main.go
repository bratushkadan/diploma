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
	"github.com/bratushkadan/floral/internal/products/presentation"
	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/internal/products/service"
	"github.com/bratushkadan/floral/internal/products/store"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/bratushkadan/floral/pkg/s3aws"
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
		setup.EnvKeyAwsAccessKeyId,
		setup.EnvKeyAwsSecretAccessKey,
		setup.EnvKeyAuthTokenPublicKey,
		setup.EnvKeyStorePicturesBucket,
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
			logger.Error("failed to close ydb", zap.Error(err))
		}
	}()

	productsStore, err := store.NewProductsBuilder().
		Logger(logger).
		YDBDriver(db).
		Build()
	if err != nil {
		logger.Fatal("failed to setup products store", zap.Error(err))
	}

	s3client, err := s3aws.New(ctx, env[setup.EnvKeyAwsAccessKeyId], env[setup.EnvKeyAwsSecretAccessKey])
	if err != nil {
		logger.Fatal("failed to setup s3 client", zap.Error(err))
	}
	pictureStore, err := store.NewPicturesBuilder().
		Bucket(env[setup.EnvKeyStorePicturesBucket]).
		S3Client(s3client).
		Build()
	if err != nil {
		logger.Fatal("failed to setup picture store", zap.Error(err))
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

	svc := service.New(productsStore, pictureStore, logger)

	apiImpl := &presentation.ApiImpl{Logger: logger, ProductsService: svc, PictureStore: pictureStore}

	bearerAuthenticator, err := auth.NewJwtBearerAuthenticator(env[setup.EnvKeyAuthTokenPublicKey])
	if err != nil {
		logger.Fatal("failed to setup jwt bearer authenticator", zap.Error(err))
	}

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

	authMiddleware, err := auth.NewBuilder().
		Authenticator(bearerAuthenticator).
		Routes(
			auth.NewRequiredRoute(
				oapi_codegen.ProductsCreateMethod,
				oapi_codegen.ProductsCreatePath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.ProductsUpdateMethod,
				oapi_codegen.ProductsUpdatePath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.ProductsDeleteMethod,
				oapi_codegen.ProductsDeletePath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.ProductsUploadPictureMethod,
				oapi_codegen.ProductsUploadPicturePath,
			),
			auth.NewRequiredRoute(
				oapi_codegen.ProductsDeletePictureMethod,
				oapi_codegen.ProductsDeletePicturePath,
			),
		).
		Build()
	if err != nil {
		logger.Fatal("failed to setup auth middleware", zap.Error(err))
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
