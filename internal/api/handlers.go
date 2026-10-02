package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
    "bricspay/internal/db"
	"bricspay/internal/ledger"
)

// Server holds shared dependencies.
type Server struct {
	DB *sql.DB
}

// NewServer creates a new Server instance.
func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

// HealthCheck returns a simple liveness response.
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleAccounts lists or creates accounts.
func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {

	case http.MethodGet:
		rows, err := s.DB.QueryContext(r.Context(),
			`SELECT id, code, balance, currency FROM accounts`)
		if err != nil {
			log.Printf("HandleAccounts query error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type account struct {
			ID       int     `json:"id"`
			Code     string  `json:"code"`
			Balance  float64 `json:"balance"`
			Currency string  `json:"currency"`
		}

		var accounts []account
		for rows.Next() {
			var a account
			if err := rows.Scan(&a.ID, &a.Code, &a.Balance, &a.Currency); err != nil {
				log.Printf("HandleAccounts scan error: %v", err)
				continue
			}
			accounts = append(accounts, a)
		}
		if err := rows.Err(); err != nil {
			log.Printf("HandleAccounts rows error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accounts)

	case http.MethodPost:
		var req struct {
			Code     string  `json:"code"`
			Currency string  `json:"currency"`
			Balance  float64 `json:"balance"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Code == "" || req.Currency == "" {
			http.Error(w, "code and currency are required", http.StatusBadRequest)
			return
		}

		_, err := s.DB.ExecContext(r.Context(),
			`INSERT INTO accounts (code, currency, balance) VALUES ($1, $2, $3)`,
			req.Code, req.Currency, req.Balance)
		if err != nil {
			log.Printf("HandleAccounts insert error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTransactions processes a ledger transfer.
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.DB == nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	// JSON را به map تبدیل می‌کنیم (بدون نیاز به struct TransactionRequest)
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	fromID := getInt64(req, "from_wallet_id")
	toID := getInt64(req, "to_wallet_id")
	amount := getFloat64(req, "amount")
	reference := getString(req, "reference")

	l := ledger.New(&db.DB{DB: s.DB})
	if err := l.Transfer(r.Context(), fromID, toID, amount, reference); err != nil {
		log.Printf("HandleTransactions ledger error: %v", err)
		http.Error(w, "transaction failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// helper functions برای تبدیل مقدار
func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
