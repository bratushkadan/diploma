package ydb_adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/resource/idhash"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"go.uber.org/zap"
)

type Account struct {
	db       *ydb.Driver
	l        *zap.Logger
	idHasher idhash.IdHasher
	ph       *auth.PasswordHasher
}

var _ domain.AccountProvider = (*Account)(nil)

type AccountConf struct {
	DbDriver       *ydb.Driver
	Logger         *zap.Logger
	IdHasher       idhash.IdHasher
	PasswordHasher *auth.PasswordHasher
}

func NewAccount(conf AccountConf) *Account {
	adapter := &Account{
		db:       conf.DbDriver,
		idHasher: conf.IdHasher,
		ph:       conf.PasswordHasher,
		l:        conf.Logger,
	}

	if conf.Logger == nil {
		adapter.l = zap.NewNop()
	}

	return adapter
}

var queryCreateAccount = fmt.Sprintf(`
DECLARE $name AS Utf8;
DECLARE $password AS Utf8;
DECLARE $email AS Utf8;
DECLARE $type AS String;
DECLARE $created_at AS Datetime;
UPSERT INTO %s ( name, password, email, type, created_at )
VALUES ( $name, $password, $email, $type, $created_at )
RETURNING id, name, email, type
`, tableAccounts)

func (a *Account) CreateAccount(ctx context.Context, in domain.CreateAccountDTOInput) (domain.CreateAccountDTOOutput, error) {
	var out domain.CreateAccountDTOOutput

	hashedPass, err := a.ph.Hash(in.Password)
	if err != nil {
		return domain.CreateAccountDTOOutput{}, fmt.Errorf("failed to hash account password: %v", err)
	}

	err = a.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		res, err := tx.Execute(ctx, queryCreateAccount, table.NewQueryParameters(
			table.ValueParam("$name", types.UTF8Value(in.Name)),
			table.ValueParam("$email", types.UTF8Value(in.Email)),
			table.ValueParam("$password", types.UTF8Value(hashedPass)),
			table.ValueParam("$type", types.StringValueFromString(in.Type)),
			table.ValueParam("$created_at", types.DatetimeValueFromTime(time.Now())),
		))
		if err != nil {
			if ydb.IsOperationError(err, Ydb.StatusIds_PRECONDITION_FAILED) {
				return fmt.Errorf("%w: %w", domain.ErrEmailIsInUse, err)
			}
			return err
		}
		if err := res.Err(); err != nil {
			return err
		}
		defer func() {
			if err := res.Close(); err != nil {
				a.l.Error("failed to close ydb result", zap.Error(err))
			}
		}()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var id int64
				if err := res.ScanNamed(
					named.Required("id", &id),
					named.Required("name", &out.Name),
					named.Required("email", &out.Email),
					named.Required("type", &out.Type),
				); err != nil {
					return err
				}

				idStr, err := a.idHasher.EncodeInt64(id)
				if err != nil {
					return fmt.Errorf("failed to hash encode int64 id %d: %v", id, err)
				}

				out.Id = idStr
			}
		}

		return nil
		// return res.Close() // <---- If I do not require RETURNING values when executing query
	})
	if err != nil {
		return out, fmt.Errorf("failed to run create account ydb query: %v", err)
	}

	return out, nil
}

var queryFindAccount = fmt.Sprintf(`
DECLARE $id AS Int64;
SELECT
  name,
  email,
  type,
  (activated_at IS NOT NULL) AS activated
FROM
  %s
WHERE
  id = $id;
`, tableAccounts)

func (a *Account) FindAccount(ctx context.Context, in domain.FindAccountDTOInput) (*domain.FindAccountDTOOutput, error) {
	intId, err := a.idHasher.DecodeInt64(in.Id)
	if err != nil {
		return nil, err
	}

	var out *domain.FindAccountDTOOutput

	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	err = a.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryFindAccount, table.NewQueryParameters(
			table.ValueParam("$id", types.Int64Value(intId)),
		))
		if err != nil {
			return err
		}
		defer res.Close()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var account domain.FindAccountDTOOutput
				if err := res.ScanNamed(
					named.Required("name", &account.Name),
					named.Required("email", &account.Email),
					named.Required("type", &account.Type),
					named.Required("activated", &account.Activated),
				); err != nil {
					return err
				}
				out = &account
			}
		}

		return res.Err()
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

