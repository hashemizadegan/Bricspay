package api

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"bricspay/internal/auth"
	"bricspay/internal/ledger"
)

type Server struct {
	DB            *sql.DB
	LedgerService *ledger.Service
	AuthService   *auth.Service
	StaticFS      fs.FS
}

func NewServer(db *sql.DB, ls *ledger.Service, as *auth.Service, staticFS fs.FS) *Server {
	sub, _ := fs.Sub(staticFS, "static")
	return &Server{
		DB:            db,
		LedgerService: ls,
		AuthService:   as,
		StaticFS:      sub,
	}
}

func (s *Server) StaticFileServer() http.Handler {
	if s.StaticFS == nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(s.StaticFS))
}

func (s *Server) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.AuthService.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		r = r.WithContext(auth.ContextWithUser(r.Context(), claims))
		next.ServeHTTP(w, r)
	}
}

func (s *Server) AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return s.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.UserFromContext(r.Context())
		if claims == nil || claims.Role != "admin" {
			http.Error(w, `{"error":"forbidden: admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "BRICS Pay Settlement API Gateway",
		"status":  "online",
		"version": "1.0.0",
	})
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbStatus := "connected"
	if s.DB == nil || s.DB.Ping() != nil {
		dbStatus = "disconnected/degraded"
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"database": dbStatus,
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		accounts, err := s.LedgerService.ListAccounts(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list accounts"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(accounts)
	case http.MethodPost:
		var req ledger.Account
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
			return
		}
		acc, err := s.LedgerService.CreateAccount(r.Context(), req.Name, req.Type, req.Currency)
		if err != nil {
			http.Error(w, `{"error":"failed to create account"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid transaction payload"}`, http.StatusBadRequest)
		return
	}
	tx, err := s.LedgerService.RecordTransaction(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	banks := []map[string]string{
		{"id": "BANK-RU-01", "name": "VTB Bank", "country": "RU", "bic": "VTBRRU22"},
		{"id": "BANK-IR-01", "name": "Mir Business Bank", "country": "IR", "bic": "MIRBIRT1"},
		{"id": "BANK-CN-01", "name": "Bank of China", "country": "CN", "bic": "BKCHCNBJ"},
	}
	json.NewEncoder(w).Encode(banks)
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": "demo-authenticated-token"})
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user registered successfully"})
}
