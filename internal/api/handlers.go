package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bricspayir/internal/ledger"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./internal/api/static/index.html")
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		rows, err := s.DB.Query("SELECT id, name, balance, currency FROM accounts")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var accounts []map[string]interface{}
		for rows.Next() {
			var id, name, currency string
			var balance float64
			if err := rows.Scan(&id, &name, &balance, &currency); err != nil {
				continue
			}
			accounts = append(accounts, map[string]interface{}{
				"id":       id,
				"name":     name,
				"balance":  balance,
				"currency": currency,
			})
		}
		json.NewEncoder(w).Encode(accounts)
		return
	}

	if r.Method == http.MethodPost {
		var acc struct {
			Name     string  `json:"name"`
			Currency string  `json:"currency"`
			Balance  float64 `json:"balance"`
		}
		if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := s.DB.Exec("INSERT INTO accounts (name, currency, balance) VALUES ($1, $2, $3)", acc.Name, acc.Currency, acc.Balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "account created"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}
