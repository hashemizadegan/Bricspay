package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"bricspay/internal/ledger"
)

// Server holds shared dependencies for all API handlers
type Server struct {
	DB  *sql.DB
	Ldg *ledger.Service
}

// NewServer initializes Server with database and ledger service
func NewServer(database *sql.DB) *Server {
	return &Server{
		DB:  database,
		Ldg: ledger.New(database),
	}
}

// Global JSON helper functions for the api package
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// HealthCheck returns server health status
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "bricspay-api",
		"version": "v11.0.0",
	})
}

// ReadyCheck verifies database connectivity
func (s *Server) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	if s.DB != nil {
		if err := s.DB.Ping(); err != nil {
			writeErr(w, http.StatusServiceUnavailable, "database_unreachable")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// HandleAccounts handles account balance and listing
func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.DB.Query(`SELECT id, user_id, currency, balance, created_at FROM accounts`)
		if err != nil {
			log.Printf("DB error querying accounts: %v", err)
			writeErr(w, http.StatusInternalServerError, "database_error")
			return
		}
		defer rows.Close()

		var accounts []map[string]interface{}
		for rows.Next() {
			var id, userID int64
			var currency string
			var balance float64
			var createdAt string
			if err := rows.Scan(&id, &userID, &currency, &balance, &createdAt); err == nil {
				accounts = append(accounts, map[string]interface{}{
					"id":         id,
					"user_id":    userID,
					"currency":   currency,
					"balance":    balance,
					"created_at": createdAt,
				})
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"accounts": accounts})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

// HandleTransactions handles transaction listing
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.DB.Query(`SELECT id, from_account_id, to_account_id, amount, currency, status, created_at FROM transactions ORDER BY created_at DESC LIMIT 50`)
		if err != nil {
			log.Printf("DB error querying transactions: %v", err)
			writeErr(w, http.StatusInternalServerError, "database_error")
			return
		}
		defer rows.Close()

		var txs []map[string]interface{}
		for rows.Next() {
			var id, fromAcc, toAcc int64
			var amount float64
			var currency, status, createdAt string
			if err := rows.Scan(&id, &fromAcc, &toAcc, &amount, &currency, &status, &createdAt); err == nil {
				txs = append(txs, map[string]interface{}{
					"id":              id,
					"from_account_id": fromAcc,
					"to_account_id":   toAcc,
					"amount":          amount,
					"currency":        currency,
					"status":          status,
					"created_at":      createdAt,
				})
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"transactions": txs})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}
