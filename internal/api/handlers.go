package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bricspayir/internal/ledger"
)

type Server struct{ DB *sql.DB }

func NewServer(db *sql.DB) *Server { return &Server{DB: db} }

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Bricspay Core Ledger</title>
			<style>
				body { font-family: -apple-system, sans-serif; padding: 50px; line-height: 1.6; color: #333; max-width: 800px; margin: auto; }
				h1 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; }
				.box { background: #f4f7f6; padding: 20px; border-radius: 8px; }
				code { background: #eee; padding: 2px 5px; border-radius: 4px; }
			</style>
		</head>
		<body>
			<h1>Bricspay Core Ledger</h1>
			<p>Operational financial ledger service for BRICS Pay consortium.</p>
			<div class="box">
				<h3>Available Endpoints:</h3>
				<ul>
					<li><code>GET /health</code> - Check service status</li>
					<li><code>GET /api/v1/accounts</code> - List all accounts</li>
					<li><code>POST /api/v1/accounts</code> - Create new account</li>
					<li><code>POST /api/v1/transactions</code> - Record financial transaction</li>
				</ul>
			</div>
			<p><small>Status: System Online</small></p>
		</body>
		</html>
	`))
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "bricspayir-core-ledger"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), s.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(accounts)
	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Type     string `json:"type"`
			Currency string `json:"currency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		acc, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := ledger.RecordTransaction(r.Context(), s.DB, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
