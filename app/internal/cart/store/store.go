package store

import (
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/zap"
)

const (
	tableCart = "`cart/cart`"
)

type CartBuilder struct {
	store Cart
}

func NewBuilder() *CartBuilder {
	return &CartBuilder{}
}

func (b *CartBuilder) Ydb(db *ydb.Driver) *CartBuilder {
	b.store.db = db
	return b
}
func (b *CartBuilder) Logger(l *zap.Logger) *CartBuilder {
	b.store.logger = l
	return b
}

func (b *CartBuilder) Build() (*Cart, error) {
	if b.store.db == nil {
		return nil, errors.New("ydb driver is nil")
	}

	if b.store.logger == nil {
		b.store.logger = zap.NewNop()
	}

	return &b.store, nil
}

type Cart struct {
	db     *ydb.Driver
	logger *zap.Logger
}
