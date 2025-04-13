package service

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
)

func (s *Orders) GetOperation(ctx context.Context, operationId string) (*oapi_codegen.OrdersGetOperationRes, error) {
	return s.store.GetOperation(ctx, operationId)
}
