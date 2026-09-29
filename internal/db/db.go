package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// ساخت دیتابیس اولیه
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// فراخوانی جدید برای ساخت جداول KYC و Users
	if err := MigrateKYC(db); err != nil {
		return nil, fmt.Errorf("failed to MigrateKYC: %w", err)
	}

	return db, nil
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		code VARCHAR(64) UNIQUE NOT NULL,
		currency VARCHAR(3) NOT NULL,
		balance BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idempotency_key VARCHAR(128) UNIQUE NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(32) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS postings (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
		account_id UUID NOT NULL REFERENCES accounts(id),
		amount BIGINT NOT NULL,
		currency VARCHAR(3) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS banks (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(128) NOT NULL,
		country VARCHAR(3) NOT NULL,
		bic VARCHAR(32) NOT NULL,
		role VARCHAR(32) NOT NULL,
		protocol VARCHAR(32) NOT NULL DEFAULT 'REST_API',
		contact_email VARCHAR(128) NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING_VERIFICATION',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_postings_tx_id ON postings(transaction_id);
	CREATE INDEX IF NOT EXISTS idx_postings_acc_id ON postings(account_id);
	CREATE INDEX IF NOT EXISTS idx_banks_bic ON banks(bic);
	`
	_, err := db.Exec(schema)
	return err
}
