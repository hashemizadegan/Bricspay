package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"bricspay/internal/ledger"
)

//go:embed static/*
var staticFS embed.FS

// StaticFS فایل‌های استاتیک را در اختیار FileServer قرار می‌دهد
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

type Server struct {
	ledger *ledger.Ledger
	db     *sql.DB
}

func NewServer(database *sql.DB) *Server {
	return &Server{
		ledger: ledger.NewLedger(database),
		db:     database,
	}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Index file not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "operational",
		"system": "BRICS Pay Core Settlement",
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		accounts, err := s.ledger.GetAccounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(accounts)
	case http.MethodPost:
		var acc ledger.Account
		if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.ledger.CreateAccount(&acc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := s.ledger.ProcessTransaction(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	banks := []map[string]string{
		{"code": "VTB", "name": "VTB Bank (Russia)", "status": "Connected"},
		{"code": "CBI", "name": "Central Bank of Iran", "status": "Connected"},
		{"code": "BOC", "name": "Bank of China", "status": "Pending"},
		{"code": "NDB", "name": "New Development Bank", "status": "Integrated"},
	}
	json.NewEncoder(w).Encode(banks)
}
