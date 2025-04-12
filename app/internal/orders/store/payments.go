package store

import "github.com/bratushkadan/floral/pkg/template"

const (
	tablePayments = "`orders/payments`"
)

var queryCreatePayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $order_id AS Utf8;
DECLARE $provider AS Json;
DECLARE $created_at AS Datetime;
DECLARE $updated_at AS Datetime;
DECLARE $refunded_at AS Optional<Datetime>;

-- $id = UNWRAP(CAST("op1" AS Utf8));
-- $order_id = UNWRAP(CAST("" AS Utf8));
-- $created_at = CurrentUtcTimestamp();
-- $updated_at = CurrentUtcTimestamp();
-- $provider = @@{"name": "yoomoney"}@@j;

INSERT INTO {{table.payments}} (id, order_id, provider, created_at, updated_at, refunded_at)
VALUES($id, $order_id, $provider, $created_at, $updated_at, $refunded_at)
RETURNING *;
`,
	"{{table.payments}}",
	tablePayments,
)

var queryGetPayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;

$id = UNWRAP(CAST("op1" AS Utf8));

SELECT
  id,
  order_id,
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

var queryUpdatePayment = template.ReplaceAllPairs(`
DECLARE $id AS Utf8;
DECLARE $refunded_at AS Optional<Datetime>;

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
