package store

import (
	"context"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
)

const (
	tableOperations = `orders/operations`

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

var queryGetOperation = template.ReplaceAllPairs(`
`,
	"{{table.operations}}",
	tableOperations,
)

var queryCreateOperation = template.ReplaceAllPairs(`
`,
	"{{table.operations}}",
	tableOperations,
)

var queryUpdateOperation = template.ReplaceAllPairs(`
`,
	"{{table.operations}}",
	tableOperations,
)
