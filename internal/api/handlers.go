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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BRICS Pay | Trade Finance & Settlement</title>
    <style>
        body { font-family: 'Inter', -apple-system, sans-serif; margin: 0; background: #0a0a0a; color: #e0e0e0; line-height: 1.6; }
        .hero { padding: 100px 20px; text-align: center; background: linear-gradient(180deg, #1a2e2a 0%, #0a0a0a 100%); }
        h1 { font-size: 3rem; color: #fff; margin-bottom: 20px; }
        .container { max-width: 800px; margin: 0 auto; padding: 40px 20px; }
        .feature-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-top: 40px; }
        .card { background: #161616; padding: 25px; border-radius: 12px; border: 1px solid #333; }
        footer { text-align: center; padding: 40px; font-size: 0.8rem; color: #666; }
    </style>
</head>
<body>
    <div class="hero">
        <h1>Move business forward, together.</h1>
        <p>Advanced cross-border settlement infrastructure for modern trade.</p>
    </div>
    <div class="container">
        <div class="feature-grid">
            <div class="card"><h3>Secure Ledger</h3><p>Immutable record keeping for international transactions.</p></div>
            <div class="card"><h3>Real-time Settlement</h3><p>Efficient processing of trade finance instruments.</p></div>
        </div>
    </div>
    <footer>© 2026 BRICS Pay Consortium</footer>
</body>
</html>`))
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet { http.Error(w, "Method not allowed", 405); return }
	accounts, _ := ledger.ListAccounts(r.Context(), s.DB)
	json.NewEncoder(w).Encode(accounts)
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost { http.Error(w, "Method not allowed", 405); return }
	var req ledger.TransactionRequest
	json.NewDecoder(r.Body).Decode(&req)
	resp, _ := ledger.RecordTransaction(r.Context(), s.DB, req)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
