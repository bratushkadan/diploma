package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/internal/catalog/store"
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
	CdcOperationUpsert CdcOperation = "u"
	CdcOperationDelete CdcOperation = "d"
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
	Id                  string   `json:"id"`
	Name                *string  `json:"name"`
	Description         *string  `json:"description"`
	PicturesJsonListStr *string  `json:"pictures"`
	Price               *float64 `json:"price"`
	Stock               *uint32  `json:"stock"`
	CreatedAtUnixMs     *int64   `json:"created_at"`
	UpdatedAtUnixMs     *int64   `json:"updated_at"`
	DeletedAtUnixMs     *int64   `json:"deleted_at"`
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

	client, err := store.NewOpenSearchClientBuilder().
		Username("admin").
		Password("iAdnWfymi1(").
		Addresses([]string{"https://localhost:9200"}).
		Build()
	if err != nil {
		logger.Fatal("failed to setup OpenSearch client: %v", zap.Error(err))
	}

	newBulkProductUpsert := func(p ProductChange) (string, error) {
		update := map[string]map[string]string{
			"update": {
				"_index": store.ProductsIndex,
				"_id":    p.Id,
			},
		}
		opData, err := json.Marshal(update)
		if err != nil {
			return "", err
		}
		doc := make(map[string]any)
		docVal := map[string]any{
			"doc":           doc,
			"doc_as_upsert": true,
		}
		doc["name"] = p.Name
		doc["description"] = p.Description
		doc["price"] = p.Price
		doc["stock"] = p.Stock
		if len(p.Pictures) > 0 {
			doc["picture"] = p.Pictures[0].Url
		} else {
			doc["picture"] = nil
		}
		docData, err := json.Marshal(docVal)
		if err != nil {
			return "", err
		}
		return string(opData) + "\n" + string(docData), nil
	}
	newBulkProductDelete := func(p ProductChange) (string, error) {
		update := map[string]map[string]string{
			"delete": {
				"_index": store.ProductsIndex,
				"_id":    p.Id,
			},
		}
		opData, err := json.Marshal(update)
		if err != nil {
			return "", err
		}
		return string(opData), nil
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
		var blkBuf bytes.Buffer
		for _, msg := range data {
			var record ProductChangeCdcMessage
			if err := json.Unmarshal(msg, &record); err != nil {
				log.Fatal("failed to unmarshal product message", zap.Error(err))
			}

			switch record.Payload.Operation {
			case CdcOperationUpsert:
				var pictures []ProductsChangePicture
				if err := json.Unmarshal([]byte(*record.Payload.After.PicturesJsonListStr), &pictures); err != nil {
					return fmt.Errorf("failed to unmarshal CDC pictures string with json list to json: %v", err)
				}
				data, err := base64.StdEncoding.DecodeString(record.Payload.After.Id)
				if err != nil {
					return fmt.Errorf(`failed to decode base64 encoded bytes field "id": %v`, err)
				}
				uuidId := string(data)

				isDeleted := record.Payload.After.DeletedAtUnixMs != nil
				isOutOfStock := *record.Payload.After.Stock == 0
				if isDeleted || isOutOfStock {
					bulkItem, err := newBulkProductDelete(ProductChange{Id: uuidId})
					if err != nil {
						logger.Fatal("failed to prepare bulk delete item", zap.Error(err))
					}
					blkBuf.WriteString(bulkItem)
				} else {
					bulkItem, err := newBulkProductUpsert(ProductChange{
						Id:          uuidId,
						Name:        *record.Payload.After.Name,
						Description: *record.Payload.After.Description,
						Price:       *record.Payload.After.Price,
						Stock:       *record.Payload.After.Stock,
						Pictures:    pictures,
					})
					if err != nil {
						logger.Fatal("failed to prepare bulk item", zap.Error(err))
					}
					blkBuf.WriteString(bulkItem)
				}
			case CdcOperationDelete:
				data, err := base64.StdEncoding.DecodeString(record.Payload.Before.Id)
				if err != nil {
					return fmt.Errorf(`failed to decode base64 encoded bytes field "id": %v`, err)
				}
				uuidId := string(data)
				bulkItemDel, err := newBulkProductDelete(ProductChange{Id: uuidId})
				if err != nil {
					logger.Fatal("failed to prepare bulk delete item", zap.Error(err))
				}
				blkBuf.WriteString(bulkItemDel)
			default:
				return fmt.Errorf("unkown CDC operation type %s", record.Payload.Operation)
			}
			blkBuf.WriteByte('\n')
		}

		blk, err := client.Bulk(&blkBuf)
		if err != nil {
			return fmt.Errorf("failed to run bulk request products: %v", err)
		}
		if blk.StatusCode > 399 {
			data, err := io.ReadAll(blk.Body)
			if err != nil {
				return fmt.Errorf("failed read OpenSearch bulk response: %v", err)
			}
			logger.Error("failed to perform bulk operation in OpenSearch", zap.Int("status", blk.StatusCode), zap.ByteString("response_body", data))
			return fmt.Errorf("failed to perform bulk operation in OpenSearch: %v", err)
		}

		return nil
	}); err != nil {
		logger.Fatal("failed to consume a message batch from topic", zap.String("topic", topic), zap.String("consumer", consumer))
	}
}
