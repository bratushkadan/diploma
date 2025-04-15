package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/template"
	ydbtopic "github.com/bratushkadan/floral/pkg/ydb/topic"
	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"go.uber.org/zap"
)

// TODO: read from config
const (
	tableProducts = "`products/products`"

	topicProductsReservedProductsTopic = "products/reserved_products_topic"
	topicOrdersCancelOperations        = "orders/cancel_operations_topic"
	topicProductsUnreservedTopic       = "products/unreserved_products_topic"

	tableProductsIndexSellerId    = "idx_seller_id"
	tableProductsIndexCreatedAtId = "idx_created_at_id"
)

type Products struct {
	db *ydb.Driver
	l  *zap.Logger

	topicProductsReservedProductsTopic *topicwriter.Writer
	topicProductsUnreservedProducts    *topicwriter.Writer
	topicOrdersCancelOperations        *topicwriter.Writer
}

type ProductsBuilder struct {
	p Products
}

func NewProductsBuilder() *ProductsBuilder {
	return &ProductsBuilder{
		p: Products{},
	}
}

func (b *ProductsBuilder) YDBDriver(driver *ydb.Driver) *ProductsBuilder {
	b.p.db = driver
	return b
}
func (b *ProductsBuilder) Logger(l *zap.Logger) *ProductsBuilder {
	b.p.l = l
	return b
}

func (b *ProductsBuilder) Build() (*Products, error) {
	if b.p.l == nil {
		b.p.l = zap.NewNop()
	}

	if b.p.db == nil {
		return nil, errors.New("YDBDriver must be set")
	}

	topicProductsReservedProductsTopic, err := ydbtopic.NewProducer(b.p.db, topicProductsReservedProductsTopic)
	if err != nil {
		return nil, errors.New("setup ProductsReservedProductsTopic topic: %w")
	}
	b.p.topicProductsReservedProductsTopic = topicProductsReservedProductsTopic

	topicProductsUnreserved, err := ydbtopic.NewProducer(b.p.db, topicProductsUnreservedTopic)
	if err != nil {
		return nil, errors.New("setup ProductsUnreservedTopic topic: %w")
	}
	b.p.topicProductsUnreservedProducts = topicProductsUnreserved

	topicOrdersCancelOperations, err := ydbtopic.NewProducer(b.p.db, topicOrdersCancelOperations)
	if err != nil {
		return nil, errors.New("setup OrdersCancelOperations topic: %w")
	}
	b.p.topicOrdersCancelOperations = topicOrdersCancelOperations

	return &b.p, nil
}

var queryGetProduct = template.ReplaceAllPairs(`
DECLARE $id AS String;

SELECT 
    id,
    seller_id,
    name,
    description,
    pictures,
    metadata,
    stock,
    price,
    created_at,
    updated_at,
    deleted_at
FROM
    {{table.tableProducts}}
WHERE
    id = $id
        AND
    deleted_at IS NULL;
`,
	"{{table.tableProducts}}", tableProducts,
)

type GetProductDTOOutput struct {
	Id          uuid.UUID
	SellerId    string
	Name        string
	Description string
	Pictures    []GetProductDTOOutputPicture
	Metadata    map[string]any
	Stock       uint32
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
type GetProductDTOOutputPicture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

func (p *Products) Get(ctx context.Context, id uuid.UUID) (*GetProductDTOOutput, error) {
	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	var outProduct *GetProductDTOOutput

	if err := p.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryGetProduct, table.NewQueryParameters(
			table.ValueParam("$id", types.StringValueFromString(id.String())),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var out GetProductDTOOutput
				var strId string
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &strId),
					named.Required("seller_id", &out.SellerId),
					named.Required("name", &out.Name),
					named.Required("description", &out.Description),
					named.Required("pictures", &picturesJson),
					named.Required("metadata", &metadataJson),
					named.Required("stock", &out.Stock),
					named.Required("price", &out.Price),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("deleted_at", &out.DeletedAt),
				); err != nil {
					return err
				}
				if err := json.Unmarshal(picturesJson, &out.Pictures); err != nil {
					return fmt.Errorf("failed to unmarshal product pictures json field: %v", err)
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return fmt.Errorf("failed to unmarshal product metadata json field: %v", err)
				}
				out.Id, err = uuid.Parse(strId)
				if err != nil {
					return fmt.Errorf("failed to parse uuid from string id: %v", err)
				}
				outProduct = &out
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return outProduct, nil
}

