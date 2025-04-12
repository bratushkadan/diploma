package store

import "github.com/bratushkadan/floral/pkg/template"

const (
	tablePayments = "`orders/payments`"
)

var queryCreatePayment = template.ReplaceAllPairs(`
`,
	"{{table.payments}}",
	tablePayments,
)

var queryGetPayment = template.ReplaceAllPairs(`
`,
	"{{table.payments}}",
	tablePayments,
)

var queryUpdatePayment = template.ReplaceAllPairs(`
`,
	"{{table.payments}}",
	tablePayments,
)
