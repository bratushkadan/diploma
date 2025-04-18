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
	tablePayments = "`orders/payments`"

	topicProcessedPaymentsNotifications = "orders/payment_notifications_topic"
)

var queryCreatePayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $order_id AS Utf8;
DECLARE $amount AS Double;
DECLARE $currency_iso_4217 AS Uint32;
DECLARE $provider AS Json;
DECLARE $created_at AS Timestamp;
DECLARE $updated_at AS Timestamp;
DECLARE $refunded_at AS Optional<Timestamp>;

-- $id = UNWRAP(CAST("op1" AS Utf8));
-- $order_id = UNWRAP(CAST("" AS Utf8));
-- $amount = 13.33;
-- $currency_iso_4217 = 643u;
-- $created_at = CurrentUtcTimestamp();
-- $updated_at = CurrentUtcTimestamp();
-- $provider = @@{"name": "yoomoney"}@@j;

INSERT INTO {{table.payments}} (id, order_id, amount, currency_iso_4217, provider, created_at, updated_at, refunded_at)
VALUES($id, $order_id, $amount, $currency_iso_4217, $provider, $created_at, $updated_at, $refunded_at)
RETURNING id, order_id, amount, currency_iso_4217, provider, created_at, updated_at, refunded_at;
`,
	"{{table.payments}}",
	tablePayments,
)

type CreatePaymentDTOInput struct {
	Id              string
	OrderId         string
	Amount          float64
	CurrencyIso4217 uint32
	Provider        map[string]any
	CreatedAt       time.Time
	RefundedAt      *time.Time
}
type CreatePaymentDTOOutput struct {
	Id              string
	OrderId         string
	Amount          float64
	CurrencyIso4217 uint32
	Provider        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	RefundedAt      *time.Time
}

func (s *Orders) CreatePayment(ctx context.Context, in CreatePaymentDTOInput) (CreatePaymentDTOOutput, error) {
	provider, err := json.Marshal(&in.Provider)
	if err != nil {
		return CreatePaymentDTOOutput{}, fmt.Errorf("serialize provider data: %v", err)
	}

	var out CreatePaymentDTOOutput

	tableQueryParameters := table.NewQueryParameters(
		table.ValueParam("$id", types.UTF8Value(in.Id)),
		table.ValueParam("$order_id", types.UTF8Value(in.Id)),
		table.ValueParam("$amount", types.DoubleValue(in.Amount)),
		table.ValueParam("$currency_iso_4217", types.Uint32Value(in.CurrencyIso4217)),
		table.ValueParam("$provider", types.JSONValueFromBytes(provider)),
		table.ValueParam("$created_at", types.TimestampValueFromTime(in.CreatedAt)),
		table.ValueParam("$updated_at", types.TimestampValueFromTime(in.CreatedAt)),
		table.ValueParam("$refunded_at", types.NullableTimestampValueFromTime(in.RefundedAt)),
	)

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryCreatePayment, tableQueryParameters)
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var providerJsonData []byte
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("order_id", &out.OrderId),
					named.Required("amount", &out.Amount),
					named.Required("currency_iso_4217", &out.CurrencyIso4217),
					named.Required("provider", &providerJsonData),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("refunded_at", &out.RefundedAt),
				); err != nil {
					return err
				}

				if err := json.Unmarshal(providerJsonData, &out.Provider); err != nil {
					return fmt.Errorf("deserialize payment provider data from database: %v", err)
				}
			}
		}

		return res.Err()
	}); err != nil {
		return CreatePaymentDTOOutput{}, err
	}

	return out, nil
}

var queryGetPayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;

SELECT
  id,
  order_id,
	amount,
	currency_iso_4217,
  provider,
  created_at,
  updated_at,
  refunded_at,
FROM
 {{table.payments}};
`,
	"{{table.payments}}",
	tablePayments,
)

type GetPaymentDTOInput struct {
	Id string
}
type GetPaymentDTOOutput struct {
	Id              string
	OrderId         string
	Amount          float64
	CurrencyIso4217 uint32
	Provider        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	RefundedAt      *time.Time
}

func (s *Orders) GetPayment(ctx context.Context, in GetPaymentDTOInput) (*GetPaymentDTOOutput, error) {
	var out *GetPaymentDTOOutput

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryGetPayment, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(in.Id)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("order_id", &out.OrderId),
					named.Required("amount", &out.Amount),
					named.Required("currency_iso_4217", &out.CurrencyIso4217),
					named.Required("provider", &out.Provider),
					named.Required("created_at", &out.CreatedAt),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("refunded_at", &out.RefundedAt),
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

var queryUpdatePayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $refunded_at AS Optional<Timestamp>;

$to_update = (
  SELECT
    $id AS id,
    $refunded_at AS refunded_at,
);

UPDATE {{table.payments}} ON
SELECT * FROM $to_update
RETURNING id, order_id, updated_at, refunded_at;
`,
	"{{table.payments}}",
	tablePayments,
)

type UpdatePaymentDTOInput struct {
	Id         string
	RefundedAt *time.Time
}
type UpdatePaymentDTOOutput struct {
	Id         string
	OrderId    string
	UpdatedAt  time.Time
	RefundedAt *time.Time
}

func (s *Orders) UpdatePayment(ctx context.Context, in UpdatePaymentDTOInput) (*UpdatePaymentDTOOutput, error) {
	var out *UpdatePaymentDTOOutput

	if err := s.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryUpdatePayment, table.NewQueryParameters(
			table.ValueParam("$id", types.UTF8Value(in.Id)),
			table.ValueParam("$refunded_at", types.NullableTimestampValueFromTime(in.RefundedAt)),
		))
		if err != nil {
			return err
		}
		defer func() { _ = res.Close() }()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				out = &UpdatePaymentDTOOutput{}
				if err := res.ScanNamed(
					named.Required("id", &out.Id),
					named.Required("order_id", &out.OrderId),
					named.Required("updated_at", &out.UpdatedAt),
					named.Optional("refunded_at", &out.RefundedAt),
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

func (s *Orders) ProduceProcessedPaymentsNotificationsMessages(ctx context.Context, messages ...oapi_codegen.PrivateOrderProcessPaymentReqMessage) error {
	dataBytes := make([][]byte, 0, len(messages))
	for _, message := range messages {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("serialize processed payment notification message: %v", err)
		}
		dataBytes = append(dataBytes, msgBytes)
	}

	if err := ydbtopic.Produce(ctx, s.topicProcessedPaymentsNotifications, dataBytes...); err != nil {
		return fmt.Errorf("publish processed payment notificaiton message: %v", err)
	}
	return nil
}
