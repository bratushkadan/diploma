package store

import (
	"context"
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
	Term string
}
type SearchDTOOutput struct {
	NextPageToken *string
	Products      []SearchDTOOutputProduct
}
type SearchDTOOutputProduct struct {
	Id      string
	Name    string
	Picture string
	Price   float64
}

// Search items in catalog
func (s *Store) Search(ctx context.Context, in SearchDTOInput) (SearchDTOOutput, error) {
	content := strings.NewReader("")

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

	fmt.Println(string(data))

	var products []SearchDTOOutputProduct

	return SearchDTOOutput{
		Products: products,
	}, nil
}
