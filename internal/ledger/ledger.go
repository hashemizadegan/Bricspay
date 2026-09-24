package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
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
	var account Account
	err := db.QueryRowContext(ctx,
		`INSERT INTO accounts (code, type, currency)
		 VALUES ($1, $2, $3)
		 RETURNING id, code, type, currency, balance`,
		code, accType, currency,
	).Scan(
		&account.ID,
		&account.Code,
		&account.Type,
		&account.Currency,
		&account.Balance,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func ListAccounts(ctx context.Context, db *sql.DB) ([]Account, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, code, type, currency, balance
		 FROM accounts
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var account Account
		if err := rows.Scan(
			&account.ID,
			&account.Code,
			&account.Type,
			&account.Currency,
			&account.Balance,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func RecordTransaction(
	ctx context.Context,
	db *sql.DB,
	req TransactionRequest,
) (*TransactionResponse, error) {
	if len(req.Postings) < 2 {
		return nil, errors.New("at least two postings required")
	}
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	var sum float64
	for _, posting := range req.Postings {
		if posting.AccountID == "" {
			return nil, errors.New("posting account ID is required")
		}
		if math.IsNaN(posting.Amount) || math.IsInf(posting.Amount, 0) {
			return nil, errors.New("posting amount must be a finite number")
		}
		sum += posting.Amount
	}
	if math.Abs(sum) > 0.00000001 {
		return nil, fmt.Errorf("unbalanced transaction, sum = %.8f", sum)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock accounts in a stable order to reduce deadlocks between concurrent transactions.
	accountIDs := make([]string, 0, len(req.Postings))
	seen := make(map[string]struct{}, len(req.Postings))
	for _, posting := range req.Postings {
		if _, exists := seen[posting.AccountID]; exists {
			continue
		}
		seen[posting.AccountID] = struct{}{}
		accountIDs = append(accountIDs, posting.AccountID)
	}
	sort.Strings(accountIDs)

	var transactionCurrency string
	for i, accountID := range accountIDs {
		var currency string
		err := tx.QueryRowContext(ctx,
			`SELECT currency
			 FROM accounts
			 WHERE id = $1
			 FOR UPDATE`,
			accountID,
		).Scan(&currency)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %s", accountID)
		}
		if err != nil {
			return nil, err
		}

		if i == 0 {
			transactionCurrency = currency
		} else if currency != transactionCurrency {
			return nil, errors.New("all postings in a transaction must use the same currency")
		}
	}

	var transactionID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO transactions (idempotency_key, description)
		 VALUES ($1, $2)
		 RETURNING id`,
		req.IdempotencyKey,
		req.Description,
	).Scan(&transactionID)
	if err != nil {
		return nil, fmt.Errorf("idempotency conflict or transaction insert failed: %w", err)
	}

	for _, posting := range req.Postings {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO postings (transaction_id, account_id, amount)
			 VALUES ($1, $2, $3)`,
			transactionID,
			posting.AccountID,
			posting.Amount,
		); err != nil {
			return nil, err
		}

		result, err := tx.ExecContext(ctx,
			`UPDATE accounts
			 SET balance = balance + $1
			 WHERE id = $2
			   AND balance + $1 >= 0`,
			posting.Amount,
			posting.AccountID,
		)
		if err != nil {
			return nil, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected == 0 {
			// The account was already checked and locked above; zero rows means
			// the balance constraint prevented the update.
			return nil, fmt.Errorf("insufficient balance for account: %s", posting.AccountID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &TransactionResponse{
		TransactionID:  transactionID,
		IdempotencyKey: req.IdempotencyKey,
		Status:         "COMMITTED",
	}, nil
}
