package service

import (
	"errors"

	"github.com/bratushkadan/floral/internal/cart/store"
	"go.uber.org/zap"
)

type Cart struct {
	l     *zap.Logger
	store *store.Cart
}

type CartBuilder struct {
	svc Cart
}

func NewBuilder() *CartBuilder {
	return &CartBuilder{}
}

func (b *CartBuilder) Logger(l *zap.Logger) *CartBuilder {
	b.svc.l = l
	return b
}
func (b *CartBuilder) Store(store *store.Cart) *CartBuilder {
	b.svc.store = store
	return b
}

func (b *CartBuilder) Build() (*Cart, error) {
	if b.svc.store == nil {
		return nil, errors.New("store is nil")
	}

	if b.svc.l == nil {
		b.svc.l = zap.NewNop()
	}

	return &b.svc, nil
}
