package provider

//
// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"
//
// 	"github.com/bratushkadan/floral/internal/auth/core/domain"
// 	"github.com/bratushkadan/floral/pkg/auth"
// 	"github.com/bratushkadan/floral/pkg/postgres"
// 	"github.com/bratushkadan/floral/pkg/resource"
// 	_ "github.com/jackc/pgx/v5/stdlib"
// )
//
// var (
// 	AuthServiceSchema    = "auth"
// 	TableUsers           = "users"
// 	TableRefreshTokens   = "refresh_tokens"
// 	TableConfirmationIds = "confirmation_ids"
// )
//
// func NewDbconnPool(conf *postgres.DBConf) (*sql.DB, error) {
// 	db, err := postgres.NewDB(conf)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize database client: %w", err)
// 	}
//
// 	return db, nil
// }
//
// type ConfirmationIdsOpts struct {
// 	// The duration confirmation id will be valid for.
// 	ExpiresAfter time.Duration
// }
//
// type PostgresUserProviderConf struct {
// 	Db               *sql.DB
// 	DbConf           *postgres.DBConf
// 	PasswordHasher   *auth.PasswordHasher
// 	ConfirmationOpts ConfirmationIdsOpts
// }
//
// type PostgresUserProvider struct {
// 	db               *sql.DB
// 	conf             *postgres.DBConf
// 	ph               *auth.PasswordHasher
// 	confirmationOpts ConfirmationIdsOpts
// }
//
// var _ domain.UserProvider = (*PostgresUserProvider)(nil)
//
// func NewPostgresUserProvider(conf PostgresUserProviderConf) *PostgresUserProvider {
// 	return &PostgresUserProvider{
// 		db:   conf.Db,
// 		conf: conf.DbConf,
// 		ph:   conf.PasswordHasher,
// 		// FIXME: only id should be returned by the db adapter
// 		confirmationOpts: conf.ConfirmationOpts,
// 	}
// }
//
// func (p *PostgresUserProvider) CreateUser(ctx context.Context, req domain.UserProviderCreateUserReq) (*domain.User, error) {
// 	id, err := UserIdToInt64(GenerateUserId())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate id for user creation: %v", err)
// 	}
//
// 	hashedPassword, err := p.ph.Hash(req.Password)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to hash password: %w", err)
// 	}
//
// 	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (id, name, password, email, type) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, type`, AuthServiceSchema, TableUsers)
// 	row := p.db.QueryRowContext(ctx, smt, id, req.Name, hashedPassword, req.Email, req.Type)
//
// 	var userId int64
// 	var user domain.User
// 	if err := row.Scan(&userId, &user.Name, &user.Type); err != nil {
// 		if postgres.IsUniqueConstraintViolation(err) {
// 			return nil, fmt.Errorf("%w: %w", domain.ErrEmailIsInUse, err)
// 		}
// 		return nil, fmt.Errorf("failed to create user: %v", err)
// 	}
// 	user.Id = Int64ToUserId(userId)
//
// 	return &user, nil
// }
//
// func (p *PostgresUserProvider) AddEmailConfirmationId(ctx context.Context, email string) (string, error) {
// 	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (email, id, expires_at) VALUES ($1, $2, $3)`, AuthServiceSchema, TableConfirmationIds)
//
// 	id := genEmailConfirmationId()
// 	expiresAt := time.Now().Add(p.confirmationOpts.ExpiresAfter)
// 	if _, err := p.db.ExecContext(ctx, smt, email, id, expiresAt); err != nil {
// 		return "", fmt.Errorf("failed to save email confirmation id: %v", err)
// 	}
//
// 	return id, nil
// }
//
// var (
// 	confirmUserAccountSqlQuerySmt = fmt.Sprintf(`
// WITH updated_row AS (
//     UPDATE "%s"."%s" u
//     SET "activated" = true
//     WHERE EXISTS (
//         SELECT 1
//         FROM "%s"."%s"
//         WHERE "id" = $1 AND "email" = u.email AND CURRENT_TIMESTAMP <= "expires_at"
//     ) AND "activated" = false
//     RETURNING id
// )
// SELECT COUNT(id) AS rows_updated FROM updated_row;
// `, AuthServiceSchema, TableUsers, AuthServiceSchema, TableConfirmationIds)
// )
//
// func (p *PostgresUserProvider) ConfirmEmailByConfirmationId(ctx context.Context, id string) error {
// 	if err := validateEmailConfirmationId(id); err != nil {
// 		return fmt.Errorf("wrong email confirmation id: %w", err)
// 	}
//
// 	smt := fmt.Sprintf(`SELECT "email", "expires_at" FROM "%s"."%s" WHERE "id" = $1`, AuthServiceSchema, TableConfirmationIds)
//
// 	var (
// 		email     string
// 		expiresAt time.Time
// 	)
// 	row := p.db.QueryRowContext(ctx, smt, id)
// 	if err := row.Scan(&email, &expiresAt); err != nil {
// 		return err
// 	}
//
// 	if time.Now().After(expiresAt) {
// 		return errors.New("confirmation id expires")
// 	}
//
// 	var rowsUpdated int
// 	row = p.db.QueryRow(confirmUserAccountSqlQuerySmt, id)
// 	if err := row.Scan(&rowsUpdated); err != nil {
// 		return err
// 	}
//
// 	if rowsUpdated == 0 {
// 		// TODO: separate errors
// 		return errors.New("wrong confirmation id, token expired or user account already activated")
// 	}
//
// 	return nil
// }
//
// func (p *PostgresUserProvider) FindUser(ctx context.Context, strId string) (*domain.User, error) {
// 	id, err := UserIdToInt64(strId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate id for user creation: %v", err)
// 	}
//
// 	smt := fmt.Sprintf(`SELECT id, name, type FROM "%s"."%s" WHERE "id" = $1`, AuthServiceSchema, TableUsers)
//
// 	row := p.db.QueryRowContext(ctx, smt, id)
// 	var intId int64
// 	var user domain.User
// 	if err := row.Scan(&intId, &user.Name, &user.Type); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, fmt.Errorf("%w: %w", domain.ErrUserNotFound, err)
// 		}
// 		return nil, err
// 	}
// 	user.Id = Int64ToUserId(intId)
//
// 	return &user, nil
// }
// func (p *PostgresUserProvider) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
// 	smt := fmt.Sprintf(`SELECT id, name, type FROM "%s"."%s" WHERE "email" = $1`, AuthServiceSchema, TableUsers)
//
// 	row := p.db.QueryRowContext(ctx, smt, email)
// 	var intId int64
// 	var user domain.User
// 	if err := row.Scan(&intId, &user.Name, &user.Type); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, fmt.Errorf("%w: %w", domain.ErrUserNotFound, err)
// 		}
// 		return nil, err
// 	}
// 	user.Id = Int64ToUserId(intId)
//
// 	return &user, nil
// }
// func (p *PostgresUserProvider) CheckUserCredentials(ctx context.Context, email string, password string) (*domain.User, error) {
// 	smt := fmt.Sprintf(`SELECT id, name, password, type FROM "%s"."%s" WHERE "email" = $1`, AuthServiceSchema, TableUsers)
//
// 	row := p.db.QueryRowContext(ctx, smt, email)
// 	var intId int64
// 	var name, dbPassword, userType string
// 	if err := row.Scan(&intId, &name, &dbPassword, &userType); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, domain.ErrInvalidCredentials
// 		}
// 		return nil, fmt.Errorf("failed to get user password from db: %v", err)
// 	}
//
// 	if ok := p.ph.Check(password, dbPassword); !ok {
// 		return nil, domain.ErrInvalidCredentials
// 	}
//
// 	return &domain.User{
// 		Id:   Int64ToUserId(intId),
// 		Name: name,
// 		Type: userType,
// 	}, nil
// }
//
// type PostgresRefreshTokenPersisterProvider struct {
// 	conf *postgres.DBConf
// 	db   *sql.DB
// }
//
// var _ domain.RefreshTokenPersisterProvider = (*PostgresRefreshTokenPersisterProvider)(nil)
//
// func NewPostgresRefreshTokenPersisterProvider(conf *postgres.DBConf, db *sql.DB) *PostgresRefreshTokenPersisterProvider {
// 	return &PostgresRefreshTokenPersisterProvider{
// 		conf: conf,
// 		db:   db,
// 	}
// }
//
// func (p *PostgresRefreshTokenPersisterProvider) Get(ctx context.Context, subjectId string) ([]string, error) {
// 	id, err := UserIdToInt64(subjectId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate id for user creation: %v", err)
// 	}
//
// 	smt := fmt.Sprintf(`SELECT id FROM "%s"."%s" WHERE "user_id" = $1 AND CURRENT_TIMESTAMP <= "expires_at"`, AuthServiceSchema, TableRefreshTokens)
//
// 	rows, err := p.db.QueryContext(ctx, smt, id)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query tokens from the database: %v", err)
// 	}
// 	defer func() {
// 		if err := rows.Close(); err != nil {
// 			log.Print(err)
// 		}
// 	}()
//
// 	var ids []string
// 	for rows.Next() {
// 		var id string
// 		if err := rows.Scan(&id); err != nil {
// 			return nil, fmt.Errorf("error scanning refresh token id: %v", err)
// 		}
// 		ids = append(ids, id)
// 	}
//
// 	return ids, nil
// }
// func (p *PostgresRefreshTokenPersisterProvider) Add(ctx context.Context, token *domain.RefreshToken) error {
// 	intId, err := UserIdToInt64(token.SubjectId)
// 	if err != nil {
// 		return err
// 	}
//
// 	smt := fmt.Sprintf(`INSERT INTO "%s"."%s" (id, user_id, expires_at) VALUES ($1, $2, $3)`, AuthServiceSchema, TableRefreshTokens)
//
// 	row := p.db.QueryRowContext(ctx, smt, token.TokenId, intId, token.ExpiresAt)
// 	if err := row.Err(); err != nil {
// 		return fmt.Errorf("failed to insert token into the database: %w", err)
// 	}
//
// 	return nil
// }
// func (p *PostgresRefreshTokenPersisterProvider) Delete(ctx context.Context, tokenId string) error {
// 	smt := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE "id" = $1`, AuthServiceSchema, TableRefreshTokens)
//
// 	row := p.db.QueryRowContext(ctx, smt, tokenId)
// 	if err := row.Err(); err != nil {
// 		return fmt.Errorf("failed to delete token from the database: %w", err)
// 	}
//
// 	return nil
// }
//
// func (p *PostgresUserProvider) GetIsUserConfirmedByEmail(ctx context.Context, email string) (bool, error) {
// 	smt := fmt.Sprintf(`SELECT "activated" FROM "%s"."%s" WHERE "email" = $1`, AuthServiceSchema, TableUsers)
//
// 	row := p.db.QueryRowContext(ctx, smt, email)
// 	var activated bool
// 	if err := row.Scan(&activated); err != nil {
// 		return false, err
// 	}
// 	return activated, nil
//
// }
//
// const (
// 	EmailConfirmationIdByteLen = 32
// 	EmailConfirmationIdPrefix  = "emconf"
// )
//
// func genEmailConfirmationId() string {
// 	return resource.GenerateIdPrefix(EmailConfirmationIdByteLen, EmailConfirmationIdPrefix)
// }
//
// func validateEmailConfirmationId(str string) error {
// 	return resource.ValidateIdByteLenPrefix(str, 32, "emconf")
// }
