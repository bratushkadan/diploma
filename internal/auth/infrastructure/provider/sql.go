package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bratushkadan/floral/internal/auth/domain"
	"github.com/bratushkadan/floral/pkg/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	AuthServiceSchema  = "auth"
	TableUsers         = "users"
	TableRefreshTokens = "refresh_tokens"
)

func NewDbconnPool(conf *postgres.DBConf) (*sql.DB, error) {
	db, err := postgres.NewDB(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database client: %w", err)
	}

	return db, nil
}

type PostgresUserProvider struct {
	conf *postgres.DBConf
	db   *sql.DB
}

func NewPostgresUserProvider(conf *postgres.DBConf, db *sql.DB) *PostgresUserProvider {
	return &PostgresUserProvider{
		conf: conf,
		db:   db,
	}
}

type PostgresRefreshTokenPersisterProvider struct {
	conf *postgres.DBConf
	db   *sql.DB
}

func NewPostgresRefreshTokenPersisterProvider(conf *postgres.DBConf, db *sql.DB) *PostgresRefreshTokenPersisterProvider {
	return &PostgresRefreshTokenPersisterProvider{
		conf: conf,
		db:   db,
	}
}

func (p *PostgresUserProvider) CreateUser(ctx context.Context, req domain.UserProviderCreateUserReq) (*domain.User, error) {
	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (id, name, password, email, type) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, type`, AuthServiceSchema, TableUsers)
	row := p.db.QueryRowContext(ctx, smt, req.Id, req.Name, req.Password, req.Email, req.Type)

	var user domain.User
	if err := row.Scan(&user.Id, &user.Name, &user.Type); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user, nil
}
