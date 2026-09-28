package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

type BankPartner struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Country   string    `json:"country"`
	BIC       string    `json:"bic"`
	Role      string    `json:"role"` // ISSUING, ADVISING, CONFIRMING, SETTLEMENT
	Protocol  string    `json:"protocol"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Server struct {
	DB    *sql.DB
	banks []BankPartner
	mu    sync.RWMutex
}

func NewServer(db *sql.DB) *Server {
	s := &Server{
		DB: db,
		banks: []BankPartner{
			{
				ID:        "bank-ir-01",
				Name:      "Tejarat Bank (IR)",
				Country:   "IR",
				BIC:       "TEJIRTH",
				Role:      "ISSUING",
				Protocol:  "REST_API / SEPAM",
				Email:     "foreign-trade@tejaratbank.ir",
				Status:    "ACTIVE",
				CreatedAt: time.Now().Add(-48 * time.Hour),
			},
			{
				ID:        "bank-ru-01",
				Name:      "Sberbank Corporate (RU)",
				Country:   "RU",
				BIC:       "SABBRUMM",
				Role:      "ADVISING",
				Protocol:  "ISO20022 / SPFS",
				Email:     "trade-settlement@sber.ru",
				Status:    "ACTIVE",
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
		},
	}
	if db != nil {
		_, _ = db.Exec(`
			CREATE TABLE IF NOT EXISTS banks (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				country TEXT NOT NULL,
				bic TEXT NOT NULL,
				role TEXT NOT NULL,
				protocol TEXT NOT NULL,
				email TEXT NOT NULL,
				status TEXT NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
		`)
	}
	return s
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/banks" {
		s.HandleBanks(w, r)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbStatus := "connected"
	if s.DB == nil {
		dbStatus = "fallback_memory"
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"database": dbStatus,
		"version":  "v2.2-bank-agnostic",
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		if s.DB == nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": "acc-1", "code": "IRR-TREASURY-01", "currency": "IRR", "balance": 15000000000},
				{"id": "acc-2", "code": "RUB-NOSTRO-01", "currency": "RUB", "balance": 45000000},
			})
			return
		}
		rows, err := s.DB.Query("SELECT id, code, currency, balance, created_at FROM accounts ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, `{"error":"failed to query accounts"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		res := []map[string]interface{}{}
		for rows.Next() {
			var id, code, curr string
			var bal int64
			var t time.Time
			if err := rows.Scan(&id, &code, &curr, &bal, &t); err == nil {
				res = append(res, map[string]interface{}{"id": id, "code": code, "currency": curr, "balance": bal, "created_at": t})
			}
		}
		json.NewEncoder(w).Encode(res)

	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Currency string `json:"currency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.Currency == "" {
			http.Error(w, `{"error":"invalid parameters"}`, http.StatusBadRequest)
			return
		}
		if s.DB == nil {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"id": "mock-acc", "code": req.Code, "currency": req.Currency, "balance": 0})
			return
		}
		var newID string
		err := s.DB.QueryRow("INSERT INTO accounts (code, currency, balance) VALUES ($1, $2, 0) RETURNING id", req.Code, req.Currency).Scan(&newID)
		if err != nil {
			http.Error(w, `{"error":"account creation failed"}`, http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "code": req.Code, "currency": req.Currency, "balance": 0})
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ACCEPTED", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) HandleBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		if s.DB != nil {
			rows, err := s.DB.Query("SELECT id, name, country, bic, role, protocol, email, status, created_at FROM banks ORDER BY created_at DESC")
			if err == nil {
				defer rows.Close()
				dbBanks := []BankPartner{}
				for rows.Next() {
					var b BankPartner
					if err := rows.Scan(&b.ID, &b.Name, &b.Country, &b.BIC, &b.Role, &b.Protocol, &b.Email, &b.Status, &b.CreatedAt); err == nil {
						dbBanks = append(dbBanks, b)
					}
				}
				if len(dbBanks) > 0 {
					json.NewEncoder(w).Encode(dbBanks)
					return
				}
			}
		}
		s.mu.RLock()
		defer s.mu.RUnlock()
		json.NewEncoder(w).Encode(s.banks)
		return
	}

	if r.Method == http.MethodPost {
		var b BankPartner
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.Name == "" || b.BIC == "" {
			http.Error(w, `{"error":"invalid bank data"}`, http.StatusBadRequest)
			return
		}
		b.ID = "bank-" + strings.ToLower(b.Country) + "-" + time.Now().Format("150405")
		b.Status = "ACTIVE"
		b.CreatedAt = time.Now().UTC()

		if s.DB != nil {
			_, _ = s.DB.Exec(
				"INSERT INTO banks (id, name, country, bic, role, protocol, email, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
				b.ID, b.Name, b.Country, b.BIC, b.Role, b.Protocol, b.Email, b.Status, b.CreatedAt,
			)
		}

		s.mu.Lock()
		s.banks = append([]BankPartner{b}, s.banks...)
		s.mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(b)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}
