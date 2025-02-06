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

// TODO: read from config
const (
	TableAccounts      = "accounts"
	TableRefreshTokens = "refresh_tokens"
)

type Account struct {
	db       *ydb.Driver
	l        *zap.Logger
	idHasher idhash.IdHasher
	ph       *auth.PasswordHasher
}

var _ domain.AccountProvider = (*Account)(nil)

type Conf struct {
	DbDriver       *ydb.Driver
	Logger         *zap.Logger
	IdHasher       idhash.IdHasher
	PasswordHasher *auth.PasswordHasher
}

func New(conf Conf) *Account {
	adapter := &Account{
		db:       conf.DbDriver,
		idHasher: conf.IdHasher,
		ph:       conf.PasswordHasher,
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
DECLARE $created_at AS Timestamp;
UPSERT INTO %s ( name, password, email, type, created_at )
VALUES ( $name, $password, $email, $type, $created_at )
RETURNING id, name, email, type
`, TableAccounts)

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
			table.ValueParam("$created_at", types.TimestampValueFromTime(time.Now())),
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
		// return res.Close()
	})
	if err != nil {
		return out, fmt.Errorf("failed to run create account ydb query: %v", err)
	}

	return out, nil
}

//	if ok := p.ph.Check(password, dbPassword); !ok {
//		return nil, domain.ErrInvalidCredentials
//	}
