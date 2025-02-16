package presentation

import (
	"net/http"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

type ApiImpl struct {
}

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

func (*ApiImpl) ProductsList(c *gin.Context, params oapi_codegen.ProductsListParams) {
	c.JSON(http.StatusOK, oapi_codegen.ListProductsRes{
		NextPageToken: "",
		Products: []oapi_codegen.ListProductsResProduct{
			{
				Id:         "1",
				Name:       "foo",
				PictureUrl: "https://www.ferra.ru/imgs/2024/05/08/05/6460496/c2150453d059e8999c5f0b211ce334f7c869147c.jpg",
				SellerId:   "3",
			},
		},
	})
}

func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
