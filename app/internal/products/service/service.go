package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/internal/products/store"
	"github.com/bratushkadan/floral/pkg/token"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Products struct {
	productsStore *store.Products
	picturesStore *store.Pictures

	l                             *zap.Logger
	encryptNextPageTokenSecretKey string
}

func New(products *store.Products, pictures *store.Pictures, logger *zap.Logger, encryptNextPageTokenSecretKey string) *Products {
	return &Products{
		productsStore:                 products,
		picturesStore:                 pictures,
		l:                             logger,
		encryptNextPageTokenSecretKey: encryptNextPageTokenSecretKey,
	}
}

type ListProductsReqFilter struct {
	SellerId *string
	InStock  *bool
	PageSize *int
}
type ListProductsReq struct {
	Filter        ListProductsReqFilter
	NextPageToken *string
}

type ListProductsNextPageSerialized struct {
	CreatedAt int64     `json:"created_at"`
	Id        uuid.UUID `json:"id"`
	InStock   *bool     `json:"in_stock"`
	SellerId  *string   `json:"seller_id"`
	PageSize  int       `json:"page_size"`
}

var (
	ErrInvalidListProductsPageSize      = errors.New("invalid list products page size")
	ErrInvalidListProductsNextPageToken = errors.New("invalid list products next page token")
)

func (s *Products) ListProducts(ctx context.Context, req ListProductsReq) (oapi_codegen.ListProductsRes, error) {
	if req.Filter.PageSize != nil {
		if *req.Filter.PageSize < 0 {
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("%w: page size can't be negative", ErrInvalidListProductsPageSize)
		}
		if *req.Filter.PageSize > 25 {
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("%w: max page size is 25", ErrInvalidListProductsPageSize)
		}
	}

	var page store.ListProductsNextPage
	if req.NextPageToken != nil {
		token, err := token.DecryptToken(*req.NextPageToken, s.encryptNextPageTokenSecretKey)
		if err != nil {
			s.l.Info("error decoding next page token", zap.Error(err))
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("%w: %w", ErrInvalidListProductsNextPageToken, err)
		}
		var deserializedPage ListProductsNextPageSerialized
		if err := json.Unmarshal([]byte(token), &deserializedPage); err != nil {
			s.l.Info("error unmarshaling next page token", zap.Error(err))
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("%w: %w", ErrInvalidListProductsNextPageToken, err)
		}
		page = store.ListProductsNextPage{
			CreatedAt: ptr(time.Unix(deserializedPage.CreatedAt, 0)),
			Id:        ptr(deserializedPage.Id),
			InStock:   deserializedPage.InStock,
			SellerId:  deserializedPage.SellerId,
			PageSize:  deserializedPage.PageSize,
		}
	} else {
		page = store.ListProductsNextPage{
			InStock:  req.Filter.InStock,
			SellerId: req.Filter.SellerId,
		}
		if req.Filter.PageSize != nil {
			page.PageSize = *req.Filter.PageSize
		} else {
			page.PageSize = 10
		}
	}

	s.l.Info("here with prepared page")
	items, err := s.productsStore.List(ctx, page)
	if err != nil {
		return oapi_codegen.ListProductsRes{}, fmt.Errorf("failed to list products from product store: %w", err)
	}

	if len(items) == 0 {
		return oapi_codegen.ListProductsRes{
			NextPageToken: nil,
			Products:      make([]oapi_codegen.ListProductsResProduct, 0),
		}, nil
	}

	products := make([]oapi_codegen.ListProductsResProduct, 0, len(items))
	var boundIndex int
	if lenItems := len(items); lenItems > page.PageSize {
		boundIndex = page.PageSize
	} else {
		boundIndex = lenItems
	}
	for _, item := range items[:boundIndex] {
		var pictureUrl string
		if len(item.Pictures) > 0 {
			pictureUrl = item.Pictures[0].Url
		}

		products = append(products, oapi_codegen.ListProductsResProduct{
			Id:         item.Id.String(),
			Name:       item.Name,
			SellerId:   item.SellerId,
			Price:      item.Price,
			PictureUrl: pictureUrl,
		})
	}

	res := oapi_codegen.ListProductsRes{
		NextPageToken: nil,
		Products:      products,
	}

	if len(items) > page.PageSize {
		nextPage := ListProductsNextPageSerialized{
			CreatedAt: items[page.PageSize].CreatedAt.Unix(),
			Id:        items[page.PageSize].Id,
			InStock:   page.InStock,
			SellerId:  page.SellerId,
			PageSize:  page.PageSize,
		}

		tokenBytes, err := json.Marshal(&nextPage)
		if err != nil {
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("failed to serialize next page token: %w", err)
		}

		token, err := token.EncryptToken(string(tokenBytes), s.encryptNextPageTokenSecretKey)
		if err != nil {
			return oapi_codegen.ListProductsRes{}, fmt.Errorf("failed to encrypt next page token: %w", err)
		}
		res.NextPageToken = &token
	}

	return res, nil
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
		Price:       product.Price,
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
		Price:       ptr(req.Price),
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
		Price:       product.Price,
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
	Price       *float64
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
		Price:       in.Price,
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
	if in.Price != nil {
		res.Price = &product.Price
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
