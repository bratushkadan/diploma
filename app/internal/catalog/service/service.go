package service

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/catalog/presentation/generated"
	"github.com/bratushkadan/floral/internal/catalog/store"
	"go.uber.org/zap"
)

type Catalog struct {
	logger *zap.Logger
	store  *store.Store
}

func NewCatalog(store *store.Store, logger *zap.Logger) *Catalog {
	return &Catalog{
		logger: logger,
		store:  store,
	}
}

type SearchReq struct {
	Term          string
	NextPageToken *string
}

func (c *Catalog) Search(ctx context.Context, req SearchReq) (oapi_codegen.CatalogGetRes, error) {
	out, err := c.store.Search(ctx, store.SearchDTOInput{
		NextPageToken: req.NextPageToken,
		Term:          req.Term,
	})
}
