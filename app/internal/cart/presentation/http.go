package presentation

import (
	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/internal/cart/service"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

type ApiImpl struct {
	Logger      *zap.Logger
	CartService *service.Cart
}

func (api *ApiImpl) CartGetCartPositions(c *gin.Context, userId string) {

}
func (api *ApiImpl) CartClearCart(c *gin.Context, userId string) {

}
func (api *ApiImpl) CartDeleteCartPosition(c *gin.Context, userId string, productId string) {

}
func (api *ApiImpl) CartSetCartPosition(c *gin.Context, userId string, productId string, params oapi_codegen.CartSetCartPositionParams) {

}
func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
