package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps *sql.DB with project-specific helpers.
type DB struct {
	*sql.DB
}

// New opens a PostgreSQL connection, verifies it, and configures the pool.
func New(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return &DB{sqlDB}, nil
}

// Migrate runs DDL migrations idempotently.
func (d *DB) Migrate() error {
	_, err := d.Exec(schema)
	if err != nil {
		return fmt.Errorf("db: migrate: %w", err)
	}
	return nil
}

// schema is the full DDL; append new tables here with IF NOT EXISTS.
const schema = `
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallets (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency   TEXT        NOT NULL DEFAULT 'USD',
    balance    NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, currency)
);

CREATE TABLE IF NOT EXISTS transactions (
    id            BIGSERIAL PRIMARY KEY,
    from_wallet   BIGINT        REFERENCES wallets(id),
    to_wallet     BIGINT        REFERENCES wallets(id),
    amount        NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency      TEXT          NOT NULL,
    status        TEXT          NOT NULL DEFAULT 'pending',
    reference     TEXT          NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_wallet);
CREATE INDEX IF NOT EXISTS idx_transactions_to   ON transactions(to_wallet);
+ kycSchema
