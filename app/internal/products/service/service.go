package service

import (
	"context"
	"errors"
	"fmt"
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
		Id:          uuid.New(),
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

type UpdateProductReq struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	Metadata    map[string]any
	Pictures    []store.UpsertProductDTOOutputPicture
	StockDelta  *int32
}

var (
	ErrInsufficientStock = errors.New("stock is not sufficient")
)

func (s *Products) UpdateProduct(ctx context.Context, in UpdateProductReq) (oapi_codegen.UpdateProductRes, error) {
	var stock *uint32

	if in.StockDelta != nil {
		product, err := s.productsStore.Get(ctx, in.Id)
		if err != nil {
			return oapi_codegen.UpdateProductRes{}, fmt.Errorf("failed to retrieve product for calculating in stock value: %w", err)
		}
		if int32(product.Stock)+*in.StockDelta < 0 {
			return oapi_codegen.UpdateProductRes{}, fmt.Errorf("%w: trying to withdraw %d units from stock when there's only %d", ErrInsufficientStock, -1*(*in.StockDelta), product.Stock)
		}
		updatedStock := uint32(int32(product.Stock) + *in.StockDelta)
		stock = &updatedStock
	}

	product, err := s.productsStore.Upsert(ctx, store.UpsertProductDTOInput{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		Metadata:    in.Metadata,
		Pictures:    in.Pictures,
		Stock:       stock,
		UpdatedAt:   ptr(time.Now()),
	})
	if err != nil {
		return oapi_codegen.UpdateProductRes{}, err
	}

	var res oapi_codegen.UpdateProductRes

	if in.Name != nil {
		res.Name = &product.Name
	}
	if in.Description != nil {
		res.Description = &product.Description
	}
	if in.Metadata != nil {
		res.Metadata = &product.Metadata
	}
	if in.StockDelta != nil {
		vstock := int(*stock)
		res.Stock = &vstock
	}

	return res, nil
}

var (
	ErrProductNotFound = errors.New("product not found")
)

func (s *Products) DeleteProduct(ctx context.Context, id uuid.UUID) (oapi_codegen.DeleteProductRes, error) {
	out, err := s.productsStore.Delete(ctx, store.DeleteProductDTOInput{Id: id, DeletedAt: time.Now()})
	if err != nil {
		return oapi_codegen.DeleteProductRes{}, err
	}
	if out == nil {
		return oapi_codegen.DeleteProductRes{}, fmt.Errorf(`failed to delete product id "%s": %w`, id.String(), ErrProductNotFound)
	}

	return oapi_codegen.DeleteProductRes{Id: out.Id.String()}, nil
}

func ptr[T any](v T) *T {
	return &v
}
