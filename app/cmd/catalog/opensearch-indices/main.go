package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/internal/catalog/constant"
	"github.com/bratushkadan/floral/internal/catalog/store"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/logging"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/zap"
)

var OpenSearchEndoints = cfg.EnvDefault(setup.EnvKeyOpenSearchEndpoints, "https://localhost:9200")

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	env := cfg.AssertEnv(
		setup.EnvKeyOpenSearchUser,
		setup.EnvKeyOpenSearchPassword,
	)

	logger, err := logging.NewZapConf("prod").Build()
	if err != nil {
		log.Fatalf("Error setting up zap: %v", err)
	}

	client, err := store.NewOpenSearchClientBuilder().
		Username(env[setup.EnvKeyOpenSearchUser]).
		Password(env[setup.EnvKeyOpenSearchPassword]).
		Addresses(strings.Split(OpenSearchEndoints, ",")).
		Build()
	if err != nil {
		logger.Fatal("failed to setup OpenSearch client: %v", zap.Error(err))
	}

	if err := createProductsIndex(ctx, client); err != nil {
		logger.Fatal("failed to create OpenSearch index", zap.String("index", constant.ProductIndex), zap.Error(err))
	}
	logger.Info("created OpenSearch index", zap.String("index", constant.ProductIndex))
}

func createProductsIndex(ctx context.Context, client *opensearch.Client) error {
	settings := strings.NewReader(`{
      "settings": {
        "index": {
          "number_of_shards": 1,
          "number_of_replicas": 0
        }
      },
      "mappings": {
        "properties": {
          "name": {
            "type": "text",
            "fields": {
              "keyword": {
                "type": "keyword"
              }
            }
          },
          "description": {
            "type": "text"
          },
          "price": {
            "type": "float"
          },
          "rating": {
            "type": "float"
          },
          "picture": {
            "type": "keyword",
            "index": false
          },
          "purchases_alltime": {
            "type": "long"
          },
          "purchases_30d": {
            "type": "long"
          },
          "ad_boost": {
            "type": "float"
          },
          "available": {
            "type": "boolean"
          },
          "stock_qty": {
            "type": "integer"
          }
        }
      }
    }`)

	req := opensearchapi.IndicesCreateRequest{
		Index: constant.ProductIndex,
		Body:  settings,
	}

	_, err := req.Do(ctx, client)
	if err != nil {
		return err
	}

	return nil
}
