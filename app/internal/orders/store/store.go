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

	topicCartPublishRequests, err := ydbtopic.NewProducer(b.store.db, topicCartPublishRequests)
	if err != nil {
		return nil, errors.New("setup CartPublishRequsts topic: %w")
	}
	b.store.topicCartPublishRequests = topicCartPublishRequests

	topicCartClearRequests, err := ydbtopic.NewProducer(b.store.db, topicCartClearRequests)
	if err != nil {
		return nil, errors.New("setup CartClearRequests topic: %w")
	}
	b.store.topicCartClearRequests = topicCartClearRequests

	topicProductsReservations, err := ydbtopic.NewProducer(b.store.db, topicProductsReservations)
	if err != nil {
		return nil, errors.New("setup ProductsReservations topic: %w")
	}
	b.store.topicProductsReservations = topicProductsReservations

	topicProductsUnreservations, err := ydbtopic.NewProducer(b.store.db, topicProductsUnreservations)
	if err != nil {
		return nil, errors.New("setup ProductsUnreservations topic: %w")
	}
	b.store.topicProductsUnreservations = topicProductsUnreservations

	topicCancelOperations, err := ydbtopic.NewProducer(b.store.db, topicCancelOperations)
	if err != nil {
		return nil, errors.New("setup ProductsUnreservations topic: %w")
	}
	b.store.topicCancelOperations = topicCancelOperations

	topicProcessedPaymentsNotifications, err := ydbtopic.NewProducer(b.store.db, topicProcessedPaymentsNotifications)
	if err != nil {
		return nil, errors.New("setup ProductsUnreservations topic: %w")
	}
	b.store.topicProcessedPaymentsNotifications = topicProcessedPaymentsNotifications

	if b.store.logger == nil {
		b.store.logger = zap.NewNop()
	}

	return &b.store, nil
}

type Orders struct {
	db     *ydb.Driver
	logger *zap.Logger

	topicCartPublishRequests *topicwriter.Writer
	topicCartClearRequests   *topicwriter.Writer

	topicProductsReservations   *topicwriter.Writer
	topicProductsUnreservations *topicwriter.Writer

	topicCancelOperations *topicwriter.Writer

	topicProcessedPaymentsNotifications *topicwriter.Writer
}
