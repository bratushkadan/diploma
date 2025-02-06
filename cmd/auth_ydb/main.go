package main

import (
	"context"
	"fmt"
	"log"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
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

	accountAdapter := ydb_adapter.NewYDBAccountAdapter(ydb_adapter.YDBAccountAdapterConf{
		DbDriver: db,
	})

	output, err := accountAdapter.CreateAccount(ctx, domain.CreateAccountDTOInput{
		Name:     "Danila",
		Email:    "someemail",
		Password: "ooga",
		Type:     "user",
	})
	if err != nil {
		log.Print(err)
	} else {
		fmt.Printf("%+v\n", output)
	}
}
