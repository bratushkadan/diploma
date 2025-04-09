package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"go.uber.org/zap"
)

const (
	tableCart = "`cart/positions`"

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
	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	out := make([]oapi_codegen.CartGetCartPositionsResPosition, 0)

	if err := c.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryGetCartPositions, table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userId)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var pos oapi_codegen.CartGetCartPositionsResPosition
				var count uint32
				if err := res.ScanNamed(
					named.Required("product_id", &pos.ProductId),
					named.Required("count", &count),
				); err != nil {
					return err
				}
				pos.Count = int(count)

				out = append(out, pos)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
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
	return nil, nil
}

var queryClear = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id
RETURNING product_id, count;
`, "{{table.cart}}", tableCart)

func (c *Cart) Clear(ctx context.Context, userId string) ([]oapi_codegen.CartClearCartResPosition, error) {
	return nil, nil
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
	return nil
}

var deleteCartPosition = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $product_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id AND product_id = $product_id
RETURNING product_id, count;`, "{{table.cart}}", tableCart)

func (c *Cart) DeleteCartPosition(ctx context.Context, userId, productId string) (oapi_codegen.CartDeleteCartPositionResPosition, error) {
	return oapi_codegen.CartDeleteCartPositionResPosition{}, nil
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
	return oapi_codegen.CartSetCartPositionResPosition{}, nil
}

func (c *Cart) PublishCartContents(ctx context.Context, messages []oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqMessage) error {
	marshaledMessages := make([][]byte, 0, len(messages))
	for _, message := range messages {
		marshaledPosition, err := json.Marshal(&message)
		if err != nil {
			return fmt.Errorf("marshal cart contents message: %w", err)
		}
		marshaledMessages = append(marshaledMessages, marshaledPosition)
	}

	return ydbtopic.Produce(ctx, c.topicCartContents, marshaledMessages...)
}
