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

	tableProductsIndexSellerId = "idx_seller_id"
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
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("seller_id", &out.SellerId),
					named.Required("name", &out.Name),
					named.Required("description", &out.Description),
					named.Required("pictures", &out.Pictures),
					named.Required("metadata", &out.Metadata),
					named.Required("stock", &out.Stock),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("deleted_at", &out.DeletedAt),
				); err != nil {
					return err
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
	Id          *uuid.UUID
	SellerId    *string
	Name        *string
	Description *string
	Pictures    []UpsertProductDTOOutputPicture
	Metadata    map[string]any
	Stock       *uint32
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

	opts = append(opts, types.StructFieldValue("id", types.NullableUUIDTypedValue(in.Id)))
	opts = append(opts, types.StructFieldValue("seller_id", types.NullableUTF8Value(in.SellerId)))
	opts = append(opts, types.StructFieldValue("name", types.NullableUTF8Value(in.Name)))
	opts = append(opts, types.StructFieldValue("description", types.NullableUTF8Value(in.Description)))
	opts = append(opts, types.StructFieldValue("stock", types.NullableUint32Value(in.Stock)))
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
	}
	if in.Metadata != nil {
		metadataJson, err := json.Marshal(in.Metadata)
		if err != nil {
			return UpsertProductDTOOutput{}, err
		}
		metadata := string(metadataJson)
		opts = append(opts, types.StructFieldValue("metadata", types.NullableJSONValue(&metadata)))
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

var queryGetProductsBySellerId = template.ReplaceAllPairs(`
DECLARE $seller_id AS Utf8;

DECLARE $id AS Utf8;

SELECT 
    id,
    seller_id,
    name,
    description,
    pictures,
    metadata,
    stock
    created_at,
    updated_at,
    deleted_at
FROM
    {{table.tableProducts}}
VIEW
    {{index.seller_id}}
WHERE
    seller_id = $seller_id;
)
`,
	"{{table.tableProducts}}", tableProducts,
	"{{index.seller_id}}", tableProductsIndexSellerId,
)
