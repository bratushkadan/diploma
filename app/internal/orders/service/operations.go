package service

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
)

func (s *Orders) GetOperation(ctx context.Context, operationId string) (*oapi_codegen.OrdersGetOperationRes, error) {
	return s.store.GetOperation(ctx, operationId)
}

// TODO: need N updates in one YQL query (cancel N operations)
func (s *Orders) CancelOperations(ctx context.Context, req oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody) error {
	return nil
}
