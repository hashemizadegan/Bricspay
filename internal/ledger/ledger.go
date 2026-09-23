package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
)

type Account struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type Posting struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type TransactionRequest struct {
	IdempotencyKey string    `json:"idempotency_key"`
	Description    string    `json:"description"`
	Postings       []Posting `json:"postings"`
}

type TransactionResponse struct {
	TransactionID  string `json:"transaction_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
}

func CreateAccount(ctx context.Context, db *sql.DB, code, accType, currency string) (*Account, error) {
	var a Account
	err := db.QueryRowContext(ctx,
		`INSERT INTO accounts (code, type, currency) VALUES ($1,$2,$3)
		 RETURNING id, code, type, currency, balance`,
		code, accType, currency,
	).Scan(&a.ID, &a.Code, &a.Type, &a.Currency, &a.Balance)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func ListAccounts(ctx context.Context, db *sql.DB) ([]Account, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, code, type, currency, balance FROM accounts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Code, &a.Type, &a.Currency, &a.Balance); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func RecordTransaction(ctx context.Context, db *sql.DB, req TransactionRequest) (*TransactionResponse, error) {
	if len(req.Postings) < 2 {
		return nil, errors.New("at least two postings required")
	}

	var sum float64
	for _, p := range req.Postings {
		sum += p.Amount
	}
	if math.Abs(sum) > 0.00000001 {
		return nil, fmt.Errorf("unbalanced transaction, sum = %.8f", sum)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var txID string
	err = tx.QueryRowContext(ctx,
		"INSERT INTO transactions (idempotency_key, description) VALUES ($1,$2) RETURNING id",
		req.IdempotencyKey, req.Description,
	).Scan(&txID)
	if err != nil {
		return nil, fmt.Errorf("idempotency conflict or insert failed: %w", err)
	}

	for _, p := range req.Postings {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO postings (transaction_id, account_id, amount) VALUES ($1,$2,$3)",
			txID, p.AccountID, p.Amount,
		); err != nil {
			return nil, err
		}
		res, err := tx.ExecContext(ctx,
			"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
			p.Amount, p.AccountID,
		)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, fmt.Errorf("account not found: %s", p.AccountID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &TransactionResponse{TransactionID: txID, IdempotencyKey: req.IdempotencyKey, Status: "COMMITTED"}, nil
}
