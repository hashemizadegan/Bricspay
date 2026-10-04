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

// NewRegistrationHandler creates the HTTP registration endpoint.
func NewRegistrationHandler(db *sql.DB) http.Handler {
	var (
		schemaOnce sync.Once
		schemaErr  error
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeRegistrationJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"error": "method_not_allowed",
			})
			return
		}

		if db == nil {
			writeRegistrationJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error": "database_unavailable",
			})
			return
		}

		// Create a dedicated user table if it does not already exist.
		// This avoids depending on an incomplete existing accounts schema.
		schemaOnce.Do(func() {
			_, schemaErr = db.ExecContext(
				r.Context(),
				`
				CREATE TABLE IF NOT EXISTS bricspay_users (
					id UUID PRIMARY KEY,
					email TEXT NOT NULL UNIQUE,
					password_hash TEXT NOT NULL,
					legal_name TEXT NOT NULL DEFAULT '',
					jurisdiction TEXT NOT NULL DEFAULT '',
					nationality TEXT NOT NULL DEFAULT '',
					residence TEXT NOT NULL DEFAULT '',
					referral_id TEXT NOT NULL DEFAULT '',
					created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
				)
				`,
			)
		})

		if schemaErr != nil {
			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "database_schema_error",
				"message": schemaErr.Error(),
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		defer r.Body.Close()

		var request registrationRequest

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&request); err != nil {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid_json",
			})
			return
		}

		request.Email = strings.ToLower(strings.TrimSpace(request.Email))
		request.LegalName = strings.TrimSpace(request.LegalName)
		request.Jurisdiction = strings.TrimSpace(request.Jurisdiction)
		request.Nationality = strings.TrimSpace(request.Nationality)
		request.Residence = strings.TrimSpace(request.Residence)
		request.ReferralID = strings.TrimSpace(request.ReferralID)

		if request.Email == "" {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{
				"error": "email_required",
			})
			return
		}

		if !strings.Contains(request.Email, "@") {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid_email",
			})
			return
		}

		if len([]byte(request.Password)) < 8 {
			writeRegistrationJSON(w, http.StatusBadRequest, map[string]any{
				"error": "password_must_contain_at_least_8_characters",
			})
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(request.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "password_hashing_failed",
			})
			return
		}

		userID := uuid.New()

		_, err = db.ExecContext(
			r.Context(),
			`
			INSERT INTO bricspay_users (
				id,
				email,
				password_hash,
				legal_name,
				jurisdiction,
				nationality,
				residence,
				referral_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`,
			userID,
			request.Email,
			string(passwordHash),
			request.LegalName,
			request.Jurisdiction,
			request.Nationality,
			request.Residence,
			request.ReferralID,
		)

		if err != nil {
			errorText := strings.ToLower(err.Error())

			if strings.Contains(errorText, "duplicate") ||
				strings.Contains(errorText, "unique constraint") {
				writeRegistrationJSON(w, http.StatusConflict, map[string]any{
					"error": "email_already_registered",
				})
				return
			}

			writeRegistrationJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "registration_failed",
				"message": err.Error(),
			})
			return
		}

		writeRegistrationJSON(w, http.StatusCreated, map[string]any{
			"status":  "created",
			"message": "Registration successful",
			"user_id": userID.String(),
		})
	})
}

func writeRegistrationJSON(
	w http.ResponseWriter,
	status int,
	payload any,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
