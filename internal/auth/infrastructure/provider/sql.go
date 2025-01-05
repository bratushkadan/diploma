package provider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/bratushkadan/floral/internal/auth/domain"
	"github.com/bratushkadan/floral/pkg/auth"
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
	ph   *auth.PasswordHasher
	conf *postgres.DBConf
	db   *sql.DB
}

var _ domain.UserProvider = (*PostgresUserProvider)(nil)

func NewPostgresUserProvider(conf *postgres.DBConf, db *sql.DB, ph *auth.PasswordHasher) *PostgresUserProvider {
	return &PostgresUserProvider{
		ph:   ph,
		conf: conf,
		db:   db,
	}
}

func (p *PostgresUserProvider) CreateUser(ctx context.Context, req domain.UserProviderCreateUserReq) (*domain.User, error) {
	id, err := UserIdToInt64(GenerateUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to generate id for user creation: %v", err)
	}

	hashedPassword, err := p.ph.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (id, name, password, email, type) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, type`, AuthServiceSchema, TableUsers)
	row := p.db.QueryRowContext(ctx, smt, id, req.Name, hashedPassword, req.Email, req.Type)

	var userId int64
	var user domain.User
	if err := row.Scan(&userId, &user.Name, &user.Type); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	user.Id = Int64ToUserId(userId)

	return &user, nil
}

func (p *PostgresUserProvider) FindUser(ctx context.Context, id string) (*domain.User, error) {
	return nil, errors.New("unimplemented")
}
func (p *PostgresUserProvider) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, errors.New("unimplemented")
}
func (p *PostgresUserProvider) CheckUserCredentials(ctx context.Context, email string, password string) (*domain.User, error) {
	smt := fmt.Sprintf(`SELECT id, name, password, type FROM "%s"."%s" WHERE "email" = $1`, AuthServiceSchema, TableUsers)

	row := p.db.QueryRowContext(ctx, smt, email)
	var intId int64
	var name, dbPassword, userType string
	if err := row.Scan(&intId, &name, &dbPassword, &userType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user password from db: %v", err)
	}

	if ok := p.ph.Check(password, dbPassword); !ok {
		return nil, domain.ErrInvalidCredentials
	}

	return &domain.User{
		Id:   Int64ToUserId(intId),
		Name: name,
		Type: userType,
	}, nil
}

type PostgresRefreshTokenPersisterProvider struct {
	conf *postgres.DBConf
	db   *sql.DB
}

var _ domain.RefreshTokenPersisterProvider = (*PostgresRefreshTokenPersisterProvider)(nil)

func NewPostgresRefreshTokenPersisterProvider(conf *postgres.DBConf, db *sql.DB) *PostgresRefreshTokenPersisterProvider {
	return &PostgresRefreshTokenPersisterProvider{
		conf: conf,
		db:   db,
	}
}

func (p *PostgresRefreshTokenPersisterProvider) Get(ctx context.Context, subjectId string) ([]string, error) {
	id, err := UserIdToInt64(subjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate id for user creation: %v", err)
	}

	smt := fmt.Sprintf(`SELECT id FROM "%s"."%s" WHERE "user_id" = $1 AND CURRENT_TIMESTAMP <= "expires_at"`, AuthServiceSchema, TableRefreshTokens)

	rows, err := p.db.QueryContext(ctx, smt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens from the database: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Print(err)
		}
	}()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning refresh token id: %v", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}
func (p *PostgresRefreshTokenPersisterProvider) Add(ctx context.Context, token *domain.RefreshToken) error {
	intId, err := UserIdToInt64(token.SubjectId)
	if err != nil {
		return err
	}

	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (id, user_id, expires_at) VALUES ($1, $2, $3)`, AuthServiceSchema, TableRefreshTokens)

	row := p.db.QueryRowContext(ctx, smt, token.TokenId, intId, token.ExpiresAt)
	if err := row.Err(); err != nil {
		return fmt.Errorf("failed to insert token into the database: %w", err)
	}

	return nil
}
func (p *PostgresRefreshTokenPersisterProvider) Delete(ctx context.Context, tokenId string) error {
	smt := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE "id" = $1`, AuthServiceSchema, TableRefreshTokens)

	row := p.db.QueryRowContext(ctx, smt, tokenId)
	if err := row.Err(); err != nil {
		return fmt.Errorf("failed to delete token from the database: %w", err)
	}

	return nil
}
