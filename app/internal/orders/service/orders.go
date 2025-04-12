package service

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
)

func (s *Orders) GetOrder(ctx context.Context, orderId string) (*oapi_codegen.OrdersGetOrderRes, error) {
	return nil, nil
}