var queryFindAccountByEmail = fmt.Sprintf(`
DECLARE $email AS Utf8;
SELECT
  id,
  name,
  type,
  (activated_at IS NOT NULL) AS activated
FROM
  %s
VIEW
  %s 
WHERE
  email = $email;
`, tableAccounts, tableAccountsIndexEmailUnique)

func (a *Account) FindAccountByEmail(ctx context.Context, in domain.FindAccountByEmailDTOInput) (*domain.FindAccountByEmailDTOOutput, error) {
	var out *domain.FindAccountByEmailDTOOutput

	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	if err := a.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryFindAccountByEmail, table.NewQueryParameters(
			table.ValueParam("$email", types.UTF8Value(in.Email)),
		))
		if err != nil {
			return err
		}
		defer res.Close()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var intId int64
				var account domain.FindAccountByEmailDTOOutput
				if err := res.ScanNamed(
					named.Required("id", &intId),
					named.Required("name", &account.Name),
					named.Required("type", &account.Type),
					named.Required("activated", &account.Activated),
				); err != nil {
					return err
				}

				if id, err := a.idHasher.EncodeInt64(intId); err != nil {
					return err
				} else {
					account.Id = id
				}
				out = &account
			}
		}

		return res.Err()
	}); err != nil {
		return nil, err
	}

	return out, nil
}

var queryCheckAccountCredentials = fmt.Sprintf(`
DECLARE $email AS Utf8;

SELECT
  id,
  password,
  (activated_at IS NOT NULL) AS activated
FROM
  %s
VIEW
  %s 
WHERE
  email = $email;
`, tableAccounts, tableAccountsIndexEmailUnique)

func (a *Account) CheckAccountCredentials(ctx context.Context, in domain.CheckAccountCredentialsDTOInput) (domain.CheckAccountCredentialsDTOOutput, error) {
	var out domain.CheckAccountCredentialsDTOOutput

	readTx := table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())

	if err := a.db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, res, err := s.Execute(ctx, readTx, queryCheckAccountCredentials, table.NewQueryParameters(
			table.ValueParam("$email", types.UTF8Value(in.Email)),
		))
		if err != nil {
			return err
		}
		defer res.Close()

		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var password string
				var intId int64
				if err := res.ScanNamed(
					named.Required("id", &intId),
					named.Required("password", &password),
					named.Required("activated", &out.Activated),
				); err != nil {
					return err
				}

				id, err := a.idHasher.EncodeInt64(intId)
				if err != nil {
					return fmt.Errorf("failed to encode account int id: %w", err)
				}
				out.AccountId = id

				isPasswordMatch := a.ph.Check(in.Password, password)
				out.Ok = isPasswordMatch
			}
		}

		return res.Err()
	}); err != nil {
		return out, err
	}

	return out, nil
}

var queryActivateAccountsByEmail = fmt.Sprintf(`
DECLARE $activated_at AS Datetime;
DECLARE $emails AS List<Utf8>;

-- 1. (TableRangeScan)

$to_update = (
    SELECT
      id,
      COALESCE(activated_at, $activated_at) AS activated_at
    FROM
      %s
    VIEW
      %s
    WHERE
      email IN $emails
);

UPDATE
  %s
ON
  SELECT * FROM $to_update;

-- 2. (TableFullScan)

-- UPDATE
--   %s
-- SET
--   activated_at = $activated_at
-- WHERE
--   email IN $emails;
`, tableAccounts, tableAccountsIndexEmailUnique, tableAccounts, tableAccounts)

func (a *Account) ActivateAccountsByEmail(ctx context.Context, in domain.ActivateAccountsByEmailDTOInput) error {
	if err := a.db.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		emailValues := make([]types.Value, 0, len(in.Emails))
		for _, v := range in.Emails {
			emailValues = append(emailValues, types.UTF8Value(v))
		}

		res, err := tx.Execute(ctx, queryActivateAccountsByEmail, table.NewQueryParameters(
			table.ValueParam("$activated_at", types.DatetimeValueFromTime(time.Now())),
			table.ValueParam("$emails", types.ListValue(emailValues...)),
		))
		if err != nil {
			return err
		}
		if err := res.Close(); err != nil {
			return err
		}

		return res.Err()
	}); err != nil {
		return err
	}

	return nil
}
