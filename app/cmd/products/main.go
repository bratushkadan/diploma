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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/internal/products/presentation"
	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	xgin "github.com/bratushkadan/floral/pkg/xhttp/gin"
	"github.com/bratushkadan/floral/pkg/xhttp/gin/middleware/auth"
	ginzap "github.com/gin-contrib/zap"
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

	apiImpl := &presentation.ApiImpl{Logger: logger}

	jwtPublicKey := cfg.MustEnv("APP_AUTH_TOKEN_PUBLIC_KEY")
	bearerAuthenticator, err := auth.NewJwtBearerAuthenticator(jwtPublicKey)
	if err != nil {
		logger.Fatal("failed to setup jwt bearer authenticator", zap.Error(err))
	}

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
