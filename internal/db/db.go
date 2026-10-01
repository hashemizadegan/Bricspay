package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(dsn string) (*DB, error) { ... } // keep the rest exactly as extracted

func InitSchema(db *sql.DB) error { ... }
func MigrateWalletAuth(db *sql.DB) error { ... }
