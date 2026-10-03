package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bricspay/internal/auth"
)

const maxUpload = 10 << 20 // 10MB

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// ---------- Auth ----------

type registerReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	EntityType        string `json:"entity_type"` // CORPORATE | BANK
	LegalName         string `json:"legal_name"`
	RegistrationNo    string `json:"registration_number"`
	TaxID             string `json:"tax_id"`
	Jurisdiction      string `json:"jurisdiction"`
	ContactPhone      string `json:"contact_phone"`
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"}); return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !strings.Contains(req.Email, "@") || len(req.Password) < 8 {
		writeJSON(w, 400, map[string]string{"error": "email invalid or password < 8 chars"}); return
	}
	if req.EntityType != "CORPORATE" && req.EntityType != "BANK" {
		writeJSON(w, 400, map[string]string{"error": "entity_type must be CORPORATE or BANK"}); return
	}
	if req.LegalName == "" || req.Jurisdiction == "" {
		writeJSON(w, 400, map[string]string{"error": "legal_name and jurisdiction required"}); return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil { writeJSON(w, 500, map[string]string{"error": "hash failed"}); return }

	var userID string
	err = s.DB.QueryRow(
		`INSERT INTO users (email, password_hash, entity_type, role, status)
		 VALUES ($1,$2,$3,'member','PENDING') RETURNING id`,
		req.Email, hash, req.EntityType).Scan(&userID)
	if err == sql.ErrNoRows || strings.Contains(fmt.Sprint(err), "duplicate key") {
		writeJSON(w, 409, map[string]string{"error": "email already registered"}); return
	}
	if err != nil { writeJSON(w, 500, map[string]string{"error": "db error"}); return }

	var profileID string
	err = s.DB.QueryRow(
		`INSERT INTO kyc_profiles (user_id, legal_name, registration_number, tax_id, jurisdiction, contact_phone, status)
		 VALUES ($1,$2,$3,$4,$5,$6,'SUBMITTED') RETURNING id`,
		userID, req.LegalName, req.RegistrationNo, req.TaxID, req.Jurisdiction, req.ContactPhone).Scan(&profileID)
	if err != nil { writeJSON(w, 500, map[string]string{"error": "profile error"}); return }

	s.logAudit(userID, req.Email, "REGISTER", "kyc_profile", profileID, r)
	writeJSON(w, 201, map[string]string{"user_id": userID, "profile_id": profileID, "status": "PENDING"})
}
func (s *Server) HandleKYCStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	p, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var (
		status, entity string
		reason         sql.NullString
		updated        time.Time
		resubmits      int
	)
	err := s.DB.QueryRow(`
SELECT COALESCE(status::text,'PENDING'), COALESCE(entity_type::text,'INDIVIDUAL'),
       rejection_reason, COALESCE(updated_at, created_at), COALESCE(resubmit_count,0)
  FROM kyc_profiles WHERE user_id = $1`, p.UserID).
		Scan(&status, &entity, &reason, &updated, &resubmits)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "PENDING", "reason": nil, "can_resubmit": true,
			"entity_type": "INDIVIDUAL", "resubmit_count": 0,
		})
		return
	}
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	can := status == "PENDING" || status == "REJECTED"
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          status,
		"reason":          reason.String,
		"can_resubmit":    can,
		"entity_type":     entity,
		"resubmit_count":  resubmits,
		"updated_at":      updated.UTC().Format(time.RFC3339),
	})
}

func (s *Server) HandleKYCResubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	p, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	res, err := s.DB.Exec(`
UPDATE kyc_profiles
   SET status = 'PENDING',
       rejection_reason = NULL,
       resubmit_count = COALESCE(resubmit_count,0) + 1,
       submitted_at = NOW(),
       updated_at = NOW()
 WHERE user_id = $1 AND status IN ('REJECTED','PENDING')`, p.UserID)
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, `{"error":"cannot_resubmit"}`, http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "PENDING"})
}

