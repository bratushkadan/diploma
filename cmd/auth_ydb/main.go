package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ydb_adapter "github.com/bratushkadan/floral/internal/auth/adapters/secondary/ydb"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3"
)

var (
	ydbFullEndpoint = cfg.MustEnv("YDB_ENDPOINT")
	authMethod      = cfg.EnvDefault("YDB_AUTH_METHOD", "metadata")
)

func main() {
	idHasher, err := idhash.New("some-very-secret-salt-phrase", idhash.WithPrefix("ie"))
	if err != nil {
		log.Fatal(err)
	}

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
		IdHasher: idHasher,
	})

	output, err := accountAdapter.CreateAccount(ctx, domain.CreateAccountDTOInput{
		Name:     "Danila",
		Email:    fmt.Sprintf(`someemail-%d@gmail.com`, time.Now().UnixMilli()),
		Password: "ooga",
		Type:     "user",
	})
	if err != nil {
		log.Print(err)
	} else {
		fmt.Printf("%+v\n", output)

		idInt64, err := idHasher.DecodeInt64(output.Id)
		if err != nil {
			log.Fatal(fmt.Errorf(`failed to decode str id "%s" to int64: %v`, output.Id, err))
		}
		log.Printf(`decoded string id "%s" to int64 id "%d"`, output.Id, idInt64)
	}
}
