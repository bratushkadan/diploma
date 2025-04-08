package store

import (
	"context"
	"errors"

	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"go.uber.org/zap"
)

const (
	tableCart = "`cart/cart`"

	topicCartContents = "cart/cart_contents_topic"
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

	topicCartContents, err := ydbtopic.NewProducer(b.store.db, topicCartContents)
	if err != nil {
		return nil, errors.New("setup CartContents topic: %w")
	}
	b.store.topicCartContents = topicCartContents

	if b.store.logger == nil {
		b.store.logger = zap.NewNop()
	}

	return &b.store, nil
}

type Cart struct {
	db     *ydb.Driver
	logger *zap.Logger

	topicCartContents *topicwriter.Writer
}

var queryGetCartPositions = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;

SELECT
    product_id,
    count
FROM {{table.cart}}
WHERE user_id = $user_id;
`, "{{table.cart}}", tableCart)

func (c *Cart) GetCartPositions(ctx context.Context, userId string) ([]oapi_codegen.CartGetCartPositionsResPosition, error) {
}

var queryGetCartPositionsMany = template.ReplaceAllPairs(`
DECLARE $user_ids AS List<Utf8>;

SELECT
    user_id,
    product_id,
    count
FROM {{table.cart}}
WHERE user_id IN $user_ids;
`, "{{table.cart}}", tableCart)

// Private endpoint version for messages array
func (c *Cart) GetCartPositionsMany(ctx context.Context, messages []oapi_codegen.PrivatePublishCartPositionsReqMessage) ([]oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqMessage, error) {
}

var queryClear = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id
RETURNING product_id, count;
`, "{{table.cart}}", tableCart)

func (c *Cart) Clear(ctx context.Context, userId string) ([]oapi_codegen.CartClearCartResPosition, error) {
}

var queryClearMany = template.ReplaceAllPairs(`
DECLARE $user_ids AS List<Utf8>;

DELETE FROM {{table.cart}}
WHERE user_id IN $user_ids
RETURNING user_id;
`, "{{table.cart}}", tableCart)

// Private endpoint version for messages array
func (c *Cart) ClearMany(ctx context.Context, messages []oapi_codegen.PrivateClearCartPositionsReqMessage) error {
	// After reading result with a list of IDs I need to
}

var deleteCartPosition = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $product_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id AND product_id = $product_id
RETURNING product_id, count;`, "{{table.cart}}", tableCart)

func (c *Cart) DeleteCartPosition(ctx context.Context, userId, productId string) (oapi_codegen.CartDeleteCartPositionResPosition, error) {
}

var querySetCartPosition = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $product_id AS Utf8;
DECLARE $count AS Uint32;

UPSERT INTO {{table.cart}} (user_id, product_id, count)
VALUES
($user_id, $product_id, $count)
RETURNING product_id, count;
`, "{{table.cart}}", tableCart)

func (c *Cart) SetCartPosition(ctx context.Context, userId, productId string, count int) (oapi_codegen.CartSetCartPositionResPosition, error) {
}
