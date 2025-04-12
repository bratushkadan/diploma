package store

import (
	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
)

const (
	tableProducts   = "`products/products`"
	tableOrders     = "`orders/orders`"
	tableOrderItems = "`orders/order_items`"

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
DECLARE $id AS Utf8;

SELECT
    o.id AS id,
    o.user_id AS user_id,
    o.status AS status,
    o.created_at AS created_at,
    o.updated_at AS updated_at,
    i.product_id AS product_id,
    i.name AS product_name,
    i.seller_id AS product_seller_id,
    i.count AS produt_count,
    i.price AS product_price,
    i.picture AS product_picture
FROM {{table.orders}} o
JOIN {{table.order_items}} i ON i.order_id = o.id
WHERE o.id = $id;
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

func (s *Orders) GetOrder() (*oapi_codegen.OrdersGetOrderRes, error) {
	return nil, nil
}

// TODO: STALE READONLY TRANSACTION (JUST LIKE IN PRODCUTS)
var queryListOrders = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $last_paginated_order_id AS Optional<Utf8>;
DECLARE $last_paginated_created_at As Optional<Datetime>;
DECLARE $page_size AS Optional<Uint32>;

-- $user_id = UNWRAP(CAST("acd559b2-def1-4b01-b501-c642e22dd7da" AS Utf8));
-- $last_paginated_order_id = UNWRAP(CAST("foo-bar-baz-qux3" AS Utf8));
-- $last_paginated_created_at = UNWRAP(CAST("2025-04-12T18:03:40Z" AS Datetime));

SELECT
    o.id AS id,
    o.user_id AS user_id,
    o.status AS status,
    o.created_at AS created_at,
    o.updated_at AS updated_at,
    i.product_id AS product_id,
    i.name AS product_name,
    i.seller_id AS product_seller_id,
    i.count AS produt_count,
    i.price AS product_price,
    i.picture AS product_picture
FROM {{table.orders}}
VIEW idx_list_orders o
JOIN {{table.order_items}} i ON i.order_id = o.id
WHERE
    o.user_id = $user_id
        AND
    (
        (
            $last_paginated_order_id IS NULL
                OR 
            $last_paginated_order_id = o.id
        )
            OR
        (
            $last_paginated_created_at IS NULL
                OR
            $last_paginated_created_at > o.created_at
        )
    )
ORDER BY created_at DESC
LIMIT COALESCE($page_size + 1, 3u);
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

var queryCreateOrder = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $user_id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $created_at AS Datetime;
DECLARE $updated_at AS Datetime;

DECLARE $order_items AS List<Struct<
  product_id:Utf8,
  seller_id:Utf8,
  name:Utf8,
  count:Uint32,
  price:Double,
  picture:Optional<Utf8>,
>>;

INSERT INTO {{table.orders}} (id, user_id, status, created_at, updated_at)
VALUES ($id, $user_id, $status, $created_at, $updated_at);

INSERT INTO {{table.order_items}} (
    order_id, product_id, seller_id, name, count, price, picture
)
SELECT
    $id AS order_id,
    product_id,
    seller_id,
    name,
    count,
    price,
    picture
FROM
    AS_TABLE($order_items);
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

var queryUpdateOrder = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $updated_at AS Datetime;

$to_update = (
    SELECT
        id,
        $status AS status,
        $updated_at AS updated_at
    FROM
        {{table.orders}}
    WHERE id = $id
);

UPDATE {{table.orders}}
ON SELECT * FROM $to_update
RETURNING id, status, updated_at;
`,
	"{{table.orders}}",
	tableOrders,
)

func a() {
	// oapi_codegen.OrdersUpdateOrderRes
}
