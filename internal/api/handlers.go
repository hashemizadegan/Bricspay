package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bricspay/internal/ledger"
)

// Server ساختار سرور API با اتصال دیتابیس
type Server struct {
	DB *sql.DB
}

// NewServer سازنده Server
func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
	}
}

// HandleRoot صفحه اصلی و سلامت سرویس
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

// HealthCheck بررسی وضعیت سلامت API و دیتابیس
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "connected"
	if s.DB != nil {
		if err := s.DB.PingContext(r.Context()); err != nil {
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

// HandleAccounts مدیریت ایجاد و دریافت لیست حساب‌ها
func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), s.DB)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(accounts)

	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Type     string `json:"type"`     // e.g. "business", "nostro", "vostro", "settlement"
			Currency string `json:"currency"` // e.g. "RUB", "IRR", "CNY"
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
			return
		}

		if req.Type == "" {
			req.Type = "business"
		}
		if req.Currency == "" {
			req.Currency = "RUB"
		}

		acc, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
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

// HandleTransactions ثبت و تسویه تراکنش‌ها با دفترکل دوبل
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req struct {
			SenderID       string  `json:"sender_id"`
			ReceiverID     string  `json:"receiver_id"`
			Amount         float64 `json:"amount"`
			Currency       string  `json:"currency"`
			Description    string  `json:"description"`
			IdempotencyKey string  `json:"idempotency_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid transaction payload"}`, http.StatusBadRequest)
			return
		}

		if req.Amount <= 0 {
			http.Error(w, `{"error": "Amount must be greater than zero"}`, http.StatusBadRequest)
			return
		}
		if req.SenderID == "" || req.ReceiverID == "" {
			http.Error(w, `{"error": "SenderID and ReceiverID are required"}`, http.StatusBadRequest)
			return
		}

		if req.IdempotencyKey == "" {
			req.IdempotencyKey = fmt.Sprintf("tx-%d", time.Now().UnixNano())
		}
		if req.Description == "" {
			req.Description = fmt.Sprintf("Transfer of %.2f %s", req.Amount, req.Currency)
		}

		// دفترکل دوبل: کسر از فرستنده و افزودن به گیرنده
		txReq := &ledger.TransactionRequest{
			IdempotencyKey: req.IdempotencyKey,
			Description:    req.Description,
			Postings: []ledger.Posting{
				{AccountID: req.SenderID, Amount: -req.Amount},
				{AccountID: req.ReceiverID, Amount: req.Amount},
			},
			Metadata: map[string]any{
				"currency": req.Currency,
			},
		}

		res, err := ledger.RecordTransaction(r.Context(), s.DB, txReq)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(res)

	default:
		w.Header().Set("Allow", "POST")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleBanks فهرست بانک‌های تسویه‌کننده در شبکه
func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	banks := []map[string]string{
		{"code": "CBR", "name": "Central Bank of Russia", "country": "RU"},
		{"code": "PBC", "name": "People's Bank of China", "country": "CN"},
		{"code": "RBI", "name": "Reserve Bank of India", "country": "IN"},
		{"code": "BCB", "name": "Banco Central do Brasil", "country": "BR"},
		{"code": "SARB", "name": "South African Reserve Bank", "country": "ZA"},
		{"code": "CBI", "name": "Central Bank of Iran", "country": "IR"},
		{"code": "VTB", "name": "VTB Bank (Trade Finance Gateway)", "country": "RU"},
	}

	_ = json.NewEncoder(w).Encode(banks)
}
