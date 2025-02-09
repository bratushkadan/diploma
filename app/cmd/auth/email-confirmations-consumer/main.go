package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/service"
	"github.com/bratushkadan/floral/pkg/cfg"
)

var (
	ydbFullEndpoint string
	authMethod      string
)

const (
	EnvKeyAccountIdHashSalt       = "APP_ID_ACCOUNT_HASH_SALT"
	EnvKeyTokenIdHashSalt         = "APP_ID_TOKEN_HASH_SALT"
	EnvKeyPasswordHashSalt        = "APP_PASSWORD_HASH_SALT"
	EnvKeyAuthTokenPrivateKeyPath = "APP_AUTH_TOKEN_PRIVATE_KEY_PATH"
	EnvKeyAuthTokenPublicKeyPath  = "APP_AUTH_TOKEN_PUBLIC_KEY_PATH"
)

type DummyAccountCreationNotificationProvider struct {
}

func (p DummyAccountCreationNotificationProvider) Send(_ context.Context, _ domain.SendAccountCreationNotificationDTOInput) (domain.SendAccountCreationNotificationDTOOutput, error) {
	return domain.SendAccountCreationNotificationDTOOutput{}, nil
}

var _ domain.AccountCreationNotificationProvider = (*DummyAccountCreationNotificationProvider)(nil)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}
}

func init() {
	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")
	authMethod = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")
}

func main() {

	svc := service.NewAuthBuilder()
	_ = svc
}
