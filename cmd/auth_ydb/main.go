package main

import (
	"context"
	"log"

	"github.com/bratushkadan/floral/pkg/cfg"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3"
)

var (
	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")
	authMethod      = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")
)

func main() {
	ctx := context.Background()
	db, err := ydb.Open(ctx, ydbFullEndpoint, ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Print()
		}
	}()
}
