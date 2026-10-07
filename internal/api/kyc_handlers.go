package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type KYCSubmitRequest struct {
	FullName       string `json:"full_name"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	Country        string `json:"country"`
}

type KYCReviewRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"` // approved, rejected
	Notes  string `json:"notes"`
}

// HandleKYCSubmit submits a new KYC verification or resubmits
// an existing verification for the authenticated user.
func (s *Server) HandleKYCSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	userID, ok := principalUserID64(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req KYCSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if req.FullName == "" ||
		req.DocumentType == "" ||
		req.DocumentNumber == "" ||
		req.Country == "" {
		writeErr(w, http.StatusBadRequest, "missing_required_fields")
		return
	}

	const query = `
		INSERT INTO kyc_verifications (
			user_id,
			full_name,
			document_type,
			document_number,
			country,
			status,
			submitted_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW(), NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET
			full_name = EXCLUDED.full_name,
			document_type = EXCLUDED.document_type,
			document_number = EXCLUDED.document_number,
			country = EXCLUDED.country,
			status = 'pending',
			notes = NULL,
			submitted_at = NOW(),
			updated_at = NOW()
	`

	if _, err := s.DB.Exec(
		query,
		userID,
		req.FullName,
		req.DocumentType,
		req.DocumentNumber,
		req.Country,
	); err != nil {
		log.Printf("error submitting KYC verification: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "pending",
		"message": "kyc_submitted_successfully",
	})
}

// HandleKYCStatus returns the KYC status of the authenticated user.
func (s *Server) HandleKYCStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	userID, ok := principalUserID64(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var (
		id             int64
		fullName       string
		documentType   string
		documentNumber string
		country        string
		status         string
		notes          sql.NullString
		submittedAt    time.Time
		updatedAt      time.Time
	)

	const query = `
		SELECT
			id,
			full_name,
			document_type,
			document_number,
			country,
			status,
			notes,
			submitted_at,
			updated_at
		FROM kyc_verifications
		WHERE user_id = $1
	`

	err := s.DB.QueryRow(query, userID).Scan(
		&id,
		&fullName,
		&documentType,
		&documentNumber,
		&country,
		&status,
		&notes,
		&submittedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "not_submitted",
		})
		return
	}

	if err != nil {
		log.Printf("error querying KYC verification status: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":              id,
		"status":          status,
		"full_name":       fullName,
		"document_type":   documentType,
		"document_number": documentNumber,
		"country":         country,
		"notes":           notes.String,
		"submitted_at":    submittedAt,
		"updated_at":      updatedAt,
	})
}

// HandleAdminPendingKYC returns all pending KYC verifications.
func (s *Server) HandleAdminPendingKYC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	const query = `
		SELECT
			id,
			user_id,
			full_name,
			document_type,
			document_number,
			country,
			status,
			notes,
			submitted_at,
			updated_at
		FROM kyc_verifications
		WHERE status = 'pending'
		ORDER BY submitted_at DESC
	`

	rows, err := s.DB.Query(query)
	if err != nil {
		log.Printf("error querying pending KYC verifications: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}
	defer rows.Close()

	list := make([]map[string]interface{}, 0)

	for rows.Next() {
		var (
			id             int64
			userID         int64
			fullName       string
			documentType   string
			documentNumber string
			country        string
			status         string
			notes          sql.NullString
			submittedAt    time.Time
			updatedAt      time.Time
		)

		if err := rows.Scan(
			&id,
			&userID,
			&fullName,
			&documentType,
			&documentNumber,
			&country,
			&status,
			&notes,
			&submittedAt,
			&updatedAt,
		); err != nil {
			log.Printf("error scanning pending KYC verification: %v", err)
			writeErr(w, http.StatusInternalServerError, "database_error")
			return
		}

		list = append(list, map[string]interface{}{
			"id":              id,
			"user_id":         userID,
			"full_name":       fullName,
			"document_type":   documentType,
			"document_number": documentNumber,
			"country":         country,
			"status":          status,
			"notes":           notes.String,
			"submitted_at":    submittedAt,
			"updated_at":      updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating pending KYC verifications: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": list,
	})
}

// HandleAdminReviewKYC approves or rejects a KYC verification.
func (s *Server) HandleAdminReviewKYC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	var req KYCReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if req.ID <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid_id")
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		writeErr(w, http.StatusBadRequest, "invalid_status")
		return
	}

	const query = `
		UPDATE kyc_verifications
		SET
			status = $1,
			notes = $2,
			updated_at = NOW()
		WHERE id = $3
	`

	result, err := s.DB.Exec(query, req.Status, req.Notes, req.ID)
	if err != nil {
		log.Printf("error updating KYC verification review: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Printf("error checking affected KYC verification rows: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	if affectedRows == 0 {
		writeErr(w, http.StatusNotFound, "kyc_verification_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "kyc_status_updated",
	})
}
