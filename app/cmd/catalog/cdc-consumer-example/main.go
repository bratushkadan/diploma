package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/zap"
)

type CdcOperation string

var (
	// There's no way to create a distinction between creating and updating, in both cases "before" is nil.
	// It's true both for UPDATE and UPSERT clauses.
	CdcOperationUpsert = "u"
	CdcOperationDelete = "d"
)

type ProductChangeCdcMessage struct {
	Payload ProductChangeCdcMessagePayload `json:"payload"`
}

type ProductChangeCdcMessagePayload struct {
	Before    *ProductChangeSchema `json:"before"`
	After     *ProductChangeSchema `json:"after"`
	Operation CdcOperation         `json:"op"`
}

type ProductChangeSchema struct {
	Id                  string  `json:"id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	PicturesJsonListStr string  `json:"pictures"`
	Price               float64 `json:"price"`
	Stock               uint32  `json:"stock"`
	CreatedAtUnixMs     int64   `json:"created_at"`
	UpdatedAtUnixMs     int64   `json:"updated_at"`
	DeletedAtUnixMs     *int64  `json:"deleted_at"`
}

type ProductChange struct {
	Id              string
	Name            string
	Description     string
	Pictures        []ProductsChangePicture
	Price           float64
	Stock           uint32
	CreatedAtUnixMs int64
	UpdatedAtUnixMs int64
	DeletedAtUnixMs *int64
}
type ProductsChangePicture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

func main() {
	topic := "products-cdc-target"
	consumer := "catalog"

	env := cfg.AssertEnv(
		setup.EnvKeyYdbEndpoint,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	authMethod := cfg.EnvDefault(setup.EnvKeyYdbAuthMethod, ydbpkg.YdbAuthMethodMetadata)
	db, err := ydb.Open(ctx, env[setup.EnvKeyYdbEndpoint], ydbpkg.GetYdbAuthOpts(authMethod)...)
	if err != nil {
		logger.Fatal("failed to setup ydb", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			logger.Error("failed to close ydb", zap.Error(err))
		}
	}()

	reader, err := ydbtopic.NewConsumer(db, topic, consumer)
	if err != nil {
		logger.Fatal("failed to setup ydb topic consumer", zap.String("topic", topic), zap.String("consumer", consumer))
	}

	if err := ydbtopic.ConsumeBatch(ctx, reader, func(data [][]byte) error {
		for _, msg := range data {
			var record ProductChangeCdcMessage
			if err := json.Unmarshal(msg, &record); err != nil {
				log.Fatal("failed to unmarshal product message", zap.Error(err))
			}
			log.Printf("consumed a message: %+v, After: %+v", record, record.Payload.After)
		}

		return errors.New("do not commit")
	}); err != nil {
		logger.Fatal("failed to consume a message batch from topic", zap.String("topic", topic), zap.String("consumer", consumer))
	}
}
