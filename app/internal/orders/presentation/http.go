package presentation

import (
	"encoding/json"
	"errors"
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

func (api *ApiImpl) PrivateOrdersBatchCancelUnpaidOrders(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersBatchCancelUnpaidOrdersJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}
	if err := api.Service.BatchCancelUnpaidOrders(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("batch cancel unpaid orders", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: "failed to batch cancel unpaid orders",
		}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (api *ApiImpl) PrivateOrdersCancelOperations(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersCancelOperationsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

	if err := api.Service.CancelOperations(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("process published cart positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: "failed to process published cart positions",
		}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (api *ApiImpl) PrivateOrdersProcessPublishedCartPositions(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessPublishedCartPositionsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

	if err := api.Service.ProcessPublishedCartPositions(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("process published cart positions", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: "failed to process published cart positions",
		}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (api *ApiImpl) PrivateOrdersProcessReservedProducts(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

	if err := api.Service.ProcessReservedProducts(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("process reserved products", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: "failed to process reserved products",
		}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (api *ApiImpl) PrivateOrdersProcessUnreservedProducts(c *gin.Context) {
	var reqBody oapi_codegen.PrivateOrdersProcessUnreservedProductsJSONRequestBody
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: fmt.Sprintf("invalid request body: %v", err),
		}))
		return
	}

	if err := api.Service.ProcessUnreservedProducts(c.Request.Context(), reqBody); err != nil {
		api.Logger.Error("process unreserved products", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{
			Code:    1,
			Message: "failed to process unreserved products",
		}))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (api *ApiImpl) OrdersGetOperation(c *gin.Context, operationId string) {
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

	op, err := api.Service.GetOperation(c.Request.Context(), operationId)
	if err != nil {
		api.Logger.Info("retrieve operation", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "failed to retrieve operation"}},
		})
		return
	}
	if op == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "operation not found"}},
		})
		return
	}

	if accessToken.SubjectType == shared_api.SubjectTypeUser && op.UserId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	c.JSON(http.StatusOK, op)
}

func (api *ApiImpl) OrdersListOrders(c *gin.Context, params oapi_codegen.OrdersListOrdersParams) {
	if params.UserId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 12, Message: `"user_id" query paramater must be provided and must not be empty`}},
		})
		return
	}

	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if accessToken.SubjectType == shared_api.SubjectTypeUser && params.UserId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	res, err := api.Service.ListOrders(c.Request.Context(), params)
	if err != nil {
		api.Logger.Error("list orders", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "failed to list orders"}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (api *ApiImpl) OrdersCreateOrder(c *gin.Context) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	if !slices.Contains([]string{shared_api.SubjectTypeUser}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	userId := accessToken.SubjectId

	createOrderRes, err := api.Service.CreateOrder(c.Request.Context(), userId)
	if err != nil {
		api.Logger.Error("create order operation and publish request cart contents", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "failed to create order operation and place order"}},
		})
		return
	}

	c.JSON(http.StatusOK, createOrderRes)
}

func (api *ApiImpl) OrdersGetOrder(c *gin.Context, orderId string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	res, err := api.Service.GetOrder(c.Request.Context(), orderId)
	if err != nil {
		api.Logger.Error("get order", zap.String("id", orderId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 125, Message: "failed to get order"}},
		})
		return
	}

	if res == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 125, Message: "order not found"}},
		})
		return
	}
	if accessToken.SubjectType == shared_api.SubjectTypeUser && res.UserId != accessToken.SubjectId {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (api *ApiImpl) OrdersUpdateOrder(c *gin.Context, orderId string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems"}},
		})
		return
	}

	var requestBody oapi_codegen.OrdersUpdateOrderReq
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "invalid request body"}},
		})
		return
	}

	res, err := api.Service.UpdateOrder(c.Request.Context(), requestBody, orderId, accessToken.SubjectType)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
			})
			return
		}
		if errors.Is(err, service.ErrOrderStateMachineInvalidStatus) {
			c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 125, Message: err.Error()}},
			})
			return
		}
		if errors.Is(err, service.ErrOrderStateMachineIncorrentStatusTransition) {
			c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 126, Message: err.Error()}},
			})
			return
		}

		api.Logger.Error("update order", zap.String("id", orderId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "failed to update order"}},
		})
		return
	}

	if res == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 125, Message: "order not found"}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
