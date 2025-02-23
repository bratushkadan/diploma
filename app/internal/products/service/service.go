package service

import (
	"context"
	"errors"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/internal/products/store"
	"github.com/google/uuid"
)

type Products struct {
	productsStore *store.Products
	picturesStore *store.Pictures
}

func New(products *store.Products, pictures *store.Pictures) *Products {
	return &Products{
		productsStore: products,
		picturesStore: pictures,
	}
}

func (s *Products) ListProducts(ctx context.Context, req ListProductsReq) (oapi_codegen.ListProductsRes, error) {
	return oapi_codegen.ListProductsRes{}, errors.New("unimplemented")
}

func (s *Products) GetProduct(ctx context.Context, id uuid.UUID) (*oapi_codegen.GetProductRes, error) {
	product, err := s.productsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	pictures := make([]oapi_codegen.GetProductResPicture, 0, len(product.Pictures))

	for _, p := range product.Pictures {
		pictures = append(pictures, oapi_codegen.GetProductResPicture{
			Id:  p.Id,
			Url: p.Url,
		})
	}

	return &oapi_codegen.GetProductRes{
		Id:          product.Id.String(),
		SellerId:    product.SellerId,
		Name:        product.Name,
		Description: product.Description,
		Pictures:    pictures,
		Metadata:    product.Metadata,
		Stock:       int(product.Stock),
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Products) CreateProduct(ctx context.Context, req *oapi_codegen.CreateProductReq, sellerId string) (oapi_codegen.CreateProductRes, error) {
	product, err := s.productsStore.Upsert(ctx, store.UpsertProductDTOInput{
		Id:          ptr(uuid.New()),
		SellerId:    ptr(sellerId),
		Name:        ptr(req.Name),
		Description: ptr(req.Description),
		Pictures:    []store.UpsertProductDTOOutputPicture{},
		Metadata:    map[string]any{},
		Stock:       ptr(uint32(req.Stock)),
		CreatedAt:   ptr(time.Now()),
		UpdatedAt:   ptr(time.Now()),
	})
	if err != nil {
		return oapi_codegen.CreateProductRes{}, err
	}

	pictures := make([]oapi_codegen.GetProductResPicture, 0, len(product.Pictures))

	for _, p := range product.Pictures {
		pictures = append(pictures, oapi_codegen.GetProductResPicture{
			Id:  p.Id,
			Url: p.Url,
		})
	}

	return oapi_codegen.CreateProductRes{
		Id:          product.Id.String(),
		SellerId:    product.SellerId,
		Name:        product.Name,
		Description: product.Description,
		Pictures:    pictures,
		Metadata:    product.Metadata,
		Stock:       int(product.Stock),
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}, nil
}

type ListProductsReqFilter struct {
	SellerId   string
	ProductIds *[]string
	Name       *string
	InStock    *bool
}
type ListProductsReq struct {
	Filter        ListProductsReqFilter
	NextPageToken *string
}

func ptr[T any](v T) *T {
	return &v
}
