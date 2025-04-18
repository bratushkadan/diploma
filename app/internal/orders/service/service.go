package service

import (
	"errors"

	"github.com/bratushkadan/floral/internal/orders/store"
	"go.uber.org/zap"
)

type Orders struct {
	l     *zap.Logger
	store *store.Orders

	yoomoneyPaymentNotificationSecret string
}

type OrdersBuilder struct {
	svc Orders
}

func NewBuilder() *OrdersBuilder {
	return &OrdersBuilder{}
}

func (b *OrdersBuilder) Logger(l *zap.Logger) *OrdersBuilder {
	b.svc.l = l
	return b
}
func (b *OrdersBuilder) Store(store *store.Orders) *OrdersBuilder {
	b.svc.store = store
	return b
}

func (b *OrdersBuilder) YoomoneyPaymentNotificationSecret(secret string) *OrdersBuilder {
	b.svc.yoomoneyPaymentNotificationSecret = secret
	return b
}

func (b *OrdersBuilder) Build() (*Orders, error) {
	if b.svc.store == nil {
		return nil, errors.New("store is nil")
	}

	if b.svc.l == nil {
		b.svc.l = zap.NewNop()
	}

	if b.svc.yoomoneyPaymentNotificationSecret == "" {
		return nil, errors.New("payment notification secret for Yoomoney is empty")
	}

	return &b.svc, nil
}

func ptr[T any](v T) *T {
	return &v
}
