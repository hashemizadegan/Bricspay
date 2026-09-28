package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bricspay/internal/ledger"
)

// Server ساختار اصلی سرور API شامل اتصال دیتابیس
type Server struct {
	DB *sql.DB
}

// NewServer سازنده استراکت Server
func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
	}
}

// HandleRoot صفحه اصلی یا وضعیت روت را هندل می‌کند
func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "online",
		"message": "BRICS Pay Settlement API Gateway",
		"version": "1.0.0",
	})
}

// HealthCheck بررسی وضعیت سلامت سرور و اتصال دیتابیس
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "connected"
	if s.DB != nil {
		if err := s.DB.Ping(); err != nil {
			dbStatus = "disconnected: " + err.Error()
		}
	} else {
		dbStatus = "database not configured"
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"database": dbStatus,
	})
}

// HandleAccounts مدیریت ایجاد و فهرست حساب‌ها
func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(s.DB)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(accounts)

	case http.MethodPost:
		var req struct {
			ID       string  `json:"id"`
			Name     string  `json:"name"`
			Currency string  `json:"currency"`
			Balance  float64 `json:"balance"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
			return
		}

		acc, err := ledger.CreateAccount(s.DB, req.ID, req.Name, req.Currency, req.Balance)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(acc)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleTransactions ثبت و مشاهده تاریخچه تراکنش‌ها
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req struct {
			SenderID   string  `json:"sender_id"`
			ReceiverID string  `json:"receiver_id"`
			Amount     float64 `json:"amount"`
			Currency   string  `json:"currency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid transaction payload"}`, http.StatusBadRequest)
			return
		}

		tx, err := ledger.RecordTransaction(s.DB, req.SenderID, req.ReceiverID, req.Amount, req.Currency)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(tx)

	default:
		w.Header().Set("Allow", "POST")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleBanks فهرست بانک‌های عضو و متصل در شبکه BRICS Pay
func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// بازگرداندن فهرست بانک‌های متصل به گیت‌وی
	banks := []map[string]string{
		{"code": "CBR", "name": "Central Bank of Russia", "country": "RU"},
		{"code": "PBC", "name": "People's Bank of China", "country": "CN"},
		{"code": "RBI", "name": "Reserve Bank of India", "country": "IN"},
		{"code": "BCB", "name": "Banco Central do Brasil", "country": "BR"},
		{"code": "SARB", "name": "South African Reserve Bank", "country": "ZA"},
		{"code": "CBI", "name": "Central Bank of Iran", "country": "IR"},
	}

	_ = json.NewEncoder(w).Encode(banks)
}
