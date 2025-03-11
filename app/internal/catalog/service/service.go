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

func (c *Catalog) Get(ctx context.Context, nextPageToken *string) (oapi_codegen.CatalogGetRes, error) {
	out, err := c.store.Search(ctx, store.SearchDTOInput{
		NextPageToken: nextPageToken,
	})
	if err != nil {
		return oapi_codegen.CatalogGetRes{}, err
	}

	res := oapi_codegen.CatalogGetRes{
		Products: make([]oapi_codegen.CatalogGetResProduct, 0, len(out.Products)),
	}

	for _, p := range out.Products {
		res.Products = append(res.Products, oapi_codegen.CatalogGetResProduct{
			Id:      p.Id,
			Name:    p.Name,
			Picture: p.Picture,
			Price:   p.Price,
		})
	}

	res.NextPageToken = out.NextPageToken

	return res, nil
}

type SearchReq struct {
	Term          *string
	NextPageToken *string
}

func (c *Catalog) Search(ctx context.Context, req SearchReq) (oapi_codegen.CatalogGetRes, error) {
	out, err := c.store.Search(ctx, store.SearchDTOInput{
		NextPageToken: req.NextPageToken,
		Term:          req.Term,
	})
	if err != nil {
		return oapi_codegen.CatalogGetRes{}, err
	}

	res := oapi_codegen.CatalogGetRes{
		Products:      make([]oapi_codegen.CatalogGetResProduct, 0, len(out.Products)),
		NextPageToken: out.NextPageToken,
	}

	for _, p := range out.Products {
		res.Products = append(res.Products, oapi_codegen.CatalogGetResProduct{
			Id:      p.Id,
			Name:    p.Name,
			Picture: p.Picture,
			Price:   p.Price,
		})
	}

	return res, nil
}