var queryUpsertProduct = template.ReplaceAllPairs(`
DECLARE $s AS Struct<
    id:Optional<String>,
    seller_id:Optional<Utf8>,
    name:Optional<Utf8>,
    description:Optional<Utf8>,
    pictures:Optional<Json>,
    metadata:Optional<Json>,
    stock:Optional<Uint32>,
    price:Optional<Double>,
    created_at:Optional<Datetime>,
    updated_at:Optional<Datetime>,
    deleted_at:Optional<Datetime>,
>;

$existing = (
    SELECT
        id,
        seller_id,
        name,
        description,
        pictures,
        metadata,
        stock,
        price,
        created_at,
        updated_at,
        deleted_at,
    FROM
        {{table.tableProducts}}
    WHERE id = TryMember($s, "id", NULL)
);

$sub = (
    SELECT
        Unwrap(COALESCE(e.id, u.id)) AS id,
        Unwrap(COALESCE(u.seller_id, e.seller_id)) AS seller_id,
        Unwrap(COALESCE(u.name, e.name)) AS name,
        Unwrap(COALESCE(u.description, e.description)) AS description,
        Unwrap(COALESCE(u.pictures, e.pictures)) AS pictures,
        Unwrap(COALESCE(u.metadata, e.metadata)) AS metadata,
        Unwrap(COALESCE(u.stock, e.stock)) AS stock,
        Unwrap(COALESCE(u.price, e.price)) AS price,
        Unwrap(COALESCE(u.created_at, e.created_at)) AS created_at,
        Unwrap(COALESCE(u.updated_at, e.updated_at)) AS updated_at,
        COALESCE(u.deleted_at, e.deleted_at) AS deleted_at,
    FROM
        $existing e
    RIGHT JOIN AS_TABLE(AsList($s)) u ON u.id = e.id
);

UPSERT INTO {{table.tableProducts}}
SELECT * FROM $sub;

SELECT * FROM $sub;
`,
	"{{table.tableProducts}}", tableProducts,
)

