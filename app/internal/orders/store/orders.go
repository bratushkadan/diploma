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
	tableProducts   = "`products/products`"
	tableOrders     = "`orders/orders`"
	tableOrderItems = "`orders/order_items`"

	topicProductsReservations   = "products/products_reservartions_topic"
	topicProductsUnreservations = "products/products_unreservartions_topic"

	topicCartPublishRequests = "cart/cart_contents_publish_requests_topic"
	topicCartClearRequests   = "cart/cart_clear_requests_topic"
)

const (
	ListOrdersPageSize uint32 = 10
)

func (s *Orders) ProduceProductsReservationMessages(ctx context.Context, messages ...oapi_codegen.PrivateReserveProductsReqMessage) error {
	dataBytes := make([][]byte, 0, len(messages))
	for _, message := range messages {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("serialize products reservation message: %v", err)
		}
		dataBytes = append(dataBytes, msgBytes)
	}

	if err := ydbtopic.Produce(ctx, s.topicProductsReservations, dataBytes...); err != nil {
		return fmt.Errorf("publish message products reservation: %v", err)
	}
	return nil
}

func (s *Orders) ProduceProductsUnreservationMessages(ctx context.Context, messages ...oapi_codegen.PrivateUnreserveProductsReqMessage) error {
	dataBytes := make([][]byte, 0, len(messages))
	for _, message := range messages {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("serialize products unreservation message: %v", err)
		}
		dataBytes = append(dataBytes, msgBytes)
	}

	if err := ydbtopic.Produce(ctx, s.topicProductsUnreservations, dataBytes...); err != nil {
		return fmt.Errorf("publish message products unreservation: %v", err)
	}
	return nil
}

func (s *Orders) ProduceCartClearMessages(ctx context.Context, messages ...oapi_codegen.PrivateClearCartPositionsReqMessage) error {
	dataBytes := make([][]byte, 0, len(messages))
	for _, message := range messages {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("serialize publish carts contents message: %v", err)
		}
		dataBytes = append(dataBytes, msgBytes)
	}

	if err := ydbtopic.Produce(ctx, s.topicCartPublishRequests, dataBytes...); err != nil {
		return fmt.Errorf("publish message request publish cart contents: %v", err)
	}
	return nil
}

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

