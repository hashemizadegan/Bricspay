package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/lib/pq"
)

var (
	ErrDuplicateAccount        = errors.New("account code already exists")
	ErrDuplicateIdempotencyKey = errors.New("idempotency key already used")
	ErrAccountNotFound         = errors.New("account not found")
	ErrCurrencyMismatch        = errors.New("postings must use the same currency")
	ErrInsufficientBalance     = errors.New("insufficient balance")
)

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

type Account struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Type      string  `json:"type"`
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type Posting struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type TransactionRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	Description    string         `json:"description"`
	Postings       []Posting      `json:"postings"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type TransactionResponse struct {
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Description    string    `json:"description"`
	Postings       []Posting `json:"postings"`
	CreatedAt      string    `json:"created_at"`
}

func CreateAccount(ctx context.Context, db *sql.DB, code, accType, currency string) (*Account, error) {
	query := `
		INSERT INTO accounts (code, type, currency, balance)
		VALUES ($1, $2, $3, 0)
		RETURNING id, code, type, currency, balance, created_at;
	`
	var acc Account
	err := db.QueryRowContext(ctx, query, code, accType, currency).Scan(
		&acc.ID,
		&acc.Code,
		&acc.Type,
		&acc.Currency,
		&acc.Balance,
		&acc.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateAccount
		}
		return nil, fmt.Errorf("create account: %w", err)
	}
	return &acc, nil
}

func ListAccounts(ctx context.Context, db *sql.DB) ([]Account, error) {
	query := `
		SELECT id, code, type, currency, balance, created_at
		FROM accounts
		ORDER BY created_at DESC;
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(
			&acc.ID,
			&acc.Code,
			&acc.Type,
			&acc.Currency,
			&acc.Balance,
			&acc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	if accounts == nil {
		accounts = []Account{}
	}
	return accounts, nil
}

func RecordTransaction(ctx context.Context, db *sql.DB, req *TransactionRequest) (*TransactionResponse, error) {
	if len(req.Postings) < 2 {
		return nil, fmt.Errorf("a transaction requires at least two postings")
	}

	var sum float64
	for _, p := range req.Postings {
		if p.Amount == 0 {
			return nil, fmt.Errorf("posting amount cannot be zero")
		}
		if p.AccountID == "" {
			return nil, fmt.Errorf("posting account_id cannot be empty")
		}
		sum += p.Amount
	}

	if math.Abs(sum) > 1e-8 {
		return nil, fmt.Errorf("transaction postings are unbalanced: sum is %f", sum)
	}

	accountIDs := make([]string, 0, len(req.Postings))
	seen := make(map[string]bool)
	for _, p := range req.Postings {
		if !seen[p.AccountID] {
			seen[p.AccountID] = true
			accountIDs = append(accountIDs, p.AccountID)
		}
	}
	sort.Strings(accountIDs)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	accountsByID := make(map[string]Account)
	for _, id := range accountIDs {
		var acc Account
		err := tx.QueryRowContext(ctx, `
			SELECT id, code, type, currency, balance
			FROM accounts
			WHERE id = $1
			FOR UPDATE;
		`, id).Scan(&acc.ID, &acc.Code, &acc.Type, &acc.Currency, &acc.Balance)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: %s", ErrAccountNotFound, id)
			}
			return nil, fmt.Errorf("lock account %s: %w", id, err)
		}
		accountsByID[id] = acc
	}

	var baseCurrency string
	for _, id := range accountIDs {
		curr := accountsByID[id].Currency
		if baseCurrency == "" {
			baseCurrency = curr
		} else if baseCurrency != curr {
			return nil, ErrCurrencyMismatch
		}
	}

	var txID string
	var createdAt string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (idempotency_key, description, metadata)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`, req.IdempotencyKey, req.Description, "{}").Scan(&txID, &createdAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateIdempotencyKey
		}
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	postingStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO postings (transaction_id, account_id, amount)
		VALUES ($1, $2, $3);
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare posting stmt: %w", err)
	}
	defer postingStmt.Close()

	for _, p := range req.Postings {
		if _, err := postingStmt.ExecContext(ctx, txID, p.AccountID, p.Amount); err != nil {
			return nil, fmt.Errorf("insert posting: %w", err)
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE accounts
			SET balance = balance + $1
			WHERE id = $2 AND (balance + $1 >= 0);
		`, p.Amount, p.AccountID)
		if err != nil {
			return nil, fmt.Errorf("update account balance: %w", err)
		}

		rowsAff, err := res.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("check rows affected: %w", err)
		}
		if rowsAff == 0 {
			return nil, fmt.Errorf("%w: %s", ErrInsufficientBalance, p.AccountID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &TransactionResponse{
		ID:             txID,
		IdempotencyKey: req.IdempotencyKey,
		Description:    req.Description,
		Postings:       req.Postings,
		CreatedAt:      createdAt,
	}, nil
}
