package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrInsufficientFunds = errors.New("ledger: insufficient funds")
	ErrSameWallet        = errors.New("ledger: cannot transfer to the same wallet")
	ErrDuplicate         = errors.New("ledger: duplicate idempotency key")
)

type TransferInput struct {
	FromWalletID   int64
	ToWalletID     int64
	AmountMinor    int64 // کوچک‌ترین یکا (cents)
	Currency       string
	Reference      string
	IdempotencyKey string
}

type Service struct{ db *sql.DB }

func New(database *sql.DB) *Service { return &Service{db: database} }

func (s *Service) Transfer(ctx context.Context, in TransferInput) error {
	if in.FromWalletID == in.ToWalletID {
		return ErrSameWallet
	}
	if in.AmountMinor <= 0 {
		return errors.New("ledger: amount must be positive")
	}
	if in.IdempotencyKey == "" {
		return errors.New("ledger: idempotency key required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("ledger: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// قفل ردیف‌ها به ترتیب صعودی برای جلوگیری از deadlock
	first, second := in.FromWalletID, in.ToWalletID
	if second < first {
		first, second = second, first
	}
	if err := lockWallet(ctx, tx, first); err != nil {
		return err
	}
	if first != second {
		if err := lockWallet(ctx, tx, second); err != nil {
			return err
		}
	}

	// idempotency: اگر همین کلید قبلاً ثبت شده، همان نتیجه برگردان
	var existing string
	err = tx.QueryRowContext(ctx,
		`SELECT idempotency_key FROM transactions WHERE idempotency_key = $1`,
		in.IdempotencyKey).Scan(&existing)
	if err == nil {
		return ErrDuplicate
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("ledger: idempotency check: %w", err)
	}

	var balance int64
	if err := tx.QueryRowContext(ctx,
		`SELECT balance FROM wallets WHERE id = $1`, in.FromWalletID).Scan(&balance); err != nil {
		return fmt.Errorf("ledger: fetch balance: %w", err)
	}
	if balance < in.AmountMinor {
		return ErrInsufficientFunds
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance - $1 WHERE id = $2`,
		in.AmountMinor, in.FromWalletID); err != nil {
		return fmt.Errorf("ledger: debit: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + $1 WHERE id = $2`,
		in.AmountMinor, in.ToWalletID); err != nil {
		return fmt.Errorf("ledger: credit: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions
			(from_wallet, to_wallet, amount, currency, status, reference, idempotency_key)
		VALUES ($1, $2, $3, $4, 'completed', $5, $6)`,
		in.FromWalletID, in.ToWalletID, in.AmountMinor, in.Currency,
		in.Reference, in.IdempotencyKey); err != nil {
		return fmt.Errorf("ledger: insert transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ledger: commit: %w", err)
	}
	committed = true
	return nil
}

func lockWallet(ctx context.Context, tx *sql.Tx, walletID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM wallets WHERE id = $1 FOR UPDATE`, walletID).Scan(&id)
	if err != nil {
		return fmt.Errorf("ledger: lock wallet %d: %w", walletID, err)
	}
	return nil
}
