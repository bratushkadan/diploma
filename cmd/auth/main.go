package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/frontend"
	"github.com/bratushkadan/floral/internal/auth/infrastructure/authn"
	"github.com/bratushkadan/floral/internal/auth/infrastructure/provider"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/postgres"
)

// "github.com/bratushkadan/floral/api/auth"

const webPort = 48612

// TODO: godot
func main() {
	// conf := auth.NewAuthServerConfig(webPort)
	// err := auth.RunServer(conf)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	userProv, rtPerProv, err := setupProviders()
	if err != nil {
		log.Fatal(err)
	}

	authSvc, err := setup(userProv, rtPerProv)
	if err != nil {
		log.Fatalf("failed to setup auth service: %v", err)
	}
	front := frontend.NewHttpImpl(authSvc, "id")
	if err := run(front, authSvc); err != nil {
		log.Fatal(err)
	}
}

func setupProviders() (*provider.PostgresUserProvider, *provider.PostgresRefreshTokenPersisterProvider, error) {
	conf, err := postgres.NewDBConf().
		WithDbHost("localhost").
		WithDbUser("root").
		WithDbPassword("root").
		WithDbPort(5432).
		WithDbName("auth").
		Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DBConf: %v", err)
	}
	db, err := provider.NewDbconnPool(conf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize db: %v", err)
	}

	hasher := auth.NewPasswordHasher("84778381-9207-4EC5-92A2-30F658D55872")
	userProv := provider.NewPostgresUserProvider(provider.PostgresUserProviderConf{
		Db:             db,
		DbConf:         conf,
		PasswordHasher: hasher,
		ConfirmationOpts: provider.ConfirmationIdsOpts{
			ExpiresAfter: 4 * time.Hour,
		},
	})
	rtPerProv := provider.NewPostgresRefreshTokenPersisterProvider(conf, db)

	return userProv, rtPerProv, nil
}

func setup(userProvider domain.UserProvider, rtPersisterProvider domain.RefreshTokenPersisterProvider) (*service.Auth, error) {
	appConfig := struct {
		JwtPrivateKeyPath    string
		JwtPublicKeyPath     string
		RefreshTokenDuration time.Duration
		AccessTokenDuration  time.Duration
		SecretToken          string
	}{
		JwtPrivateKeyPath:    os.Getenv("AUTH_JWT_PRIVATE_KEY_PATH"),
		JwtPublicKeyPath:     os.Getenv("AUTH_JWT_PUBLIC_KEY_PATH"),
		RefreshTokenDuration: 30 * 24 * 60 * time.Minute,
		AccessTokenDuration:  5 * time.Minute,
		SecretToken:          "foobar",
	}

	jwtProvider, err := authn.NewJwtProvider(authn.JwtProviderConf{
		PrivateKeyPath: appConfig.JwtPrivateKeyPath,
		PublicKeyPath:  appConfig.JwtPublicKeyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup jwt provider: %v", err)
	}

	yandexMailAppPassword := os.Getenv("YANDEX_MAIL_APP_PASSWORD")

	conf := &service.AuthConf{
		RefreshTokenProvider: authn.NewRefreshTokenJwtProvider(jwtProvider, appConfig.RefreshTokenDuration),
		AccessTokenProvider:  authn.NewAccessTokenJwtProvider(jwtProvider, appConfig.AccessTokenDuration),

		ConfirmationProvider: provider.NewYandexMailConfirmationProvider(provider.YandexMailConfirmationProviderConf{
			SenderMail: "danilabratushka@ya.ru",
			SenderPass: yandexMailAppPassword,
			ConfirmationOpts: provider.YandexMailConfirmationOpts{
				Endpoint:            "http://localhost:8080/api/v1/usersConfirmation/:confirmEmail",
				TokenQueryParameter: "id",
			},
		}),
		UserProvider:                  userProvider,
		RefreshTokenPersisterProvider: rtPersisterProvider,

		SecretToken: appConfig.SecretToken,
	}
	authSvc := service.NewAuth(conf)
	return authSvc, nil
}

func run(front frontend.FrontEnd, svc *service.Auth) error {
	return front.Start(svc)
}
