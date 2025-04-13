package service

import (
	"context"
	"fmt"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/store"
	"github.com/google/uuid"
)

const (
	OperationTypeCreateOrder = "create_order"

	OperationTypeCreateOrderStatusStarted = "started"
)

func (s *Orders) GetOrder(ctx context.Context, orderId string) (*oapi_codegen.OrdersGetOrderRes, error) {
	return s.store.GetOrder(ctx, orderId)
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

	if err := s.store.PublishGetCartContentsRequest(ctx, operation.Id, userId); err != nil {
		return oapi_codegen.OrdersCreateOrderRes{}, fmt.Errorf("publish get cart contents request: %v", err)
	}

	return oapi_codegen.OrdersCreateOrderRes{Operation: operation}, nil
}
