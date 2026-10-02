package api

import (
	"database/sql"
	"log"
	"net/http"

	"bricspay/internal/ledger"
)

type Server struct {
	DB  *sql.DB
	Ldg *ledger.Service
}

func NewServer(database *sql.DB) *Server {
	return &Server{DB: database, Ldg: ledger.New(database)}
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ReadyCheck آمادگی واقعی سرویس را بررسی می‌کند.
func (s *Server) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.PingContext(r.Context()); err != nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type account struct {
	ID      int64  `json:"id"`
	Code    string `json:"code"`
	Balance int64  `json:"balance_minor"` // کوچک‌ترین یکای پول، نه float64
	Currency string `json:"currency"`
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := s.DB.QueryContext(r.Context(),
			`SELECT id, code, balance, currency FROM accounts`)
		if err != nil {
			log.Printf("HandleAccounts query error: %v", err)
			writeErr(w, http.StatusInternalServerError, "internal server error")
			return
		}
		defer rows.Close()

		accounts := make([]account, 0)
		for rows.Next() {
			var a account
			if err := rows.Scan(&a.ID, &a.Code, &a.Balance, &a.Currency); err != nil {
				log.Printf("HandleAccounts scan error: %v", err)
				writeErr(w, http.StatusInternalServerError, "internal server error")
				return
			}
			accounts = append(accounts, a)
		}
		if err := rows.Err(); err != nil {
			log.Printf("HandleAccounts rows error: %v", err)
			writeErr(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, accounts)

	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Currency string `json:"currency"`
			Balance  int64  `json:"balance_minor"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Code == "" || req.Currency == "" {
			writeErr(w, http.StatusBadRequest, "code and currency are required")
			return
		}
		if req.Balance < 0 {
			writeErr(w, http.StatusBadRequest, "balance must not be negative")
			return
		}
		if _, err := s.DB.ExecContext(r.Context(),
			`INSERT INTO accounts (code, currency, balance) VALUES ($1, $2, $3)`,
			req.Code, req.Currency, req.Balance); err != nil {
			log.Printf("HandleAccounts insert error: %v", err)
			writeErr(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type transferRequest struct {
	FromWalletID   int64  `json:"from_wallet_id"`
	ToWalletID     int64  `json:"to_wallet_id"`
	AmountMinor    int64  `json:"amount_minor"`
	Currency       string `json:"currency"`
	Reference      string `json:"reference"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.DB == nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	var req transferRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FromWalletID == 0 || req.ToWalletID == 0 || req.AmountMinor <= 0 {
		writeErr(w, http.StatusBadRequest, "from_wallet_id, to_wallet_id and positive amount_minor are required")
		return
	}
	if req.IdempotencyKey == "" {
		writeErr(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}

	err := s.Ldg.Transfer(r.Context(), ledger.TransferInput{
		FromWalletID:   req.FromWalletID,
		ToWalletID:     req.ToWalletID,
		AmountMinor:    req.AmountMinor,
		Currency:       req.Currency,
		Reference:      req.Reference,
		IdempotencyKey: req.IdempotencyKey,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
	case errors.Is(err, ledger.ErrInsufficientFunds):
		writeErr(w, http.StatusConflict, "insufficient funds")
	case errors.Is(err, ledger.ErrSameWallet):
		writeErr(w, http.StatusBadRequest, "cannot transfer to the same wallet")
	default:
		log.Printf("HandleTransactions ledger error: %v", err) // جزئیات فقط در لاگ
		writeErr(w, http.StatusUnprocessableEntity, "transaction failed")
	}
}
