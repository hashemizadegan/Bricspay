package api

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"bricspayir/internal/ledger"
)

//go:embed static/index.html
var staticFS embed.FS

type Server struct {
	db *ledgerAccountService
}

type ledgerAccountService struct {
	db any
}

func NewServer(db any) *Server {
	return &Server{db: &ledgerAccountService{db: db}}
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

	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load UI")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	sqlDB := s.db.db.(*databaseWrapper).DB

	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), sqlDB)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not retrieve accounts")
			return
		}
		writeJSON(w, http.StatusOK, accounts)

	case http.MethodPost:
		var body struct {
			Code     string `json:"code"`
			Type     string `json:"type"`
			Currency string `json:"currency"`
		}

		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		body.Code = strings.TrimSpace(body.Code)
		body.Type = strings.TrimSpace(body.Type)
		body.Currency = strings.ToUpper(strings.TrimSpace(body.Currency))

		if body.Code == "" || body.Type == "" || body.Currency == "" {
			writeError(w, http.StatusBadRequest, "code, type, and currency are required")
			return
		}

		if len(body.Code) > 50 || len(body.Type) > 20 || len(body.Currency) > 10 {
			writeError(w, http.StatusBadRequest, "field length exceeds limits")
			return
		}

		for _, ch := range body.Currency {
			if ch < 'A' || ch > 'Z' {
				writeError(w, http.StatusBadRequest, "currency must be uppercase ASCII letters")
				return
			}
		}

		acc, err := ledger.CreateAccount(r.Context(), sqlDB, body.Code, body.Type, body.Currency)
		if err != nil {
			if errors.Is(err, ledger.ErrDuplicateAccount) {
				writeError(w, http.StatusConflict, "account code already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not create account")
			return
		}

		writeJSON(w, http.StatusCreated, acc)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sqlDB := s.db.db.(*databaseWrapper).DB

	var req ledger.TransactionRequest
	if err := decodeJSON(r, &req); err != nil {
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
		writeError(w, http.StatusBadRequest, "idempotency_key exceeds maximum length of 100")
		return
	}

	if len(req.Postings) < 2 {
		writeError(w, http.StatusBadRequest, "a transaction requires at least two postings")
		return
	}

	for _, p := range req.Postings {
		if strings.TrimSpace(p.AccountID) == "" {
			writeError(w, http.StatusBadRequest, "account_id cannot be empty")
			return
		}
		if p.Amount == 0 {
			writeError(w, http.StatusBadRequest, "posting amount cannot be zero")
			return
		}
	}

	txResp, err := ledger.RecordTransaction(r.Context(), sqlDB, &req)
	if err != nil {
		switch {
		case errors.Is(err, ledger.ErrDuplicateIdempotencyKey):
			writeError(w, http.StatusConflict, "idempotency key already used")
		case errors.Is(err, ledger.ErrAccountNotFound):
			writeError(w, http.StatusBadRequest, "unknown account in postings")
		case errors.Is(err, ledger.ErrCurrencyMismatch):
			writeError(w, http.StatusBadRequest, "all postings must use the same currency")
		case errors.Is(err, ledger.ErrInsufficientBalance):
			writeError(w, http.StatusUnprocessableEntity, "insufficient balance")
		default:
			writeError(w, http.StatusInternalServerError, "could not record transaction")
		}
		return
	}

	writeJSON(w, http.StatusCreated, txResp)
}

type databaseWrapper struct {
	DB any
}
