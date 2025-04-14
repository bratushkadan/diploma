package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/internal/orders/store"
	shared_api "github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "created"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusCancelling OrderStatus = "cancelling"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusProcessed  OrderStatus = "processed"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCompleted  OrderStatus = "completed"
)

var (
	ErrPermissionDenied = errors.New("permission denied")

	ErrOrderStateMachineInvalidStatus             = errors.New("invalid order status")
	ErrOrderStateMachineIncorrentStatusTransition = errors.New("incorrect order status transition")
)

type OrderStateMachine struct {
	status OrderStatus
}

func NewOrderStateMachine(status OrderStatus) (*OrderStateMachine, error) {
	switch status {
	case OrderStatusCreated, OrderStatusPaid, OrderStatusCancelling, OrderStatusCancelled, OrderStatusProcessed, OrderStatusShipped, OrderStatusDelivered, OrderStatusCompleted:
		return &OrderStateMachine{status: status}, nil
	default:
		return nil, fmt.Errorf(`invalid initial order status for state machine: "%s"`, status)
	}
}

func (o *OrderStateMachine) TransitionString(newStatus string) error {
	switch newStatus {
	case string(OrderStatusCreated), string(OrderStatusPaid), string(OrderStatusCancelling), string(OrderStatusCancelled), string(OrderStatusProcessed), string(OrderStatusShipped), string(OrderStatusDelivered), string(OrderStatusCompleted):
		return o.Transition(OrderStatus(newStatus))
	default:
		return fmt.Errorf(`%w: "%s"`, ErrOrderStateMachineInvalidStatus, newStatus)
	}
}

func NewErrUnavailableTransition(fromStatus, toStatus OrderStatus, availableTransitions ...OrderStatus) error {
	strAvailableTransitions := make([]string, 0, len(availableTransitions))
	for _, t := range availableTransitions {
		strAvailableTransitions = append(strAvailableTransitions, string(t))
	}

	return fmt.Errorf(
		`%w: no status transition "%s" -> "%s", available transitions are: "%s" -> "%s"`,
		ErrOrderStateMachineIncorrentStatusTransition,
		fromStatus,
		toStatus,
		fromStatus,
		strings.Join(strAvailableTransitions, ", "),
	)
}

func NewErrNoTransitionAvailable(fromStatus OrderStatus) error {
	return fmt.Errorf(
		`%w: no available transition for status "%s"`,
		ErrOrderStateMachineIncorrentStatusTransition,
		fromStatus,
	)
}

func (o *OrderStateMachine) Transition(newStatus OrderStatus) error {
	switch o.status {
	case OrderStatusCreated:
		switch newStatus {
		case OrderStatusPaid, OrderStatusCancelling:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusPaid, OrderStatusCancelling)
		}
	case OrderStatusPaid:
		switch newStatus {
		case OrderStatusProcessed, OrderStatusCancelling:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusProcessed, OrderStatusCancelling)
		}
	case OrderStatusCancelling:
		switch newStatus {
		case OrderStatusCancelled:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusCancelled)
		}
	case OrderStatusCancelled:
		return NewErrNoTransitionAvailable(o.status)
	case OrderStatusProcessed:
		switch newStatus {
		case OrderStatusShipped, OrderStatusCancelling:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusShipped, OrderStatusCancelling)
		}
	case OrderStatusShipped:
		switch newStatus {
		case OrderStatusDelivered, OrderStatusCompleted:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusDelivered, OrderStatusCompleted)
		}
	case OrderStatusDelivered:
		switch newStatus {
		case OrderStatusCompleted:
			o.status = newStatus
			return nil
		default:
			return NewErrUnavailableTransition(o.status, newStatus, OrderStatusCompleted)
		}
	case OrderStatusCompleted:
		return NewErrNoTransitionAvailable(o.status)
	default:
		return fmt.Errorf(`%w: unknown status to transition from: "%s"`, ErrOrderStateMachineIncorrentStatusTransition, o.status)
	}
}

func (o *OrderStateMachine) Status() OrderStatus {
	return o.status
}

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

	order, err := s.store.GetOrder(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("retrieve order: %v", err)
	}
	// not found
	if order == nil {
		return nil, nil
	}

	orderStateMachine, err := NewOrderStateMachine(OrderStatus(order.Status))
	if err != nil {
		return nil, fmt.Errorf("new order state machine: %v", err)
	}

	if err := orderStateMachine.TransitionString(req.Status); err != nil {
		return nil, err
	}

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
		go func() {
			if err := s.store.ProduceCancelOperationMessages(ctx, cancelOperationsMessages...); err != nil {
				defer m.Unlock()
				m.Lock()

				errs = append(errs, err)
			}
		}()
	}

	wg.Wait()

	if err := s.store.ProduceProductsReservationMessages(ctx, productsReservationMessages...); err != nil {
		return fmt.Errorf("produce products reservation messages: %v", err)
	}

	return nil
}

