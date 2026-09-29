package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"bricspay/internal/ledger"
)

// Server ساختار اصلی برای نگهداری اتصال دیتابیس
type Server struct {
	DB *sql.DB
}

// NewServer سازنده سرور
func NewServer(database *sql.DB) *Server {
	return &Server{DB: database}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

// HandleTransactions متد اصلاح شده با اتصال دیتابیس
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// ارسال اتصال دیتابیس به تابع لجر
	resp, err := ledger.RecordTransaction(r.Context(), s.DB, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, resp)
}
