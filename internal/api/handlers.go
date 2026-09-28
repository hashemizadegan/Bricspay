package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode"

	"bricspay/internal/ledger"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./static/index.html")
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "online",
		"service": "BRICS Pay Settlement Gateway",
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), s.DB)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not list accounts")
			return
		}
		WriteJSON(w, http.StatusOK, accounts)

	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Type     string `json:"type"`
			Currency string `json:"currency"`
		}
		if err := DecodeJSON(w, r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		req.Code = strings.TrimSpace(req.Code)
		req.Type = strings.TrimSpace(req.Type)
		req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
		if req.Code == "" || req.Type == "" || req.Currency == "" {
			WriteError(w, http.StatusBadRequest, "code, type, and currency are required")
			return
		}
		if len(req.Code) > 50 || len(req.Type) > 20 || len(req.Currency) > 10 {
			WriteError(w, http.StatusBadRequest, "code, type, or currency exceeds the allowed length")
			return
		}
		if !isASCIIAlphaNumeric(req.Currency) {
			WriteError(w, http.StatusBadRequest, "currency must contain only ASCII letters or digits")
			return
		}

		account, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not create account; code may already exist")
			return
		}
		WriteJSON(w, http.StatusCreated, account)

	default:
		w.Header().Set("Allow", "GET, POST")
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ledger.TransactionRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	req.Description = strings.TrimSpace(req.Description)
	if req.IdempotencyKey == "" {
		WriteError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}
	if len(req.IdempotencyKey) > 100 {
		WriteError(w, http.StatusBadRequest, "idempotency_key must be at most 100 characters")
		return
	}
	if len(req.Postings) < 2 {
		WriteError(w, http.StatusBadRequest, "at least two postings are required")
		return
	}
	for _, posting := range req.Postings {
		if strings.TrimSpace(posting.AccountID) == "" {
			WriteError(w, http.StatusBadRequest, "every posting must have an account_id")
			return
		}
		if posting.Amount == 0 {
			WriteError(w, http.StatusBadRequest, "posting amounts must not be zero")
			return
		}
	}

	result, err := ledger.RecordTransaction(r.Context(), s.DB, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, result)
}

func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	banks := []map[string]string{
		{"id": "vtb", "name": "VTB Bank (PJSC)", "country": "RU", "bic": "044525187"},
		{"id": "sber", "name": "Sberbank", "country": "RU", "bic": "044525225"},
		{"id": "bmi", "name": "Bank Melli Iran", "country": "IR", "bic": "MELIIRTH"},
	}
	WriteJSON(w, http.StatusOK, banks)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return errors.New("request body must be at most 1 MB")
		}
		return errors.New("invalid JSON body")
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON value")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func isASCIIAlphaNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII || (!(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
