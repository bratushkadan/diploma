package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 25
	DefaultConnMaxIdleTime = 1 * time.Second
	DefaultDbPingTimeout   = 3 * time.Second
)

type DBConf struct {
	MaxOpenConns    *int
	MaxIdleConns    *int
	ConnMaxIdleTime *time.Duration

	DbPingTimeout *time.Duration

	DbHost     *string
	DbUser     *string
	DbPassword *string
	DbPort     *int
	DbName     *string
}

func NewDBConf() *DBConf {
	return &DBConf{}
}

func (c *DBConf) WithMaxOpenConns(n int) *DBConf {
	c.MaxOpenConns = &n
	return c
}
func (c *DBConf) WithMaxIdleConns(n int) *DBConf {
	c.MaxIdleConns = &n
	return c
}
func (c *DBConf) WithConnMaxIdleTime(n time.Duration) *DBConf {
	c.ConnMaxIdleTime = &n
	return c
}
func (c *DBConf) WithDbPingTimeout(d time.Duration) *DBConf {
	c.DbPingTimeout = &d
	return c
}
func (c *DBConf) WithDbHost(host string) *DBConf {
	c.DbHost = &host
	return c
}
func (c *DBConf) WithDbUser(user string) *DBConf {
	c.DbUser = &user
	return c
}
func (c *DBConf) WithDbPassword(password string) *DBConf {
	c.DbPassword = &password
	return c
}
func (c *DBConf) WithDbPort(port int) *DBConf {
	c.DbPort = &port
	return c
}
func (c *DBConf) WithDbName(name string) *DBConf {
	c.DbName = &name
	return c
}

func (c *DBConf) Build() (*DBConf, error) {
	if c.DbHost == nil {
		return nil, errors.New("database host must be set in DBConf")
	}
	if c.DbUser == nil {
		return nil, errors.New("database user must be set in DBConf")
	}
	if c.DbPassword == nil {
		return nil, errors.New("database password must be set in DBConf")
	}
	if c.DbPort == nil {
		return nil, errors.New("database port must be set in DBConf")
	}
	if c.DbName == nil {
		return nil, errors.New("database name must be set in DBConf")
	}

	if c.MaxOpenConns == nil {
		v := DefaultMaxOpenConns
		c.MaxOpenConns = &v
	}
	if c.MaxIdleConns == nil {
		v := DefaultMaxIdleConns
		c.MaxIdleConns = &v
	}
	if c.ConnMaxIdleTime == nil {
		v := DefaultConnMaxIdleTime
		c.ConnMaxIdleTime = &v
	}
	if c.DbPingTimeout == nil {
		v := DefaultDbPingTimeout
		c.DbPingTimeout = &v
	}

	return c, nil
}

func NewDB(conf *DBConf) (*sql.DB, error) {
	ps := fmt.Sprintf("host=%s user=%s password=%s port=%d dbname=%s sslmode=disable", *conf.DbHost, *conf.DbUser, *conf.DbPassword, *conf.DbPort, *conf.DbName)

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database client: %w", err)
	}

	db.SetMaxOpenConns(*conf.MaxOpenConns)
	db.SetMaxIdleConns(*conf.MaxIdleConns)
	db.SetConnMaxIdleTime(*conf.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), *conf.DbPingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
