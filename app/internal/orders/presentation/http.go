package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/service"
	shared_api "github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/bratushkadan/floral/pkg/xhttp/gin/middleware/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

type ApiImpl struct {
	Logger  *zap.Logger
	Service *service.Orders
}

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

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

	positions, err := api.Service.GetCartPositions(c.Request.Context(), userId)
	if err != nil {
		api.Logger.Error("get cart positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.CartGetCartPositionsRes{Positions: positions})
}

func (api *ApiImpl) PrivateCartsClearContents(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersCancelOperationsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

	if err := api.Service.ClearCarts(c.Request.Context(), reqBody.Messages); err != nil {
		api.Logger.Error("clear cart", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 1, Message: err.Error()}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "clear carts applied"})
}

func PrivateOrdersBatchCancelUnpaidOrders(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersBatchCancelUnpaidOrdersJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

}
func PrivateOrdersCancelOperations(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersCancelOperationsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

}
func PrivateOrdersProcessPublishedCartPositions(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessPublishedCartPositionsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}
}
func PrivateOrdersProcessReservedProducts(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

}
func PrivateOrdersProcessUnreservedProducts(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessUnreservedProductsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

}
func OrdersGetOperation(c *gin.Context, operationId string) {
	// oapi_codegen.OrdersGetOperationRes
}
func OrdersListOrders(c *gin.Context, params oapi_codegen.OrdersListOrdersParams) {
	// oapi_codegen.OrdersListOrdersRes{}

}
func OrdersCreateOrder(c *gin.Context) {
	// no request body
	// oapi_codegen.OrdersCreateOrderRes{}

}
func OrdersGetOrder(c *gin.Context, orderId string) {
	// oapi_codegen.OrdersGetOrderRes

}
func OrdersUpdateOrder(c *gin.Context, orderId string) {
	// oapi_codegen.OrdersUpdateOrderRes
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
