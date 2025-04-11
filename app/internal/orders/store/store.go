package store

import (
	"errors"

	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"go.uber.org/zap"
)

type OrdersBuilder struct {
	store Orders
}

func NewBuilder() *OrdersBuilder {
	return &OrdersBuilder{}
}

func (b *OrdersBuilder) Ydb(db *ydb.Driver) *OrdersBuilder {
	b.store.db = db
	return b
}
func (b *OrdersBuilder) Logger(l *zap.Logger) *OrdersBuilder {
	b.store.logger = l
	return b
}

func (b *OrdersBuilder) Build() (*Orders, error) {
	if b.store.db == nil {
		return nil, errors.New("ydb driver is nil")
	}

	topicCartContents, err := ydbtopic.NewProducer(b.store.db, topicA)
	if err != nil {
		return nil, errors.New("setup <> topic: %w")
	}
	b.store.topicA = topicCartContents

	if b.store.logger == nil {
		b.store.logger = zap.NewNop()
	}

	return &b.store, nil
}

type Orders struct {
	db     *ydb.Driver
	logger *zap.Logger

	topicA *topicwriter.Writer
}
