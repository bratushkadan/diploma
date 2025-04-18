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
	out := make([]oapi_codegen.CartGetCartPositionsResPosition, 0)

	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

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
	positions := make(map[string][]oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqCartPosition, len(messages))

	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	var userIds []types.Value
	for _, msg := range messages {
		userIds = append(userIds, types.UTF8Value(msg.UserId))
	}

	if err := c.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryGetCartPositionsMany, table.NewQueryParameters(
			table.ValueParam("$user_ids", types.ListValue(userIds...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var userId string
				var pos oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqCartPosition
				var count uint32
				if err := res.ScanNamed(
					named.Required("user_id", &userId),
					named.Required("product_id", &pos.ProductId),
					named.Required("count", &count),
				); err != nil {
					return err
				}
				pos.Count = int(count)

				positions[userId] = append(positions[userId], pos)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	out := make([]oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqMessage, 0, len(messages))
	for _, msg := range messages {
		cartPositions, ok := positions[msg.UserId]
		if !ok {
			cartPositions = make([]oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqCartPosition, 0)
		}
		out = append(out, oapi_codegen.PrivateOrderProcessPublishedCartPositionsReqMessage{
			OperationId:   msg.OperationId,
			CartPositions: cartPositions,
		})
	}

	return out, nil
}

var queryClear = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id
RETURNING product_id, count;
`, "{{table.cart}}", tableCart)

func (c *Cart) Clear(ctx context.Context, userId string) ([]oapi_codegen.CartClearCartResPosition, error) {
	out := make([]oapi_codegen.CartClearCartResPosition, 0)

	if err := c.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryClear, table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userId)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var pos oapi_codegen.CartClearCartResPosition
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

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryClearMany = template.ReplaceAllPairs(`
DECLARE $user_ids AS List<Utf8>;

DELETE FROM {{table.cart}}
WHERE user_id IN $user_ids
RETURNING user_id;
`, "{{table.cart}}", tableCart)

// Private endpoint version for messages array
func (c *Cart) ClearMany(ctx context.Context, messages []oapi_codegen.PrivateClearCartPositionsReqMessage) error {
	var userIds []types.Value
	for _, msg := range messages {
		userIds = append(userIds, types.UTF8Value(msg.UserId))
	}

	cartClearedUserIdsList := make([]string, len(messages))

	if err := c.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryClearMany, table.NewQueryParameters(
			table.ValueParam("$user_ids", types.ListValue(userIds...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		cartClearedUserIds := make(map[string]struct{}, len(messages))
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var userId string
				if err := res.ScanNamed(
					named.Required("user_id", &userId),
				); err != nil {
					return err
				}
				cartClearedUserIds[userId] = struct{}{}
			}
		}

		for userId := range cartClearedUserIds {
			cartClearedUserIdsList = append(cartClearedUserIdsList, userId)
		}

		return res.Err()
	}); err != nil {
		return err
	}

	c.logger.Info("clear user carts", zap.Any("user_ids", cartClearedUserIdsList))

	return nil
}

var queryDeleteCartPosition = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $product_id AS Utf8;

DELETE FROM {{table.cart}}
WHERE user_id = $user_id AND product_id = $product_id
RETURNING product_id, count;`, "{{table.cart}}", tableCart)

func (c *Cart) DeleteCartPosition(ctx context.Context, userId, productId string) (*oapi_codegen.CartDeleteCartPositionResPosition, error) {
	var out *oapi_codegen.CartDeleteCartPositionResPosition

	if err := c.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryDeleteCartPosition, table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userId)),
			table.ValueParam("$product_id", types.UTF8Value(productId)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var pos oapi_codegen.CartDeleteCartPositionResPosition
				var count uint32
				if err := res.ScanNamed(
					named.Required("product_id", &pos.ProductId),
					named.Required("count", &count),
				); err != nil {
					return err
				}
				pos.Count = int(count)
				out = &pos
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
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
	var out oapi_codegen.CartSetCartPositionResPosition

	if err := c.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, querySetCartPosition, table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userId)),
			table.ValueParam("$product_id", types.UTF8Value(productId)),
			table.ValueParam("$count", types.Uint32Value(uint32(count))),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var count uint32
				if err := res.ScanNamed(
					named.Required("product_id", &out.ProductId),
					named.Required("count", &count),
				); err != nil {
					return err
				}
				out.Count = int(count)
			}
		}

		return res.Err()
	}); err != nil {
		return oapi_codegen.CartSetCartPositionResPosition{}, err
	}

	return out, nil
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

	c.logger.Info("publish cart contents", zap.Any("messages", marshaledMessages))
	return ydbtopic.Produce(ctx, c.topicCartContents, marshaledMessages...)
}
