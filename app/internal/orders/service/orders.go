package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/store"
	shared_api "github.com/bratushkadan/floral/pkg/shared/api"
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

func (s *Orders) UpdateOrder(ctx context.Context, req oapi_codegen.OrdersUpdateOrderReq, orderId, subjectType string) (*oapi_codegen.OrdersUpdateOrderRes, error) {
	if !slices.Contains([]string{shared_api.SubjectTypeSeller, shared_api.SubjectTypeAdmin}, subjectType) {
		return nil, ErrPermissionDenied
	}

	// TODO: validate status update check here - invariants according to the state diagram only!

	orderUpdateRes, err := s.store.UpdateOrder(ctx, req.Status, orderId)
	if err != nil {
		return nil, fmt.Errorf("update order: %v", err)
	}

	return orderUpdateRes, nil
}

func (s *Orders) ProcessPublishedCartPositions(ctx context.Context, req oapi_codegen.PrivateOrderProcessPublishedCartPositionsReq) error {
	var productsReservationMessages []oapi_codegen.PrivateReserveProductsReqMessage
	var cancelOperationsMessages []oapi_codegen.PrivateOrderCancelOperationsReqMessage

	for _, message := range req.Messages {
		if len(message.CartPositions) == 0 {
			cancelOperationsMessages = append(cancelOperationsMessages, oapi_codegen.PrivateOrderCancelOperationsReqMessage{
				OperationId: message.OperationId,
				Details:     "error creating order: cart is empty",
			})
			continue
		}

		products := make([]oapi_codegen.PrivateReserveProductsReqProduct, 0, len(message.CartPositions))
		for _, pos := range message.CartPositions {
			products = append(products, oapi_codegen.PrivateReserveProductsReqProduct{
				Id:    pos.ProductId,
				Count: pos.Count,
			})
		}

		productsReservationMessages = append(productsReservationMessages, oapi_codegen.PrivateReserveProductsReqMessage{
			OperationId: message.OperationId,
			Products:    products,
		})
	}

	var wg sync.WaitGroup
	var errs []error
	var m sync.Mutex
	if len(productsReservationMessages) > 0 {
		wg.Add(1)
		go func() {
			if err := s.store.ProduceProductsReservationMessages(ctx, productsReservationMessages...); err != nil {
				defer m.Unlock()
				m.Lock()

				errs = append(errs, err)
			}
		}()
	}
	if len(cancelOperationsMessages) > 0 {
		wg.Add(1)
		go func() {}()
	}

	wg.Wait()

	s.store.ProduceProductsReservationMessages(ctx, productsReservationMessages...)

	return nil
}

// TODO: need N updates in one YQL query (cancel N operations)
// TODO: need N creates in one YQL query (create N orders)
func (s *Orders) ProcessReservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody) error {
	return nil
}

// TODO: need N updates in one YQL query (update N orders)
func (s *Orders) ProcessUnreservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessUnreservedProductsJSONRequestBody) error {
	return nil
}

// TODO: need N updates in one YQL query (update N orders)
func (s *Orders) BatchCancelUnpaidOrders(ctx context.Context, req oapi_codegen.PrivateOrderBatchCancelUnpaidOrdersReq) error {
	return nil
}