func (s *Orders) ProcessReservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessReservedProductsJSONRequestBody) error {
	ops := make([]store.UpdateOperationManyDTOInputOperation, 0, len(req.Messages))
	for _, message := range req.Messages {
		ops = append(ops, store.UpdateOperationManyDTOInputOperation{
			Id:        message.OperationId,
			Status:    OperationTypeCreateOrderStatusCompleted,
			OrderId:   ptr(uuid.NewString()),
			UpdatedAt: time.Now(),
		})
	}

	updateOpsManyRes, err := s.store.UpdateOperationMany(ctx, store.UpdateOperationManyDTOInput{Operations: ops})
	if err != nil {
		return fmt.Errorf("update operations many: %v", err)
	}

	products := make(map[string][]oapi_codegen.PrivateOrderProcessReservedProductsReqProduct, len(req.Messages))
	for _, msg := range req.Messages {
		products[msg.OperationId] = msg.Products
	}

	orders := make([]store.CreateOrderManyDTOInputOrder, 0, len(req.Messages))
	for _, opUpdate := range updateOpsManyRes.OperationsUpdates {
		orders = append(orders, store.CreateOrderManyDTOInputOrder{
			Id:        *opUpdate.OrderId,
			UserId:    opUpdate.UserId,
			Status:    string(OrderStatusCreated),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Products:  products[opUpdate.OperationId],
		})
	}

	_, err = s.store.CreateOrderMany(ctx, store.CreateOrderManyDTOInput{Orders: orders})
	if err != nil {
		return fmt.Errorf("create order many: %v", err)
	}

	clearCartMessages := make([]oapi_codegen.PrivateClearCartPositionsReqMessage, 0, len(req.Messages))

	if err := s.store.ProduceCartClearMessages(ctx, clearCartMessages...); err != nil {
		return fmt.Errorf("publish clear cart messages: %v", err)
	}

	return nil
}

func (s *Orders) ProcessUnreservedProducts(ctx context.Context, req oapi_codegen.PrivateOrdersProcessUnreservedProductsJSONRequestBody) error {
	orderUpdates := make([]store.UpdateOrderManyDTOInputOrderUpdate, 0, len(req.Messages))

	for _, msg := range req.Messages {
		orderUpdates = append(orderUpdates, store.UpdateOrderManyDTOInputOrderUpdate{
			OrderId:   msg.OrderId,
			Status:    string(OrderStatusCancelled),
			UpdatedAt: time.Now(),
		})
	}

	_, err := s.store.UpdateOrderMany(ctx, store.UpdateOrderManyDTOInput{
		OrderUpdates: orderUpdates,
	})
	if err != nil {
		return fmt.Errorf("update orders: %v", err)
	}

	return nil
}

func (s *Orders) BatchCancelUnpaidOrders(ctx context.Context, req oapi_codegen.PrivateOrderBatchCancelUnpaidOrdersReq) error {
	messages := make([]oapi_codegen.PrivateUnreserveProductsReqMessage, 0)

	res, err := s.store.ListUnpaidOrders(ctx)
	if err != nil {
		return fmt.Errorf("list unpaid orders: %v", err)
	}
	unpaidOrders := res.Orders

	orderUpdates := make([]store.UpdateOrderManyDTOInputOrderUpdate, 0, len(unpaidOrders))
	for _, order := range unpaidOrders {
		orderUpdates = append(orderUpdates, store.UpdateOrderManyDTOInputOrderUpdate{
			OrderId:   order.Id,
			Status:    string(OrderStatusCancelling),
			UpdatedAt: time.Now(),
		})
	}

	_, err = s.store.UpdateOrderMany(ctx, store.UpdateOrderManyDTOInput{OrderUpdates: orderUpdates})
	if err != nil {
		return fmt.Errorf("update orders: %v", err)
	}

	for _, order := range unpaidOrders {
		products := make([]oapi_codegen.PrivateUnreserveProductsReqProduct, 0)
		for _, item := range order.Items {
			products = append(products, oapi_codegen.PrivateUnreserveProductsReqProduct{
				Id:    item.Id,
				Count: item.Count,
			})
		}
		messages = append(messages, oapi_codegen.PrivateUnreserveProductsReqMessage{
			OrderId:  order.Id,
			Products: products,
		})
	}

	if err := s.store.ProduceProductsUnreservationMessages(ctx, messages...); err != nil {
		return fmt.Errorf("publish products unreservation messages: %v", err)
	}

	return nil
}
