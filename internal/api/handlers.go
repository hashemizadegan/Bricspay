package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bricspay/internal/ledger"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"failed to marshal response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

// HandleTransactions متد اصلی ثبت تراکنش‌ها با اتصال به دیتابیس
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	resp, err := ledger.RecordTransaction(r.Context(), s.DB, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// سازگاری برای تست‌ها و فراخوانی قدیمی
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotImplemented, "Use Server.HandleTransactions")
}
