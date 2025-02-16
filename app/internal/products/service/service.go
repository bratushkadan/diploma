package service

import (
	"context"
	"errors"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
)

type Service struct {
	repo any
}

type ServiceIf interface {
	CreateProduct()
	UpdateProduct()
	DeleteProduct()
}

func (s *Service) ListProducts(ctx context.Context, req ListProductsReq) (oapi_codegen.ListProductsRes, error) {
	return oapi_codegen.ListProductsRes{}, errors.New("unimplemented")
}

func (s *Service) GetProduct(ctx context.Context, id string) (oapi_codegen.GetProductRes, error) {
	return oapi_codegen.GetProductRes{}, errors.New("unimplemented")
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

func New() {

}
