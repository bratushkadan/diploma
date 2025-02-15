package ydbpkg

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	environ "github.com/ydb-platform/ydb-go-sdk-auth-environ"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	yc "github.com/ydb-platform/ydb-go-yc"
)

const (
	YdbAuthMethodMetadata = "metadata"
	YdbAuthMethodEnviron  = "environ"
)

func GetYdbAuthOpts(ydbAuthMethod string) []ydb.Option {
	//  yc.WithInternalCA(), // используем сертификаты Яндекс Облака
	//  ydb.WithAccessTokenCredentials(iamToken), // аутентификация с помощью токена
	//  ydb.WithAnonymousCredentials(), // анонимная аутентификация (например, в docker ydb)
	//  yc.WithMetadataCredentials(token), // аутентификация изнутри виртуальной машины в Яндекс Облаке или из Яндекс Функции
	//  yc.WithServiceAccountKeyFileCredentials("~/.ydb/sa.json"), // аутентификация в Яндекс Облаке с помощью файла сервисного аккаунта
	var opts []ydb.Option

	switch ydbAuthMethod {
	case YdbAuthMethodEnviron:
		opts = append(opts, environ.WithEnvironCredentials())
	case YdbAuthMethodMetadata:
		fallthrough
	default:
		opts = append(opts, yc.WithMetadataCredentials())
	}

	return opts
}

func IsUniqueConstraintViolation(err error) bool {
	return ydb.IsOperationError(err, Ydb.StatusIds_PRECONDITION_FAILED)
}
