package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/internal/cart/service"
	shared_api "github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/bratushkadan/floral/pkg/xhttp/gin/middleware/auth"
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
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if !slices.Contains([]string{shared_api.SubjectTypeUser, shared_api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}
	if accessToken.SubjectType == shared_api.SubjectTypeUser && userId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	positions, err := api.CartService.GetCartPositions(c.Request.Context(), userId)
	if err != nil {
		api.Logger.Error("get cart positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartGetCartPositionsRes{Positions: positions})
}
func (api *ApiImpl) CartClearCart(c *gin.Context, userId string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if !slices.Contains([]string{shared_api.SubjectTypeUser, shared_api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}
	if accessToken.SubjectType == shared_api.SubjectTypeUser && userId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	if err := api.CartService.ClearCart(c.Request.Context(), userId); err != nil {
		api.Logger.Error("clear cart", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartClearCartRes{"message": "ok"})
}
func (api *ApiImpl) CartDeleteCartPosition(c *gin.Context, userId string, productId string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if !slices.Contains([]string{shared_api.SubjectTypeUser, shared_api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}
	if accessToken.SubjectType == shared_api.SubjectTypeUser && userId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	position, err := api.CartService.DeleteCartPosition(c.Request.Context(), userId, productId)
	if err != nil {
		api.Logger.Error("delete cart position", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	if position == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: fmt.Sprintf("position productId=%s not found", productId)}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartDeleteCartPositionRes{DeletedPosition: *position})
}
func (api *ApiImpl) CartSetCartPosition(c *gin.Context, userId string, productId string, params oapi_codegen.CartSetCartPositionParams) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if !slices.Contains([]string{shared_api.SubjectTypeUser, shared_api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}
	if accessToken.SubjectType == shared_api.SubjectTypeUser && userId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

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
		api.Logger.Error("publish carts positions", zap.Any("opertions", reqBody.Messages), zap.Error(err))
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
		api.Logger.Error("clear cart", zap.Any("users", reqBody.Messages), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "clear carts applied"})
}

func (api *ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	api.Logger.Info("validation handled", zap.String("validation_message", message))
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (api *ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	api.Logger.Error("error handler", zap.Error(err))
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
