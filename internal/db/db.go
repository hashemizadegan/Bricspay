package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(connStr string) (*sql.DB, error) {
	d, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	d.SetMaxOpenConns(25)
	d.SetMaxIdleConns(10)
	d.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	log.Println("database connection established")

	if err := runMigrations(d); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return d, nil
}

func runMigrations(d *sql.DB) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		code VARCHAR(50) UNIQUE NOT NULL,
		type VARCHAR(20) NOT NULL,
		currency VARCHAR(10) NOT NULL,
		balance NUMERIC(28,8) NOT NULL DEFAULT 0.00000000,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idempotency_key VARCHAR(100) UNIQUE NOT NULL,
		description TEXT,
		metadata JSONB DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS postings (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
		account_id UUID NOT NULL REFERENCES accounts(id),
		amount NUMERIC(28,8) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_postings_tx ON postings(transaction_id);
	CREATE INDEX IF NOT EXISTS idx_postings_acc ON postings(account_id);
	`
	if _, err := d.Exec(schema); err != nil {
		return err
	}
	log.Println("schema migrations verified")
	return nil
}
