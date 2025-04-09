package presentation

import (
	"encoding/json"
	"net/http"

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
	positions, err := api.CartService.GetCartPositions(c.Request.Context(), userId)
	if err != nil {
		api.Logger.Error("get cart positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartGetCartPositionsRes{Positions: positions})
}
func (api *ApiImpl) CartClearCart(c *gin.Context, userId string) {
	if err := api.CartService.ClearCart(c.Request.Context(), userId); err != nil {
		api.Logger.Error("clear cart", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartClearCartRes{"message": "ok"})
}
func (api *ApiImpl) CartDeleteCartPosition(c *gin.Context, userId string, productId string) {
	position, err := api.CartService.DeleteCartPosition(c.Request.Context(), userId, productId)
	if err != nil {
		api.Logger.Error("delete cart position", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartDeleteCartPositionRes{DeletedPosition: position})
}
func (api *ApiImpl) CartSetCartPosition(c *gin.Context, userId string, productId string, params oapi_codegen.CartSetCartPositionParams) {
	position, err := api.CartService.SetCartPosition(c.Request.Context(), userId, productId, params.Count)
	if err != nil {
		api.Logger.Error("publish carts positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartSetCartPositionRes{SetPosition: position})
}

func (api *ApiImpl) PrivateCartPublishContents(c *gin.Context) {
	var reqBody oapi_codegen.PrivateCartPublishContentsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	if err := api.CartService.CartsPublishPositions(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("publish carts positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cart contents published"})
}
func (api *ApiImpl) PrivateCartsClearContents(c *gin.Context) {
	var reqBody oapi_codegen.PrivateCartsClearContentsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	if err := api.CartService.ClearCarts(c.Request.Context(), reqBody.Messages); err != nil {
		api.Logger.Error("clear cart", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "clear carts applied"})
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
