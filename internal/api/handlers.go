package api

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

type Server struct {
	DB        *sql.DB
	mockBanks []BankPartner
	mu        sync.RWMutex
}

type BankPartner struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	BIC          string    `json:"bic"`
	Role         string    `json:"role"`
	Protocol     string    `json:"protocol"`
	ContactEmail string    `json:"contact_email"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type AccountRequest struct {
	Code     string `json:"code"`
	Currency string `json:"currency"`
}

type PostingRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type TransactionRequest struct {
	IdempotencyKey string           `json:"idempotency_key"`
	Description    string           `json:"description"`
	Postings       []PostingRequest `json:"postings"`
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
		mockBanks: []BankPartner{
			{
				ID:           "b1-mock",
				Name:         "Tejarat Bank (IR)",
				Country:      "IR",
				BIC:          "TEJIRTH",
				Role:         "ISSUING",
				Protocol:     "REST_API",
				ContactEmail: "fx@tejaratbank.ir",
				Status:       "ACTIVE",
				CreatedAt:    time.Now().Add(-72 * time.Hour),
			},
			{
				ID:           "b2-mock",
				Name:         "Sberbank Corporate (RU)",
				Country:      "RU",
				BIC:          "SABBRUMM",
				Role:         "ADVISING",
				Protocol:     "ISO20022",
				ContactEmail: "trade-brics@sber.ru",
				Status:       "ACTIVE",
				CreatedAt:    time.Now().Add(-48 * time.Hour),
			},
		},
	}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(indexHTML)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "healthy"
	dbStatus := "connected"
	if s.DB == nil {
		dbStatus = "fallback_mock"
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status":   status,
		"database": dbStatus,
		"version":  "v2.1-agnostic",
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		if s.DB == nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": "acc-1", "code": "IRR-TREASURY", "currency": "IRR", "balance": 5000000000},
				{"id": "acc-2", "code": "RUB-NOSTRO", "currency": "RUB", "balance": 12000000},
			})
			return
		}
		rows, err := s.DB.Query("SELECT id, code, currency, balance, created_at FROM accounts ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, `{"error":"failed to query accounts"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		accounts := []map[string]interface{}{}
		for rows.Next() {
			var id, code, currency string
			var balance int64
			var createdAt time.Time
			if err := rows.Scan(&id, &code, &currency, &balance, &createdAt); err == nil {
				accounts = append(accounts, map[string]interface{}{
					"id":         id,
					"code":       code,
					"currency":   currency,
					"balance":    balance,
					"created_at": createdAt,
				})
			}
		}
		json.NewEncoder(w).Encode(accounts)

	case http.MethodPost:
		var req AccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.Currency == "" {
			http.Error(w, `{"error":"invalid account parameters"}`, http.StatusBadRequest)
			return
		}

		if s.DB == nil {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":       "mock-created-id",
				"code":     req.Code,
				"currency": req.Currency,
				"balance":  0,
			})
			return
		}

		var newID string
		err := s.DB.QueryRow(
			"INSERT INTO accounts (code, currency, balance) VALUES ($1, $2, 0) RETURNING id",
			req.Code, req.Currency,
		).Scan(&newID)

		if err != nil {
			http.Error(w, `{"error":"failed to create account or duplicate code"}`, http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       newID,
			"code":     req.Code,
			"currency": req.Currency,
			"balance":  0,
		})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IdempotencyKey == "" || len(req.Postings) < 2 {
		http.Error(w, `{"error":"invalid transaction payload or insufficient postings"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "ACCEPTED",
		"idempotency_key": req.IdempotencyKey,
		"description":     req.Description,
	})
}

func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if s.DB == nil {
			s.mu.RLock()
			defer s.mu.RUnlock()
			json.NewEncoder(w).Encode(s.mockBanks)
			return
		}

		rows, err := s.DB.Query("SELECT id, name, country, bic, role, protocol, contact_email, status, created_at FROM banks ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, `{"error":"failed to query banks"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		banks := []BankPartner{}
		for rows.Next() {
			var b BankPartner
			if err := rows.Scan(&b.ID, &b.Name, &b.Country, &b.BIC, &b.Role, &b.Protocol, &b.ContactEmail, &b.Status, &b.CreatedAt); err == nil {
				banks = append(banks, b)
			}
		}
		json.NewEncoder(w).Encode(banks)

	case http.MethodPost:
		var req BankPartner
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.BIC == "" || req.ContactEmail == "" {
			http.Error(w, `{"error":"invalid bank onboarding data"}`, http.StatusBadRequest)
			return
		}

		req.Status = "PENDING_VERIFICATION"
		req.CreatedAt = time.Now().UTC()

		if s.DB == nil {
			s.mu.Lock()
			req.ID = "bank-" + time.Now().Format("150405")
			s.mockBanks = append([]BankPartner{req}, s.mockBanks...)
			s.mu.Unlock()

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(req)
			return
		}

		var newID string
		err := s.DB.QueryRow(
			`INSERT INTO banks (name, country, bic, role, protocol, contact_email, status, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			req.Name, req.Country, req.BIC, req.Role, req.Protocol, req.ContactEmail, req.Status, req.CreatedAt,
		).Scan(&newID)

		if err != nil {
			http.Error(w, `{"error":"failed to register bank partner"}`, http.StatusInternalServerError)
			return
		}

		req.ID = newID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
