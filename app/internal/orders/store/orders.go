package store

import (
	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
)

const (
	tableProducts   = "products/products"
	tableOrders     = "orders/orders"
	tableOrderItems = "orders/order_items"

	topicProductsReservations   = "products/products_reservartions_topic"
	topicProductsUnreservations = "products/products_unreservartions_topic"

	topicCartContents        = "cart/cart_contents_topic"
	topicCartPublishRequests = "cart/cart_publish_requests_topic"
	topicCartClearRequests   = "cart/cart_clear_requests_topic"
)

var queryUnreserveProducts2 = template.ReplaceAllPairs(`
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

var queryGetOrder = template.ReplaceAllPairs(`
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

func (s *Orders) GetOrder() (*oapi_codegen.OrdersGetOrderRes, error) {
	return nil, nil
}

var queryListOrders = template.ReplaceAllPairs(`
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

var queryCreateOrder = template.ReplaceAllPairs(`
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

var queryUpdateOrder = template.ReplaceAllPairs(`
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)
