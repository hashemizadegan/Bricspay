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

// InitSchema sets up all required database tables idempotently.
func InitSchema(db *sql.DB) error {
	if err := MigrateKYC(db); err != nil {
		return err
	}
	return MigrateWalletAuth(db)
}

// MigrateWalletAuth creates wallet auth and challenge tracking tables.
func MigrateWalletAuth(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS wallet_accounts (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		wallet_address TEXT NOT NULL UNIQUE,
		chain_id       BIGINT NOT NULL DEFAULT 1,
		is_primary     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS wallet_challenges (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		wallet_address TEXT NOT NULL,
		nonce          TEXT NOT NULL UNIQUE,
		domain         TEXT NOT NULL,
		message        TEXT NOT NULL,
		expires_at     TIMESTAMPTZ NOT NULL,
		consumed       BOOLEAN NOT NULL DEFAULT FALSE,
		created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		actor_id      TEXT,
		actor_email   TEXT,
		action        TEXT NOT NULL,
		target_entity TEXT NOT NULL,
		target_id     TEXT,
		ip_address    TEXT,
		metadata      JSONB,
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_wallet_challenges_addr ON wallet_challenges(wallet_address);
	CREATE INDEX IF NOT EXISTS idx_wallet_challenges_nonce ON wallet_challenges(nonce);
	CREATE INDEX IF NOT EXISTS idx_wallet_accounts_addr ON wallet_accounts(wallet_address);
	`
	_, err := db.Exec(schema)
	return err
}
