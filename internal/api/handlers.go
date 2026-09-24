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
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	html, err := content.ReadFile("static/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load page")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
			writeError(w, http.StatusBadRequest, "invalid JSON request")
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
			writeError(w, http.StatusBadRequest, "one or more fields are too long")
			return
		}
		if !isASCIIAlphaNumeric(req.Code) || !isASCIIAlphaNumeric(req.Type) ||
			!isASCIIAlphaNumeric(req.Currency) {
			writeError(w, http.StatusBadRequest, "fields may contain only letters and numbers")
			return
		}

				account, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
		if err != nil {
			if errors.Is(err, ledger.ErrDuplicateAccount) {
				writeError(w, http.StatusConflict, "account code already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not create account")
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
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ledger.TransactionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}
	if len(req.IdempotencyKey) > 100 {
		writeError(w, http.StatusBadRequest, "idempotency_key is too long")
		return
	}
	if len(req.Postings) < 2 {
		writeError(w, http.StatusBadRequest, "at least two postings are required")
		return
	}
	for _, posting := range req.Postings {
		if strings.TrimSpace(posting.AccountID) == "" || posting.Amount == 0 {
			writeError(w, http.StatusBadRequest, "each posting needs an account_id and non-zero amount")
			return
		}
	}

	result, err := ledger.RecordTransaction(r.Context(), s.DB, &req)
	if err != nil {
		switch {
		case errors.Is(err, ledger.ErrDuplicateIdempotencyKey):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, ledger.ErrAccountNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ledger.ErrInsufficientBalance):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ledger.ErrCurrencyMismatch):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			msg := err.Error()
			if strings.HasPrefix(msg, "a transaction requires") ||
				strings.HasPrefix(msg, "posting amount cannot") ||
				strings.HasPrefix(msg, "posting account_id cannot") ||
				strings.HasPrefix(msg, "transaction postings are unbalanced") {
				writeError(w, http.StatusBadRequest, msg)
				return
			}
			writeError(w, http.StatusInternalServerError, "could not process transaction")
		}
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request must contain a single JSON value")
		}
		return err
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
	for _, r := range value {
		if r > unicode.MaxASCII ||
			!(r >= 'a' && r <= 'z') &&
				!(r >= 'A' && r <= 'Z') &&
				!(r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