type UpsertProductDTOInput struct {
	Id          uuid.UUID
	SellerId    *string
	Name        *string
	Description *string
	Pictures    []UpsertProductDTOOutputPicture
	Metadata    map[string]any
	Stock       *uint32
	Price       *float64
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}
type UpsertProductDTOOutput struct {
	Id          uuid.UUID
	SellerId    string
	Name        string
	Description string
	Pictures    []UpsertProductDTOOutputPicture
	Metadata    map[string]any
	Stock       uint32
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
type UpsertProductDTOOutputPicture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

func (p *Products) Upsert(ctx context.Context, in UpsertProductDTOInput) (UpsertProductDTOOutput, error) {
	var out UpsertProductDTOOutput

	var opts []types.StructValueOption

	strId := in.Id.String()
	opts = append(opts, types.StructFieldValue("id", types.NullableStringValueFromString(&strId)))
	opts = append(opts, types.StructFieldValue("seller_id", types.NullableUTF8Value(in.SellerId)))
	opts = append(opts, types.StructFieldValue("name", types.NullableUTF8Value(in.Name)))
	opts = append(opts, types.StructFieldValue("description", types.NullableUTF8Value(in.Description)))
	opts = append(opts, types.StructFieldValue("stock", types.NullableUint32Value(in.Stock)))
	opts = append(opts, types.StructFieldValue("price", types.NullableDoubleValue(in.Price)))
	opts = append(opts, types.StructFieldValue("created_at", types.NullableDatetimeValueFromTime(in.CreatedAt)))
	opts = append(opts, types.StructFieldValue("updated_at", types.NullableDatetimeValueFromTime(in.UpdatedAt)))
	opts = append(opts, types.StructFieldValue("deleted_at", types.NullableDatetimeValueFromTime(in.DeletedAt)))

	if in.Pictures != nil {
		picturesJson, err := json.Marshal(in.Pictures)
		if err != nil {
			return UpsertProductDTOOutput{}, err
		}
		pictures := string(picturesJson)
		opts = append(opts, types.StructFieldValue("pictures", types.NullableJSONValue(&pictures)))
	} else {
		opts = append(opts, types.StructFieldValue("pictures", types.NullableJSONValue(nil)))
	}
	if in.Metadata != nil {
		metadataJson, err := json.Marshal(in.Metadata)
		if err != nil {
			return UpsertProductDTOOutput{}, err
		}
		metadata := string(metadataJson)
		opts = append(opts, types.StructFieldValue("metadata", types.NullableJSONValue(&metadata)))
	} else {
		opts = append(opts, types.StructFieldValue("metadata", types.NullableJSONValue(nil)))
	}

	inYql := types.StructValue(opts...)

	if err := p.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryUpsertProduct, table.NewQueryParameters(
			table.ValueParam("$s", inYql),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var strId string
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &strId),
					named.Required("seller_id", &out.SellerId),
					named.Required("name", &out.Name),
					named.Required("description", &out.Description),
					named.Required("pictures", &picturesJson),
					named.Required("metadata", &metadataJson),
					named.Required("stock", &out.Stock),
					named.Required("price", &out.Price),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("deleted_at", &out.DeletedAt),
				); err != nil {
					return err
				}

				if err := json.Unmarshal(picturesJson, &out.Pictures); err != nil {
					return fmt.Errorf("failed to unmarshal product pictures json field: %v", err)
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return fmt.Errorf("failed to unmarshal product metadata json field: %v", err)
				}
				out.Id, err = uuid.Parse(strId)
				if err != nil {
					return errors.New("failed to parse uuid from string id")
				}
			}
		}

		return res.Err()
	}); err != nil {
		return UpsertProductDTOOutput{}, err
	}

	return out, nil
}

var queryDeleteProduct = template.ReplaceAllPairs(`
DECLARE $id AS String;
DECLARE $deleted_at AS Datetime;

$existing = (
    SELECT
        id,
        $deleted_at AS deleted_at
    FROM
        {{table.tableProducts}}
    WHERE 
        id = $id
            AND
        deleted_at IS NULL
);

UPDATE
    {{table.tableProducts}}
ON 
    SELECT * FROM $existing
RETURNING id;
`,
	"{{table.tableProducts}}", tableProducts,
)

type DeleteProductDTOInput struct {
	Id        uuid.UUID
	DeletedAt time.Time
}
type DeleteProductDTOOutput struct {
	Id uuid.UUID
}

