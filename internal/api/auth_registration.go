package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registrationRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	LegalName    string `json:"legal_name"`
	Jurisdiction string `json:"jurisdiction"`
	Nationality  string `json:"nationality"`
	Residence    string `json:"residence"`
	ReferralID   string `json:"referral_id"`
}

func NewRegistrationHandler(db *sql.DB) http.Handler {
	var (
		schemaOnce sync.Once
		schemaErr  error
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeRegistrationJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
			return
		}

		if db == nil {
			writeRegistrationJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "database_unavailable"})
			return
		}

		schemaOnce.Do(func() {
			_, schemaErr = db.ExecContext(
				r.Context(),
				`CREATE TABLE IF NOT EXISTS bricspay_users (
					id UUID PRIMARY KEY,
					email TEXT NOT NULL UNIQUE,
					password_hash TEXT NOT NULL,
					legal_name TEXT NOT NULL DEFAULT '',
					jurisdiction TEXT NOT NULL DEFAULT '',
					nationality TEXT NOT NULL DEFAULT '',
					residence TEXT NOT NULL DEFAULT '',
					referral_id TEXT NOT NULL DEFAULT '',
					created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`,
			)
		})

		if schemaErr != nil {
			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{"error": "database_schema_error", "message": schemaErr.Error()})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		defer r.Body.Close()

		var req registrationRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if req.Email == "" || !strings.Contains(req.Email, "@") {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_email"})
			return
		}

		if len([]byte(req.Password)) < 8 {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{"error": "password_must_contain_at_least_8_characters"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{"error": "password_hashing_failed"})
			return
		}

		uid := uuid.New()
		_, err = db.ExecContext(
			r.Context(),
			`INSERT INTO bricspay_users (id, email, password_hash, legal_name, jurisdiction, nationality, residence, referral_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uid, req.Email, string(hash), req.LegalName, req.Jurisdiction, req.Nationality, req.Residence, req.ReferralID,
		)

		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				writeRegistrationJSON(w, http.StatusConflict, map[string]any{"error": "email_already_registered"})
				return
			}
			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{"error": "registration_failed", "message": err.Error()})
			return
		}

		writeRegistrationJSON(w, http.StatusCreated, map[string]any{
			"status":  "created",
			"message": "Registration successful",
			"user_id": uid.String(),
		})
	})
}

func writeRegistrationJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
