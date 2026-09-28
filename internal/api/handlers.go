package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"bricspay/internal/ledger"
)

//go:embed static/*
var staticFS embed.FS

// Server ساختار اصلی سرور API است
type Server struct {
	DB     *sql.DB
	Router http.Handler
}

// ServeHTTP به ساختار Server اجازه می‌دهد مستقیماً به عنوان http.Handler عمل کند
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

// StaticFS سیستم فایل استاتیک تعبیه شده را برمی‌گرداند
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// NewServer نمونه جدید سرور به همراه کلیه مسیرها (Routes) را مقداردهی می‌کند
func NewServer(db *sql.DB) *Server {
	s := &Server{
		DB: db,
	}

	mux := http.NewServeMux()

	// 1. Health Checks
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/health", s.handleHealth)

	// 2. Ledger Endpoints
	mux.HandleFunc("/api/v1/accounts", s.handleAccounts)
	mux.HandleFunc("/api/v1/transactions", s.handleTransactions)

	// 3. Auth Endpoints (پیاده‌سازی شده در kyc_handlers.go)
	mux.HandleFunc("/api/auth/register", s.HandleRegister)
	mux.HandleFunc("/api/auth/login", s.HandleLogin)

	// 4. KYC Endpoints (پیاده‌سازی شده در kyc_handlers.go)
	mux.HandleFunc("/api/kyc/upload", s.HandleKYCUpload)

	// 5. Admin Endpoints (پیاده‌سازی شده در kyc_handlers.go)
	mux.HandleFunc("/api/admin/profiles", s.HandleAdminProfiles)
	mux.HandleFunc("/api/admin/decision", s.HandleAdminDecision)
	mux.HandleFunc("/api/admin/audit", s.HandleAdminAudit)

	// 6. Static Web UI Files
	fileServer := http.FileServer(StaticFS())
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/static/") {
			// اگر مسیر API نباشد، کاربر به صفحه اصلی هدایت می‌شود
			fileServer.ServeHTTP(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	s.Router = mux
	return s
}

// handleHealth وضعیت سلامت سرویس را بررسی می‌کند
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbStatus := "connected"
	if s.DB != nil {
		if err := s.DB.PingContext(r.Context()); err != nil {
			dbStatus = "error: " + err.Error()
		}
	} else {
		dbStatus = "database not initialized"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "operational",
		"service":  "bricspay-ledger",
		"database": dbStatus,
	})
}

// handleAccounts مدیریت ایجاد و دریافت لیست حساب‌های لجر
func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListAccounts(w, r)
	case http.MethodPost:
		s.handleCreateAccount(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateAccount حساب جدید در لجر ثبت می‌کند
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code     string `json:"code"`
		Type     string `json:"type"`
		Currency string `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.Type == "" || req.Currency == "" {
		http.Error(w, "code, type, and currency are required", http.StatusBadRequest)
		return
	}

	acc, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

// handleListAccounts فهرست کلیه حساب‌ها را بازمی‌گرداند
func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := ledger.ListAccounts(r.Context(), s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if accounts == nil {
		accounts = []ledger.Account{}
	}
	json.NewEncoder(w).Encode(accounts)
}

// handleTransactions ثبت تراکنش جدید در سیستم دفتر کل
func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var txReq ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, "invalid transaction payload", http.StatusBadRequest)
		return
	}

	res, err := ledger.RecordTransaction(r.Context(), s.DB, &txReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
