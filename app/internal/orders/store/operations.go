package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

const (
	tableOperations = "`orders/operations`"

	topicCancelOperations = "orders/cancel_operations_topic"
)

var queryGetOperation = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;

SELECT
	id, type, status, details, user_id, order_id, created_at, updated_at
FROM
	{{table.operations}}
WHERE id = $id;
`,
	"{{table.operations}}",
	tableOperations,
)

func (s *Orders) GetOperation(ctx context.Context, operationId string) (*oapi_codegen.OrdersGetOperationRes, error) {
	var out *oapi_codegen.OrdersGetOperationRes

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryGetOperation, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(operationId)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				out = &oapi_codegen.OrdersGetOperationRes{}
				var createdAt, updatedAt time.Time
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("type", &out.Type),
					named.Required("status", &out.Status),
					named.Optional("details", &out.Details),
					named.Required("user_id", &out.UserId),
					named.Optional("order_id", &out.OrderId),
					named.Required("created_at", &createdAt),
					named.Required("updated_at", &updatedAt),
				); err != nil {
					return err
				}
				out.CreatedAt = createdAt.Format(time.RFC3339)
				out.UpdatedAt = updatedAt.Format(time.RFC3339)
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryCreateOperation = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $type AS Utf8;
DECLARE $status AS Utf8;
DECLARE $details AS Optional<Utf8>;
DECLARE $user_id AS Utf8;
DECLARE $order_id AS Optional<Utf8>;
DECLARE $created_at AS Timestamp;
DECLARE $updated_at AS Timestamp;

INSERT INTO {{table.operations}} (id, type, status, details, user_id, order_id, created_at, updated_at)
VALUES
($id, $type, $status, $details, $user_id, $order_id, $created_at, $updated_at)
RETURNING id, type, status, user_id, order_id, created_at, updated_at;
`,
	"{{table.operations}}",
	tableOperations,
)

type CreateOperationDTOInput struct {
	Id        string
	Type      string
	Status    string
	Details   *string
	UserId    string
	OrderId   *string
	CreatedAt time.Time
}

func (s *Orders) CreateOperation(ctx context.Context, in CreateOperationDTOInput) (oapi_codegen.OrdersCreateOrderResOperation, error) {
	var out oapi_codegen.OrdersCreateOrderResOperation

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryCreateOperation, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(in.Id)),
			table.ValueParam("$type", types.UTF8Value(in.Type)),
			table.ValueParam("$status", types.UTF8Value(in.Status)),
			table.ValueParam("$details", types.NullableUTF8Value(in.Details)),
			table.ValueParam("$user_id", types.UTF8Value(in.UserId)),
			table.ValueParam("$order_id", types.NullableUTF8Value(in.OrderId)),
			table.ValueParam("$created_at", types.TimestampValueFromTime(in.CreatedAt)),
			table.ValueParam("$updated_at", types.TimestampValueFromTime(in.CreatedAt)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var createdAt, updatedAt time.Time
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("type", &out.Type),
					named.Required("status", &out.Status),
					named.Required("user_id", &out.UserId),
					named.Optional("order_id", &out.OrderId),
					named.Required("created_at", &createdAt),
					named.Required("updated_at", &updatedAt),
				); err != nil {
					return err
				}
				out.CreatedAt = createdAt.Format(time.RFC3339)
				out.UpdatedAt = updatedAt.Format(time.RFC3339)
			}
		}

		return res.Err()
	}); err != nil {
		return oapi_codegen.OrdersCreateOrderResOperation{}, err
	}

	return out, nil
}

func (s *Orders) PublishGetCartContentsRequest(ctx context.Context, operationId, userId string) error {
	msgBytes, err := json.Marshal(&oapi_codegen.PrivatePublishCartPositionsReqMessage{
		OperationId: operationId,
		UserId:      userId,
	})
	if err != nil {
		return fmt.Errorf("serialize publish carts contents message: %v", err)
	}
	if err := ydbtopic.Produce(ctx, s.topicCartPublishRequests, msgBytes); err != nil {
		return fmt.Errorf("publish message request publish cart contents: %v", err)
	}

	return nil
}

var queryUpdateOperation = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $details AS Optional<Utf8>;
DECLARE $updated_at AS Timestamp;

-- $id = "foo";
-- $status = Unwrap(CAST("cancelled" AS Utf8));
-- $details = CAST("not enought positions" AS Utf8);
-- $updated_at = CurrentUtcTimestamp();

$to_update = (
    SELECT
        id,
        $status AS status,
        COALESCE($details, details) AS details,
        $updated_at AS updated_at,
    FROM {{table.operations}}
    WHERE id = $id
);

UPDATE {{table.operations}} ON
SELECT * FROM $to_update
RETURNING *;
`,
	"{{table.operations}}",
	tableOperations,
)
