package api

import (
	"encoding/json"
	"net/http"

	"bricspay/internal/ledger"
)

// Helper for JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// Helper for Error responses
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// TransactionHandler handles ledger transaction recording
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// نکته کلیدی رفع خطای بیلد: ارسال اشاره‌گر (&req) به متد
	resp, err := ledger.RecordTransaction(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}