func (p *Products) Delete(ctx context.Context, in DeleteProductDTOInput) (*DeleteProductDTOOutput, error) {
	var out *DeleteProductDTOOutput

	if err := p.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryDeleteProduct, table.NewQueryParameters(
			table.ValueParam("$id", types.StringValueFromString(in.Id.String())),
			table.ValueParam("$deleted_at", types.DatetimeValueFromTime(in.DeletedAt)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var outV DeleteProductDTOOutput
				var strId string
				if err := res.ScanNamed(
					named.Required("id", &strId),
				); err != nil {
					return err
				}
				outV.Id, err = uuid.Parse(strId)
				if err != nil {
					return errors.New("failed to parse uuid from string id")
				}
				out = &outV
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryListProducts = template.ReplaceAllPairs(`
DECLARE $seller_id AS Optional<Utf8>;
DECLARE $in_stock AS Optional<Bool>;

DECLARE $page_created_at AS Optional<Datetime>;
DECLARE $page_id AS Optional<String>;
DECLARE $page_size AS Uint64;

SELECT 
   ca.id          AS id, 
   ca.seller_id   AS seller_id,
   ca.name        AS name,
   ca.description AS description,
   ca.pictures    AS pictures,
   ca.metadata    AS metadata,
   ca.stock       AS stock,
   ca.price       AS price,
   ca.created_at  AS created_at,
   ca.updated_at  AS updated_at,
FROM 
  {{table.table_products}}
VIEW {{index.created_at_id}} ca
JOIN {{table.table_products}} 
VIEW {{index.seller_id}} s
ON ca.id = s.id
WHERE
    ($page_created_at IS NULL OR ca.created_at >= $page_created_at)
        AND
    ($page_id IS NULL OR ca.id >= $page_id)
        AND
    (s.seller_id = COALESCE($seller_id, s.seller_id))
        AND
    ca.deleted_at IS NULL
        AND
    ($in_stock IS NULL OR (ca.stock > 0 AND $in_stock) OR (ca.stock = 0 AND NOT $in_stock))
ORDER BY id, created_at
LIMIT MIN_OF($page_size, 25) + 1;
`,
	"{{table.table_products}}", tableProducts,
	"{{index.seller_id}}", tableProductsIndexSellerId,
	"{{index.created_at_id}}", tableProductsIndexCreatedAtId,
)

type ListProductsDTOOutputItem struct {
	Id          uuid.UUID
	SellerId    string
	Name        string
	Description string
	Pictures    []ListProductsDTOOutputPicture
	Metadata    map[string]any
	Stock       uint32
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type ListProductsDTOOutputPicture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type ListProductsNextPage struct {
	CreatedAt *time.Time `json:"created_at"`
	Id        *uuid.UUID `json:"id"`
	InStock   *bool      `json:"in_stock"`
	SellerId  *string    `json:"seller_id"`
	PageSize  int        `json:"page_size"`
}

func (p *Products) List(ctx context.Context, nextPage ListProductsNextPage) ([]ListProductsDTOOutputItem, error) {
	readTx := table.TxControl(table.BeginTx(table.WithStaleReadOnly()), table.CommitTx())

	tableParams := make([]table.ParameterOption, 0, 5)
	tableParams = append(tableParams, table.ValueParam("$seller_id", types.NullableUTF8Value(nextPage.SellerId)))
	tableParams = append(tableParams, table.ValueParam("$in_stock", types.NullableBoolValue(nextPage.InStock)))
	tableParams = append(tableParams, table.ValueParam("$page_created_at", types.NullableDatetimeValueFromTime(nextPage.CreatedAt)))
	if nextPage.Id != nil {
		strId := nextPage.Id.String()
		tableParams = append(tableParams, table.ValueParam("$page_id", types.NullableStringValueFromString(&strId)))
	} else {
		tableParams = append(tableParams, table.ValueParam("$page_id", types.NullableStringValueFromString(nil)))
	}
	tableParams = append(tableParams, table.ValueParam("$page_size", types.Uint64Value(uint64(nextPage.PageSize))))

	var outProducts []ListProductsDTOOutputItem

	if err := p.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryListProducts, table.NewQueryParameters(tableParams...))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var out ListProductsDTOOutputItem
				var strId string
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &strId),
					named.Required("seller_id", &out.SellerId),
					named.Required("name", &out.Name),
					named.Required("description", &out.Description),
					named.Required("pictures", &picturesJson),
					named.Required("metadata", &metadataJson),
					named.Required("stock", &out.Stock),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
				); err != nil {
					return err
				}
				if err := json.Unmarshal(picturesJson, &out.Pictures); err != nil {
					return fmt.Errorf("failed to unmarshal product pictures json field: %v", err)
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return fmt.Errorf("failed to unmarshal product metadata json field: %v", err)
				}
				out.Id, err = uuid.Parse(strId)
				if err != nil {
					return fmt.Errorf("failed to parse uuid from string id: %v", err)
				}
				outProducts = append(outProducts, out)
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return outProducts, nil
}

var queryListProductsForReservation = template.ReplaceAllPairs(`
DECLARE $product_ids AS List<String>;

SELECT 
  id,
  seller_id,
  name,
  stock,
  price,
  pictures
FROM {{table.table_products}}
WHERE 
    id IN $product_ids
        AND
    deleted_at IS NULL;
`,
	"{{table.table_products}}", tableProducts,
)
var queryReserveProducts = template.ReplaceAllPairs(`
DECLARE $updates AS List<Struct<
    id:String, 
    stock:Uint32
>>;

UPDATE {{table.table_products}} ON
SELECT
  update.id AS id,
  update.stock AS stock
FROM
  AS_TABLE($updates) AS update;
`,
	"{{table.table_products}}", tableProducts,
)

func (p *Products) ReserveProducts(ctx context.Context, messages []oapi_codegen.PrivateReserveProductsReqMessage) error {
	reserved := make([]oapi_codegen.PrivateOrderProcessReservedProductsReqMessage, 0)
	failedToReserve := make([]oapi_codegen.PrivateOrderCancelOperationsReqMessage, 0)

	productsToQuery := make(map[string]struct{}, 0)
	for _, m := range messages {
		for _, p := range m.Products {
			productsToQuery[p.Id] = struct{}{}
		}
	}

	if err := p.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		// 1. Read
		var productIdsList []types.Value
		for id := range productsToQuery {
			productIdsList = append(productIdsList, types.StringValueFromString(id))
		}
		res, err := tx.Execute(ctx, queryListProductsForReservation, table.NewQueryParameters(
			table.ValueParam("$product_ids", types.ListValue(productIdsList...)),
		))

		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		// count is "stock" here
		products := make(map[string]oapi_codegen.PrivateOrderProcessReservedProductsReqProduct)

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var product oapi_codegen.PrivateOrderProcessReservedProductsReqProduct
				var stock uint32
				var picturesJson []byte
				if err := res.ScanNamed(
					named.Required("id", &product.Id),
					named.Required("seller_id", &product.SellerId),
					named.Required("name", &product.Name),
					named.Required("stock", &stock),
					named.Required("price", &product.Price),
					named.Required("pictures", &picturesJson),
				); err != nil {
					return err
				}

				var pictures []ListProductsDTOOutputPicture
				if err := json.Unmarshal(picturesJson, &pictures); err != nil {
					return fmt.Errorf("failed to unmarshal product pictures json field: %v", err)
				}

				if len(pictures) > 0 {
					picture := pictures[0]
					product.Picture = &picture.Url
				}

				product.Count = int(stock)
				products[product.Id] = product
			}
		}

		if err := res.Err(); err != nil {
			return err
		}

		// 2. Compute
		for _, msg := range messages {
			ok := true
			reservedPositions := make(map[string]int)
			var detailsMessages []string
			for _, p := range msg.Products {
				product, exists := products[p.Id]
				if !exists {
					detailsMessages = append(detailsMessages, fmt.Sprintf(`product id="%s" does not exist`, p.Id))
					ok = false
					continue
				}
				if product.Count < p.Count {
					detailsMessages = append(detailsMessages, fmt.Sprintf(`product id="%s" stock (%d) is less than requested (%d)`, p.Id, product.Count, p.Count))
					ok = false
				} else if ok {
					reservedPositions[p.Id] = p.Count
				}
			}
			if !ok {
				failedToReserve = append(failedToReserve, oapi_codegen.PrivateOrderCancelOperationsReqMessage{
					OperationId: msg.OperationId,
					Details:     strings.Join(detailsMessages, ", "),
				})
				continue
			}

			reservedPositionsRes := make([]oapi_codegen.PrivateOrderProcessReservedProductsReqProduct, 0, len(reservedPositions))
			for productId, count := range reservedPositions {
				product := products[productId]
				product.Count -= count
				products[productId] = product

				reservedPosition := products[productId]
				reservedPosition.Count = count
				reservedPositionsRes = append(reservedPositionsRes, reservedPosition)
			}

			reserved = append(reserved, oapi_codegen.PrivateOrderProcessReservedProductsReqMessage{
				OperationId: msg.OperationId,
				Products:    reservedPositionsRes,
			})
		}

		p.l.Info("to reserve", zap.Any("products", reserved))
		p.l.Info("to report reservation failure", zap.Any("products", failedToReserve))

		// 3. Write
		var updates []types.Value
		for _, p := range products {
			updates = append(updates, types.StructValue(
				types.StructFieldValue("id", types.StringValueFromString(p.Id)),
				types.StructFieldValue("stock", types.Uint32Value(uint32(p.Count))),
			))
		}

		res, err = tx.Execute(ctx, queryReserveProducts, table.NewQueryParameters(
			table.ValueParam("$updates", types.ListValue(updates...)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		// 4. Publish reserved
		productsReservedData := make([][]byte, 0, len(reserved))
		for _, r := range reserved {
			data, err := json.Marshal(&r)
			if err != nil {
				return fmt.Errorf("marshal reserved products: %v", err)
			}
			productsReservedData = append(productsReservedData, data)
		}

		if err := ydbtopic.Produce(ctx, p.topicProductsReservedProductsTopic, productsReservedData...); err != nil {
			return fmt.Errorf("produce products reserved messages: %v", err)
		}

		// 5. Publish reserve failures
		orderCancelOperationsData := make([][]byte, 0, len(failedToReserve))
		for _, r := range failedToReserve {
			data, err := json.Marshal(&r)
			if err != nil {
				return fmt.Errorf("marshal order cancel operations: %v", err)
			}
			orderCancelOperationsData = append(orderCancelOperationsData, data)
		}
		if err := ydbtopic.Produce(ctx, p.topicOrdersCancelOperations, orderCancelOperationsData...); err != nil {
			return fmt.Errorf("produce orders cancel operations messages: %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

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
        (p.stock + u.stock) AS stock,
        CurrentUtcDatetime() AS updated_at,
    FROM {{table.table_products}} p
    JOIN AS_TABLE($updates) u ON p.id = u.id
)
RETURNING id, stock, updated_at;
`,
	"{{table.table_products}}", tableProducts,
)

func (p *Products) UnreserveProducts(ctx context.Context, messages []oapi_codegen.PrivateUnreserveProductsReqMessage) error {
	unreserveProductsMessages := make([]oapi_codegen.PrivateOrderProcessUnreservedProductsReqMessage, 0, len(messages))

	if err := p.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		// 1. Compute
		toUnreserve := make(map[string]uint32)
		for _, msg := range messages {
			unreserveProductsMessages = append(unreserveProductsMessages, oapi_codegen.PrivateOrderProcessUnreservedProductsReqMessage{
				OrderId: msg.OrderId,
			})

			for _, product := range msg.Products {
				count, _ := toUnreserve[product.Id]
				toUnreserve[product.Id] = count + uint32(product.Count)
			}
		}

		// 2. Write
		updates := make([]types.Value, 0, len(toUnreserve))
		for productId, count := range toUnreserve {
			updates = append(updates, types.StructValue(
				types.StructFieldValue("id", types.StringValueFromString(productId)),
				types.StructFieldValue("stock", types.Uint32Value(count)),
			))
		}

		res, err := tx.Execute(ctx, queryUnreserveProducts, table.NewQueryParameters(
			table.ValueParam("$updates", types.ListValue(updates...)),
		))

		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		results := make([]any, 0)

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var row struct {
					Id        string    `json:"id"`
					Stock     uint32    `json:"stock"`
					UpdatedAt time.Time `json:"updated_at"`
				}
				if err := res.ScanNamed(
					named.Required("id", &row.Id),
					named.Required("stock", &row.Stock),
					named.Required("updated_at", &row.UpdatedAt),
				); err != nil {
					return err
				}
				results = append(results, row)
			}
		}

		if err := res.Err(); err != nil {
			return err
		}

		// Publish
		unreserveProductsByteMessages := make([][]byte, 0, len(messages))
		for _, msg := range unreserveProductsMessages {
			data, err := json.Marshal(&msg)
			if err != nil {
				return fmt.Errorf("serialize unreserved product message: %v", err)
			}
			unreserveProductsByteMessages = append(unreserveProductsByteMessages, data)
		}
		if err := ydbtopic.Produce(ctx, p.topicProductsUnreservedProducts, unreserveProductsByteMessages...); err != nil {
			return fmt.Errorf("publish unreserved products messages: %v", err)
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}
