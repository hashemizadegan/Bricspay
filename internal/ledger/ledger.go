package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"bricspay/internal/db"
)

// ErrInsufficientFunds is returned when a wallet has an insufficient balance.
var ErrInsufficientFunds = errors.New("ledger: insufficient funds")

// ErrSameWallet is returned when source and destination are identical.
var ErrSameWallet = errors.New("ledger: cannot transfer to the same wallet")

// Service provides ledger operations.
type Service struct {
	db *db.DB
}

// New creates a ledger Service.
func New(database *db.DB) *Service {
	return &Service{db: database}
}

// Transfer moves amount from one wallet to another atomically under
// SERIALIZABLE isolation, preventing double-spending and phantom reads.
func (s *Service) Transfer(ctx context.Context, fromWalletID, toWalletID int64, amount float64, reference string) error {
	if fromWalletID == toWalletID {
		return ErrSameWallet
	}
	if amount <= 0 {
		return errors.New("ledger: amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("ledger: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Lock both rows in a consistent order (lower ID first) to prevent deadlocks.
	first, second := fromWalletID, toWalletID
	if second < first {
		first, second = second, first
	}

	if err = lockWallet(ctx, tx, first); err != nil {
		return err
	}
	if first != second {
		if err = lockWallet(ctx, tx, second); err != nil {
			return err
		}
	}

	// Verify source balance.
	var balance float64
	err = tx.QueryRowContext(ctx,
		`SELECT balance FROM wallets WHERE id = $1`, fromWalletID,
	).Scan(&balance)
	if err != nil {
		return fmt.Errorf("ledger: fetch balance: %w", err)
	}
	if balance < amount {
		err = ErrInsufficientFunds
		return err
	}

	// Debit source.
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance - $1 WHERE id = $2`,
		amount, fromWalletID,
	); err != nil {
		return fmt.Errorf("ledger: debit: %w", err)
	}

	// Credit destination.
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + $1 WHERE id = $2`,
		amount, toWalletID,
	); err != nil {
		return fmt.Errorf("ledger: credit: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (from_wallet, to_wallet, amount, currency, status, reference)
		VALUES ($1, $2, $3, 'USD', 'completed', $4)
	`, fromWalletID, toWalletID, amount, reference); err != nil {
		return fmt.Errorf("ledger: insert transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ledger: commit: %w", err)
	}
	return nil
}

func lockWallet(ctx context.Context, tx *sql.Tx, walletID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM wallets WHERE id = $1 FOR UPDATE`, walletID,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("ledger: lock wallet %d: %w", walletID, err)
	}
	return nil
}
