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

func (s *Orders) ProduceCancelOperationMessages(ctx context.Context, messages ...oapi_codegen.PrivateOrderCancelOperationsReqMessage) error {
	dataBytes := make([][]byte, 0, len(messages))
	for _, message := range messages {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("serialize cancel operation message: %v", err)
		}
		dataBytes = append(dataBytes, msgBytes)
	}

	if err := ydbtopic.Produce(ctx, s.topicCancelOperations, dataBytes...); err != nil {
		return fmt.Errorf("publish messages cancel operations: %v", err)
	}
	return nil
}

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

func (s *Orders) ProducePublishCartContentsRequest(ctx context.Context, operationId, userId string) error {
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
DECLARE $order_id AS Optional<Utf8>;
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
				COALESCE($order_id, order_id) AS order_id,
        $updated_at AS updated_at,
    FROM {{table.operations}}
    WHERE id = $id
);

UPDATE {{table.operations}} ON
SELECT * FROM $to_update
RETURNING id, status, details, order_id, updated_at;
`,
	"{{table.operations}}",
	tableOperations,
)

type UpdateOperationDTOInput struct {
	Id        string
	Status    string
	Details   *string
	OrderId   *string
	UpdatedAt time.Time
}
type UpdateOperationDTOOutput struct {
	Id        string
	Status    string
	Details   *string
	OrderId   *string
	UpdatedAt time.Time
}

func (s *Orders) UpdateOperation(ctx context.Context, in UpdateOperationDTOInput) (*UpdateOperationDTOOutput, error) {
	var out *UpdateOperationDTOOutput

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryUpdateOperation, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(in.Id)),
			table.ValueParam("$status", types.UTF8Value(in.Status)),
			table.ValueParam("$details", types.NullableUTF8Value(in.Details)),
			table.ValueParam("$order_id", types.NullableUTF8Value(in.OrderId)),
			table.ValueParam("$updated_at", types.TimestampValueFromTime(in.UpdatedAt)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				out = &UpdateOperationDTOOutput{}
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("status", &out.Status),
					named.Optional("details", &out.Details),
					named.Optional("order_id", &out.OrderId),
					named.Required("updated_at", &out.UpdatedAt),
				); err != nil {
					return err
				}
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

// queryUpdateOperationManyTestData
/*
-- $operations = AsList(
--     AsStruct(
--       UNWRAP(CAST("364b00de-64db-4186-81ba-7ef1be9964e3" AS Utf8)) AS id,
--       UNWRAP(CAST("completed" AS Utf8)) as status,
--       NULL as details,
--       UNWRAP(CAST("order-1" AS Utf8)) AS order_id,
--       CurrentUtcTimestamp() AS updated_at,
--     ),
--     AsStruct(
--       UNWRAP(CAST("40170abd-1137-4e9f-9b28-a20f85ee0235" AS Utf8)) AS id,
--       UNWRAP(CAST("completed" AS Utf8)) as status,
--       NULL as details,
--       UNWRAP(CAST("order-2" AS Utf8)) AS order_id,
--       CurrentUtcTimestamp() AS updated_at,
--     ),
--     AsStruct(
--       UNWRAP(CAST("d2409b05-700c-4ad8-bf5e-d3ae2805644e" AS Utf8)) AS id,
--       UNWRAP(CAST("completed" AS Utf8)) as status,
--       NULL as details,
--       UNWRAP(CAST("order-3" AS Utf8)) AS order_id,
--       CurrentUtcTimestamp() AS updated_at,
--     ),
-- );
*/

var queryUpdateOperationMany = template.ReplaceAllPairs(`
DECLARE $operations AS List<Struct<
  id:Utf8,
  status:Utf8,
  details:Optional<Utf8>,
  order_id:Optional<Utf8>,
  updated_at:Timestamp,
>>;

$to_update = (
    SELECT
        u.id AS id,
        u.status AS status,
        COALESCE(u.details, o.details) AS details,
        COALESCE(u.order_id, o.order_id) AS order_id,
        u.updated_at AS updated_at,
    FROM AS_TABLE($operations) u
    JOIN {{table.operations}} o ON o.id = u.id
);

UPDATE {{table.operations}} ON
SELECT * FROM $to_update
RETURNING id, user_id, status, details, order_id, updated_at;
`,
	"{{table.operations}}",
	tableOperations,
)

type UpdateOperationManyDTOInput struct {
	Operations []UpdateOperationManyDTOInputOperation
}
type UpdateOperationManyDTOInputOperation struct {
	Id        string
	Status    string
	Details   *string
	OrderId   *string
	UpdatedAt time.Time
}
type UpdateOperationManyDTOOutput struct {
	OperationsUpdates []UpdateOperationManyDTOOutputOperationUpdate
}
type UpdateOperationManyDTOOutputOperationUpdate struct {
	OperationId string
	UserId      string
	Status      string
	Details     *string
	OrderId     *string
	UpdatedAt   time.Time
}

func (s *Orders) UpdateOperationMany(ctx context.Context, in UpdateOperationManyDTOInput) (UpdateOperationManyDTOOutput, error) {
	out := UpdateOperationManyDTOOutput{
		OperationsUpdates: make([]UpdateOperationManyDTOOutputOperationUpdate, 0, len(in.Operations)),
	}

	operations := make([]types.Value, 0, len(in.Operations))
	for _, op := range in.Operations {
		operations = append(operations, types.StructValue(
			types.StructFieldValue("id", types.UTF8Value(op.Id)),
			types.StructFieldValue("status", types.UTF8Value(op.Status)),
			types.StructFieldValue("details", types.NullableUTF8Value(op.Details)),
			types.StructFieldValue("order_id", types.NullableUTF8Value(op.OrderId)),
			types.StructFieldValue("updated_at", types.TimestampValueFromTime(op.UpdatedAt)),
		))
	}

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryUpdateOperationMany, table.NewQueryParameters(
			table.ValueParam("$operations", types.ListValue(operations...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var opUpdate UpdateOperationManyDTOOutputOperationUpdate
				if err := res.ScanNamed(
					named.Required("id", &opUpdate.OperationId),
					named.Required("user_id", &opUpdate.UserId),
					named.Required("status", &opUpdate.Status),
					named.Optional("details", &opUpdate.Details),
					named.Optional("order_id", &opUpdate.OrderId),
					named.Required("updated_at", &opUpdate.UpdatedAt),
				); err != nil {
					return err
				}
				out.OperationsUpdates = append(out.OperationsUpdates, opUpdate)
			}
		}

		return res.Err()
	}); err != nil {
		return out, err
	}

	return out, nil
}