func (s *Orders) GetOrder(ctx context.Context, orderId string) (*oapi_codegen.OrdersGetOrderRes, error) {
	var out *oapi_codegen.OrdersGetOrderRes

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryGetOperation, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(orderId)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				out = &oapi_codegen.OrdersGetOrderRes{}
				var orderItem oapi_codegen.OrdersGetOrderResItem
				var createdAt, updatedAt time.Time
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("user_id", &out.UserId),
					named.Required("status", &out.Status),
					named.Required("created_at", &createdAt),
					named.Required("updated_at", &updatedAt),

					named.Required("product_id", &orderItem.ProductId),
					named.Required("product_name", &orderItem.Name),
					named.Required("product_seller_id", &orderItem.SellerId),
					named.Required("product_count", &orderItem.Count),
					named.Required("product_price", &orderItem.Price),
					named.Required("optional", &orderItem.PictureUrl),
				); err != nil {
					return err
				}
				out.CreatedAt = createdAt.Format(time.RFC3339)
				out.UpdatedAt = updatedAt.Format(time.RFC3339)

				out.Items = append(out.Items, orderItem)
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryListOrders = template.ReplaceAllPairs(`
DECLARE $user_id AS Utf8;
DECLARE $last_paginated_order_id AS Optional<Utf8>;
DECLARE $last_paginated_created_at As Optional<Timestamp>;
DECLARE $page_size AS Optional<Uint32>;

-- $user_id = UNWRAP(CAST("acd559b2-def1-4b01-b501-c642e22dd7da" AS Utf8));
-- $last_paginated_order_id = UNWRAP(CAST("foo-bar-baz-qux3" AS Utf8));
-- $last_paginated_created_at = UNWRAP(CAST("2025-04-12T18:03:40Z" AS Timestamp));

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

type ListOrdersNextPage struct {
	// OrderId to start next page request with.
	OrderId *string `json:"last_page_order_id"`
	// Timestamp to start next page from
	CreatedAt *time.Time `json:"last_page_created_at"`
}
type ListOrdersNextPageDto struct {
	OrderId   string `json:"last_page_order_id"`
	CreatedAt string `json:"last_page_created_at"`
}

type ListOrdersRow struct {
	Id        string
	UserId    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time

	ProductId       string
	ProductName     string
	ProductSellerId string
	ProductCount    uint32
	ProductPrice    float64
	ProductPicture  *string
}

// TODO: понять, LIMIT чего происходит - LIMIT после JOIN'а заказов + позиций?
// FIXME: Мне нужно LIMIT заказов, а потом JOIN - нужно будет переписать запрос (но интерфейс останется таким же)
func (s *Orders) ListOrders(ctx context.Context, userId string, nextPageToken *string) (oapi_codegen.OrdersListOrdersRes, error) {
	var out oapi_codegen.OrdersListOrdersRes

	var nextPage ListOrdersNextPage
	if nextPageToken != nil {
		var nextPageDto ListOrdersNextPageDto
		if err := json.Unmarshal([]byte(*nextPageToken), &nextPageDto); err != nil {
			return oapi_codegen.OrdersListOrdersRes{}, fmt.Errorf("decode next page token: %v", err)
		}
		createdAt, err := time.Parse(time.RFC3339, nextPageDto.CreatedAt)
		if err != nil {
			return oapi_codegen.OrdersListOrdersRes{}, fmt.Errorf("parse next page token created_at as RFC3339: %v", err)
		}
		nextPage.CreatedAt = &createdAt
		nextPage.OrderId = &nextPageDto.OrderId

	}

	readTx := table.TxControl(table.BeginTx(table.WithStaleReadOnly()), table.CommitTx())

	var ordersRows []ListOrdersRow

	if err := s.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryListOrders, table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userId)),
			table.ValueParam("$last_paginated_order_id", types.NullableUTF8Value(nextPage.OrderId)),
			table.ValueParam("$last_paginated_created_at", types.NullableTimestampValueFromTime(nextPage.CreatedAt)),
			table.ValueParam("$page_size", types.NullableUint32Value(ptr(ListOrdersPageSize))),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var orderRow ListOrdersRow
				if err := res.ScanNamed(
					named.Required("id", &orderRow.Id),
					named.Required("user_id", &orderRow.UserId),
					named.Required("status", &orderRow.Status),
					named.Required("created_at", &orderRow.CreatedAt),
					named.Required("updated_at", &orderRow.UpdatedAt),
					named.Required("product_id", &orderRow.ProductId),
					named.Required("product_name", &orderRow.ProductName),
					named.Required("product_seller_id", &orderRow.ProductSellerId),
					named.Required("produt_count", &orderRow.ProductCount),
					named.Required("product_price", &orderRow.ProductPrice),
					named.Required("product_picture", &orderRow.ProductPicture),
				); err != nil {
					return err
				}
				ordersRows = append(ordersRows, orderRow)
			}
		}

		return res.Err()
	}); err != nil {
		return oapi_codegen.OrdersListOrdersRes{}, err
	}

	var lastOrderId *string
	if len(ordersRows) > 0 {
		lastOrderId = &ordersRows[len(ordersRows)-1].Id
	}
	orders := make(map[string]*oapi_codegen.OrdersListOrdersResOrder)
	for _, row := range ordersRows {
		_, ok := orders[row.Id]
		if !ok {
			orders[row.Id] = &oapi_codegen.OrdersListOrdersResOrder{
				Id:        row.Id,
				Status:    row.Status,
				UserId:    row.UserId,
				CreatedAt: ptr(row.CreatedAt.Format(time.RFC3339)),
				UpdatedAt: ptr(row.UpdatedAt.Format(time.RFC3339)),
			}
		}

		orders[row.Id].Items = append(orders[row.Id].Items, oapi_codegen.OrdersListOrdersResItem{
			ProductId:  row.ProductId,
			Name:       row.ProductName,
			Count:      int(row.ProductCount),
			Price:      row.ProductPrice,
			SellerId:   row.ProductSellerId,
			PictureUrl: row.ProductPicture,
		})
	}

	if len(orders) > int(ListOrdersPageSize) {
		newNextPage := &ListOrdersNextPageDto{
			OrderId:   *lastOrderId,
			CreatedAt: *orders[*lastOrderId].CreatedAt,
		}
		delete(orders, *lastOrderId)

		data, err := json.Marshal(&newNextPage)
		if err != nil {
			return oapi_codegen.OrdersListOrdersRes{}, fmt.Errorf("serialize next page: %v", err)
		}

		out.NextPageToken = ptr(string(data))
	}

	out.Orders = make([]oapi_codegen.OrdersListOrdersResOrder, 0, len(orders))
	for _, order := range orders {
		out.Orders = append(out.Orders, *order)
	}

	return out, nil
}

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

type CreateOrderDTOInput struct {
	OrderId   string
	UserId    string
	Status    string
	CreatedAt time.Time
	Products  []oapi_codegen.PrivateOrderProcessReservedProductsReqProduct
}
type CreateOrderDTOOutput struct {
	OrderId   string
	UserId    string
	Status    string
	CreatedAt time.Time
}

func (s *Orders) CreateOrder(ctx context.Context, in CreateOrderDTOInput) (CreateOrderDTOOutput, error) {
	var orderItems []types.Value
	for _, p := range in.Products {
		orderItems = append(orderItems, types.StructValue(
			types.StructFieldValue("product_id", types.StringValueFromString(p.Id)),
			types.StructFieldValue("seller_id", types.StringValueFromString(p.SellerId)),
			types.StructFieldValue("name", types.StringValueFromString(p.Name)),
			types.StructFieldValue("count", types.Uint32Value(uint32(p.Count))),
			types.StructFieldValue("price", types.DoubleValue(p.Price)),
			types.StructFieldValue("picture", types.NullableUTF8Value(p.Picture)),
		))
	}

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryCreateOrder, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(in.OrderId)),
			table.ValueParam("$user_id", types.UTF8Value(in.UserId)),
			table.ValueParam("$status", types.UTF8Value(in.Status)),
			table.ValueParam("$created_at", types.DatetimeValueFromTime(in.CreatedAt)),
			table.ValueParam("$updated_at", types.DatetimeValueFromTime(in.CreatedAt)),
			table.ValueParam("$order_items", types.ListValue(orderItems...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		return res.Err()
	}); err != nil {
		return CreateOrderDTOOutput{}, err
	}

	return CreateOrderDTOOutput{
		OrderId:   in.OrderId,
		UserId:    in.UserId,
		Status:    in.Status,
		CreatedAt: in.CreatedAt,
	}, nil
}

// queryCreateOrderManyTestData
/*
$orders = AsList(
  AsStruct(
    UNWRAP(CAST("foo-bar-baz-qux7" AS Utf8)) AS id,
    UNWRAP(CAST("acd559b2-def1-4b01-b501-c642e22dd7da" AS Utf8)) AS user_id,
    UNWRAP(CAST("created" AS Utf8)) as status,
    CurrentUtcDatetime() AS created_at,
    CurrentUtcDatetime() AS updated_at,
    AsList(
      AsStruct(
        UNWRAP(CAST("product-1" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 1 name" AS Utf8)) AS name,
        3u AS count,
        23.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
      AsStruct(
        UNWRAP(CAST("product-2" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 2 name" AS Utf8)) AS name,
        5u AS count,
        43.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
      AsStruct(
        UNWRAP(CAST("product-3" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 3 name" AS Utf8)) AS name,
        2u AS count,
        53.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
    ) AS order_items,
  ),
  AsStruct(    UNWRAP(CAST("foo-bar-baz-qux8" AS Utf8)) AS id,
    UNWRAP(CAST("acd559b2-def1-4b01-b501-c642e22dd7da" AS Utf8)) AS user_id,
    UNWRAP(CAST("created" AS Utf8)) as status,
    CurrentUtcDatetime() AS created_at,
    CurrentUtcDatetime() AS updated_at,
    AsList(
      AsStruct(
        UNWRAP(CAST("product-1" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 1 name" AS Utf8)) AS name,
        1u AS count,
        23.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
      AsStruct(
        UNWRAP(CAST("product-2" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 2 name" AS Utf8)) AS name,
        2u AS count,
        43.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
      AsStruct(
        UNWRAP(CAST("product-3" AS Utf8)) AS product_id,
        UNWRAP(CAST("seller-1" AS Utf8)) AS seller_id,
        UNWRAP(CAST("product 3 name" AS Utf8)) AS name,
        1u AS count,
        43.49 AS price,
        CAST(NULL AS Optional<Utf8>) AS picture,
      ),
    ) AS order_items,
  ),
);
*/

var queryCreateOrderMany = template.ReplaceAllPairs(`
DECLARE $orders AS List<Struct<
  id:Utf8,
  user_id:Utf8,
  status:Utf8,
  created_at:Datetime,
  updated_at:Datetime,
  order_items:List<Struct<
  	product_id:Utf8,
  	seller_id:Utf8,
  	name:Utf8,
  	count:Uint32,
  	price:Double,
  	picture:Optional<Utf8>,
  >>
>>;

INSERT INTO {{table.orders}} (id, user_id, status, created_at, updated_at)
SELECT
  id,
  user_id,
  status,
  created_at,
  updated_at
FROM AS_TABLE($orders);

INSERT INTO {{table.order_items}} (product_id, order_id, seller_id, name, count, price, picture)
SELECT 
  oi.product_id AS product_id,
  o.id AS order_id,
  oi.seller_id AS seller_id,
  oi.name AS name,
  oi.count AS count,
  oi.price AS price,
  oi.picture AS picture,
FROM AS_TABLE($orders) o
FLATTEN LIST BY order_items AS oi;
`,
	"{{table.orders}}",
	tableOrders,
	"{{table.order_items}}",
	tableOrderItems,
)

type CreateOrderManyDTOInput struct {
	Orders []CreateOrderManyDTOInputOrder
}
type CreateOrderManyDTOInputOrder struct {
	Id        string
	Status    string
	Details   *string
	OrderId   *string
	UpdatedAt time.Time
	Products  []oapi_codegen.PrivateOrderProcessReservedProductsReqProduct
}

type CreateOrderManyDTOOutput struct{}

func (s *Orders) CreateOrderMany(ctx context.Context, in CreateOrderManyDTOInput) (*CreateOrderManyDTOOutput, error) {
	var out *CreateOrderManyDTOOutput

	orders := make([]types.Value, 0, len(in.Orders))
	for _, order := range in.Orders {
		order_items := make([]types.Value, 0, len(order.Products))
		for _, product := range order.Products {
			order_items = append(order_items, types.StructValue(
				types.StructFieldValue("id", types.UTF8Value(order.Id)),
				types.StructFieldValue("status", types.UTF8Value(order.Status)),
				types.StructFieldValue("details", types.NullableUTF8Value(order.Details)),
				types.StructFieldValue("order_id", types.NullableUTF8Value(order.OrderId)),
				types.StructFieldValue("updated_at", types.TimestampValueFromTime(order.UpdatedAt)),
			))
		}

		orders = append(orders, types.StructValue(
			types.StructFieldValue("id", types.UTF8Value(order.Id)),
			types.StructFieldValue("status", types.UTF8Value(order.Status)),
			types.StructFieldValue("details", types.NullableUTF8Value(order.Details)),
			types.StructFieldValue("order_id", types.NullableUTF8Value(order.OrderId)),
			types.StructFieldValue("updated_at", types.TimestampValueFromTime(order.UpdatedAt)),
			types.StructFieldValue("order_items", types.ListItems(orders...)),
		))
	}

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryCreateOrderMany, table.NewQueryParameters(
			table.ValueParam("$orders", types.ListValue(orders...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryUpdateOrder = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $updated_at AS Timestamp;

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

func (s *Orders) UpdateOrder(ctx context.Context, orderId, orderStatus string) (*oapi_codegen.OrdersUpdateOrderRes, error) {
	var out *oapi_codegen.OrdersUpdateOrderRes

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryUpdateOrder, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(orderId)),
			table.ValueParam("$status", types.UTF8Value(orderStatus)),
			table.ValueParam("$updated_at", types.TimestampValueFromTime(time.Now())),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var id string
				out = &oapi_codegen.OrdersUpdateOrderRes{}
				if err := res.ScanNamed(
					named.Required("id", &id),
					named.Required("status", &out.Status),
					named.Required("updated_at", &out.UpdatedAt),
				); err != nil {
					return err
				}
				_ = id
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

// -- Existing orders only
// $orders = (
//   SELECT
//     o1.id AS id,
//     o1.user_id AS user_id,
//     o1.status AS status,
//     o1.created_at AS created_at,
//     o1.updated_at AS updated_at,
//     o1.order_items AS order_items,
//   FROM AS_TABLE($orders) o1
//   JOIN `orders/orders` o2 ON o1.id = o2.id
// );

func UpdateOrderMany() {

}

func ptr[T any](v T) *T {
	return &v
}
