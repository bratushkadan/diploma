package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/store"
	"github.com/google/uuid"
)

const (
	OperationTypeCreateOrder = "create_order"

	OperationTypeCreateOrderStatusStarted    = "started"
	OperationTypeCreateOrderStatusAborted    = "aborted"
	OperationTypeCreateOrderStatusTerminated = "terminated"
	OperationTypeCreateOrderStatusCompleted  = "completed"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
)

func (s *Orders) ListOrders(ctx context.Context, req oapi_codegen.OrdersListOrdersParams) (oapi_codegen.OrdersListOrdersRes, error) {
	return s.store.ListOrders(ctx, req.UserId, req.NextPageToken)
}

func (s *Orders) CreateOrder(ctx context.Context, userId string) (oapi_codegen.OrdersCreateOrderRes, error) {
	operation, err := s.store.CreateOperation(ctx, store.CreateOperationDTOInput{
		Id:        uuid.NewString(),
		Type:      OperationTypeCreateOrder,
		Status:    OperationTypeCreateOrderStatusStarted,
		UserId:    userId,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return oapi_codegen.OrdersCreateOrderRes{}, fmt.Errorf("create operation: %v", err)
	}

	if err := s.store.ProducePublishCartContentsRequest(ctx, operation.Id, userId); err != nil {
		return oapi_codegen.OrdersCreateOrderRes{}, fmt.Errorf("publish get cart contents request: %v", err)
	}

	return oapi_codegen.OrdersCreateOrderRes{Operation: operation}, nil
}

func (s *Orders) GetOrder(ctx context.Context, orderId string) (*oapi_codegen.OrdersGetOrderRes, error) {
	return s.store.GetOrder(ctx, orderId)
}

func (s *Orders) UpdateOrder(ctx context.Context, req oapi_codegen.OrdersUpdateOrderReq, subjectType, subjectId string) (*oapi_codegen.OrdersUpdateOrderReq, error) {

	return nil, nil
}

func (s *Orders) ProcessPublishedCartPositions(ctx context.Context, req oapi_codegen.PrivateOrderProcessPublishedCartPositionsReq) error {
	return nil
}

func (s *Orders) ProcessReservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody) error {
	return nil
}

func (s *Orders) ProcessUnreservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessUnreservedProductsJSONRequestBody) error {
	return nil
}

func (s *Orders) BatchCancelUnpaidOrders(ctx context.Context, req oapi_codegen.PrivateOrdersBatchCancelUnpaidOrdersJSONRequestBody) error {
	return nil
}
