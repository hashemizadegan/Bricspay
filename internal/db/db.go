package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// InitDB opens and configures the PostgreSQL connection pool,
// then runs all schema migrations.
func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	if err := MigrateKYC(db); err != nil {
		return nil, fmt.Errorf("migrate kyc: %w", err)
	}

	return db, nil
}

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name     TEXT        NOT NULL,
			code     TEXT        NOT NULL UNIQUE,
			type     TEXT        NOT NULL CHECK (type IN ('asset','liability','equity','income','expense')),
			balance  NUMERIC(20,8) NOT NULL DEFAULT 0,
			currency CHAR(3)     NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS transactions (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			idempotency_key  TEXT        NOT NULL UNIQUE,
			description      TEXT,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS postings (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_id UUID        NOT NULL REFERENCES transactions(id),
			account_id     UUID        NOT NULL REFERENCES accounts(id),
			amount         NUMERIC(20,8) NOT NULL,
			currency       CHAR(3)     NOT NULL,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}
