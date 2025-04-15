package service

import (
	"context"
	"fmt"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/store"
)

const (
	OperationTypeCreateOrder = "create_order"

	OperationTypeCreateOrderStatusStarted    = "started"
	OperationTypeCreateOrderStatusAborted    = "aborted"
	OperationTypeCreateOrderStatusTerminated = "terminated"
	OperationTypeCreateOrderStatusCompleted  = "completed"
)

func (s *Orders) GetOperation(ctx context.Context, operationId string) (*oapi_codegen.OrdersGetOperationRes, error) {
	return s.store.GetOperation(ctx, operationId)
}

func (s *Orders) CancelOperations(ctx context.Context, req oapi_codegen.PrivateOrderCancelOperationsReq) error {
	ops := make([]store.UpdateOperationManyDTOInputOperation, 0, len(req.Messages))

	for _, message := range req.Messages {
		ops = append(ops, store.UpdateOperationManyDTOInputOperation{
			Id:        message.OperationId,
			Status:    OperationTypeCreateOrderStatusAborted,
			Details:   &message.Details,
			UpdatedAt: time.Now(),
		})
	}

	_, err := s.store.UpdateOperationMany(ctx, store.UpdateOperationManyDTOInput{Operations: ops})
	if err != nil {
		return fmt.Errorf("update operations: %v", err)
	}

	return nil
}
