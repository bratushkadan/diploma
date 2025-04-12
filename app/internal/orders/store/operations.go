package store

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
)

const (
	tableOperations = "`orders/operations`"

	topicCancelOperations = "orders/cancel_operations_topic"
)

var queryUnreserveProducts = template.ReplaceAllPairs(`
DECLARE $updates AS List<Struct<
    id:String,
    stock:Uint32,
>>;

UPDATE {{table.table_products}} ON
SELECT * FROM 
(
    SELECT
        p.id AS id,
        u.stock AS stock,
        CurrentUtcDatetime() AS updated_at,
    FROM {{table.table_products}} p
    JOIN AS_TABLE($updates) u ON p.id = u.id
)
RETURNING id, stock, updated_at;
`,
	"{{table.table_products}}", tableProducts,
)

func (s *Orders) UnreserveProducts(ctx context.Context, messages []oapi_codegen.PrivateUnreserveProductsReqMessage) error {
	ydbtopic.Produce(ctx, nil)
	_ = queryUnreserveProducts
	return nil
}

// Careful with "store"s, here Service's own DTOs are required I think

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

var queryCreateOperation = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $type AS Utf8;
DECLARE $status AS Utf8;
DECLARE $details AS Optional<Utf8>;
DECLARE $user_id AS Utf8;
DECLARE $order_id AS Utf8;
DECLARE $created_at AS Timestamp;
DECLARE $updated_at AS Timestamp;

INSERT INTO {{table.operations}} (id, type, status, details, user_id, order_id, created_at, updated_at)
VALUES
($id, $type, $status, $details, $user_id, $order_id, $created_at, $updated_at)
RETURNING *;
`,
	"{{table.operations}}",
	tableOperations,
)

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
