package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/zap"
)

const ProductsIndex = "products"

type Store struct {
	logger     *zap.Logger
	opensearch *opensearch.Client
}

type StoreBuilder struct {
	store Store
}

func NewStoreBuilder() *StoreBuilder {
	return &StoreBuilder{}
}

func (b *StoreBuilder) Logger(logger *zap.Logger) *StoreBuilder {
	b.store.logger = logger
	return b
}
func (b *StoreBuilder) OpenSearchClient(client *opensearch.Client) *StoreBuilder {
	b.store.opensearch = client
	return b
}

func (b *StoreBuilder) Build() *Store {
	if b.store.logger == nil {
		b.store.logger = zap.NewNop()
	}
	return &b.store
}

// List catalog items.
func (s *Store) List(ctx context.Context) error { return nil }

type SearchDTOInput struct {
	NextPageToken *string
	// Search term.
	Term *string
}
type SearchDTOOutput struct {
	NextPageToken *string
	Products      []SearchDTOOutputProduct
}
type SearchDTOOutputProduct struct {
	Id      string
	Name    string
	Picture *string
	Price   float64
}

type ProductsOpenSearchResp struct {
	Hits struct {
		Hits []struct {
			Id     string `json:"_id"`
			Source struct {
				Name string `json:"name"`
				// Description string  `json:"description"`
				Price   float64 `json:"price"`
				Picture *string `json:"picture"`
				// Stock       int     `json:"stock"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (s *Store) Search(ctx context.Context, in SearchDTOInput) (SearchDTOOutput, error) {
	var content io.Reader
	if in.Term != nil {
		content = strings.NewReader(newCatalogSearchQuery(20, 0, *in.Term))
	} else {
		content = strings.NewReader(newCatalogQuery(20, 0))
	}

	req := opensearchapi.SearchRequest{
		Index: []string{ProductsIndex},
		Body:  content,
	}
	searchResp, err := req.Do(ctx, s.opensearch)
	if err != nil {
		return SearchDTOOutput{}, fmt.Errorf("failed to search documents: %w", err)
	}
	data, err := io.ReadAll(searchResp.Body)
	if err != nil {
		return SearchDTOOutput{}, fmt.Errorf("failed read OpenSearch search documents response: %v", err)
	}
	if searchResp.StatusCode > 399 {
		s.logger.Error("failed to perform search operation in OpenSearch", zap.Int("status", searchResp.StatusCode), zap.ByteString("response_body", data))
		return SearchDTOOutput{}, err
	}

	var hits ProductsOpenSearchResp
	if err := json.Unmarshal(data, &hits); err != nil {
		return SearchDTOOutput{}, fmt.Errorf("failed to unmarshal OpenSearch search products response: %v", err)
	}

	products := make([]SearchDTOOutputProduct, 0, len(hits.Hits.Hits))

	for _, hit := range hits.Hits.Hits {
		products = append(products, SearchDTOOutputProduct{
			Id:      hit.Id,
			Name:    hit.Source.Name,
			Price:   hit.Source.Price,
			Picture: hit.Source.Picture,
		})
	}

	return SearchDTOOutput{
		Products: products,
	}, nil
}

func newCatalogQuery(limit, offset int) string {
	// GET /products/_search
	return fmt.Sprintf(`
{
  "size": %d,
  "from": %d,
  "query": {
    "function_score": {
      "query": { "match_all": {} },
      "functions": [
        {
          "field_value_factor": {
            "field": "rating",
            "factor": 1.5,
            "modifier": "sqrt",
            "missing": 3.5
          }
        },
        {
          "field_value_factor": {
            "field": "purchases_30d",
            "factor": 0.1,
            "modifier": "log1p",
            "missing": 0
          }
        },
        {
          "field_value_factor": {
            "field": "ad_boost",
            "factor": 2.0,
            "modifier": "none",
            "missing": 1
          }
        }
      ],
      "score_mode": "sum",
      "boost_mode": "multiply"
    }
  }
}
`, limit, offset)

}

// name^3 - priority of field "name" is increased 3 times to "description field"
func newCatalogSearchQuery(limit, offset int, query string) string {
	// GET /products/_search
	return fmt.Sprintf(`
{
  "size": %d,
  "from": %d,
  "query": {
    "function_score": {
    "query": {
      "multi_match": {
        "query": "%s",
        "fields": ["name^3", "description"]
      }
    },
      "functions": [
        {
          "field_value_factor": {
            "field": "rating",
            "factor": 1.5,
            "modifier": "sqrt",
            "missing": 3.5
          }
        },
        {
          "field_value_factor": {
            "field": "purchases_30d",
            "factor": 0.1,
            "modifier": "log1p",
            "missing": 0
          }
        },
        {
          "field_value_factor": {
            "field": "ad_boost",
            "factor": 2.0,
            "modifier": "none",
            "missing": 1
          }
        }
      ],
      "score_mode": "sum",
      "boost_mode": "multiply"
    }
  }
}
`, limit, offset, query)

}
