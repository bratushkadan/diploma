package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bratushkadan/floral/pkg/template"
	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"go.uber.org/zap"
)

// TODO: read from config
const (
	tableProducts = "`products/products`"

	tableProductsIndexSellerId    = "idx_seller_id"
	tableProductsIndexCreatedAtId = "idx_created_at_id"
)

type Products struct {
	db *ydb.Driver
	l  *zap.Logger
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

	return &b.p, nil
}

var queryGetProduct = template.ReplaceAllPairs(`
DECLARE $id AS Uuid;

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
			table.ValueParam("$id", types.UuidValue(id)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var out GetProductDTOOutput
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
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
					return errors.New("failed to unmarshal product pictures json field")
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return errors.New("failed to unmarshal product metadata json field")
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
    id:Optional<Uuid>,
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

	opts = append(opts, types.StructFieldValue("id", types.NullableUUIDTypedValue(&in.Id)))
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
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
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
					return errors.New("failed to unmarshal product pictures json field")
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return errors.New("failed to unmarshal product metadata json field")
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
DECLARE $id AS Uuid;
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
			table.ValueParam("$id", types.UuidValue(in.Id)),
			table.ValueParam("$deleted_at", types.DatetimeValueFromTime(in.DeletedAt)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var outV DeleteProductDTOOutput
				if err := res.ScanNamed(
					named.Required("id", &outV.Id),
				); err != nil {
					return err
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
DECLARE $page_id AS Optional<Uuid>;
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
	tableParams = append(tableParams, table.ValueParam("$page_id", types.NullableUUIDTypedValue(nextPage.Id)))
	tableParams = append(tableParams, table.ValueParam("$page_size", types.Uint64Value(uint64(nextPage.PageSize))))

	var outProducts []ListProductsDTOOutputItem

	p.l.Info("query ydb")
	if err := p.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryListProducts, table.NewQueryParameters(tableParams...))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var out ListProductsDTOOutputItem
				var picturesJson, metadataJson []byte
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
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
					return errors.New("failed to unmarshal product pictures json field")
				}
				if err := json.Unmarshal(metadataJson, &out.Metadata); err != nil {
					return errors.New("failed to unmarshal product metadata json field")
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
