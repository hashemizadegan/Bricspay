package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode"

	"bricspayir/internal/ledger"
)

//go:embed static/index.html
var content embed.FS

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

	file, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(file)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), s.DB)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not list accounts")
			return
		}
		writeJSON(w, http.StatusOK, accounts)

	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Type     string `json:"type"`
			Currency string `json:"currency"`
		}
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		req.Code = strings.TrimSpace(req.Code)
		req.Type = strings.TrimSpace(req.Type)
		req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
		if req.Code == "" || req.Type == "" || req.Currency == "" {
			writeError(w, http.StatusBadRequest, "code, type, and currency are required")
			return
		}
		if len(req.Code) > 50 || len(req.Type) > 20 || len(req.Currency) > 10 {
			writeError(w, http.StatusBadRequest, "code, type, or currency exceeds the allowed length")
			return
		}
		if !isASCIIAlphaNumeric(req.Currency) {
			writeError(w, http.StatusBadRequest, "currency must contain only ASCII letters or digits")
			return
		}

		account, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create account; code may already exist")
			return
		}
		writeJSON(w, http.StatusCreated, account)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ledger.TransactionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	req.Description = strings.TrimSpace(req.Description)

	if req.IdempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}
	if len(req.IdempotencyKey) > 100 {
		writeError(w, http.StatusBadRequest, "idempotency_key must be at most 100 characters")
		return
	}
	if len(req.Postings) < 2 {
		writeError(w, http.StatusBadRequest, "at least two postings are required")
		return
	}
	for _, posting := range req.Postings {
		if strings.TrimSpace(posting.AccountID) == "" {
			writeError(w, http.StatusBadRequest, "every posting must have an account_id")
			return
		}
		if posting.Amount == 0 {
			writeError(w, http.StatusBadRequest, "posting amounts must not be zero")
			return
		}
	}

	result, err := ledger.RecordTransaction(r.Context(), s.DB, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func isASCIIAlphaNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII || !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

