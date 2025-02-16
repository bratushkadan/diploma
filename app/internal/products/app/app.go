package app

import oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"

type App struct {
}

type AppIf interface {
	ListProducts(req ListProductsReq) (oapi_codegen.ListProductsRes, error)
	GetProduct()
	CreateProduct()
	UpdateProduct()
	DeleteProduct()
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
