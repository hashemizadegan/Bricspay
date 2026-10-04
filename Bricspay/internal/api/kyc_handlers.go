package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bricspay/internal/auth"
)

type KYCSubmitRequest struct {
	FullName      string `json:"full_name"`
	DocumentType  string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	Country       string `json:"country"`
}

type KYCReviewRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"` // approved, rejected
	Notes  string `json:"notes"`
}

func (s *Server) HandleKYCSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	principal, ok := auth.FromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req KYCSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if req.FullName == "" || req.DocumentType == "" || req.DocumentNumber == "" || req.Country == "" {
		writeErr(w, http.StatusBadRequest, "missing_required_fields")
		return
	}

	query := `
		INSERT INTO kyc_documents (user_id, full_name, document_type, document_number, country, status, submitted_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW(), NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET full_name = $2, document_type = $3, document_number = $4, country = $5, status = 'pending', updated_at = NOW();
	`
	_, err := s.DB.Exec(query, principal.UserID, req.FullName, req.DocumentType, req.DocumentNumber, req.Country)
	if err != nil {
		log.Printf("Error submitting KYC: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "pending",
		"message": "kyc_submitted_successfully",
	})
}

func (s *Server) HandleKYCStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	principal, ok := auth.FromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var status, fullName, country string
	var submittedAt time.Time
	query := `SELECT status, full_name, country, submitted_at FROM kyc_documents WHERE user_id = $1`
	err := s.DB.QueryRow(query, principal.UserID).Scan(&status, &fullName, &country, &submittedAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "not_submitted",
		})
		return
	} else if err != nil {
		log.Printf("Error querying KYC: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       status,
		"full_name":    fullName,
		"country":      country,
		"submitted_at": submittedAt,
	})
}

func (s *Server) HandleAdminPendingKYC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	rows, err := s.DB.Query(`SELECT id, user_id, full_name, document_type, document_number, country, status, submitted_at FROM kyc_documents WHERE status = 'pending' ORDER BY submitted_at DESC`)
	if err != nil {
		log.Printf("Error querying pending KYC: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, userID int64
		var fullName, docType, docNum, country, status string
		var submittedAt time.Time
		if err := rows.Scan(&id, &userID, &fullName, &docType, &docNum, &country, &status, &submittedAt); err == nil {
			list = append(list, map[string]interface{}{
				"id":              id,
				"user_id":         userID,
				"full_name":       fullName,
				"document_type":   docType,
				"document_number": docNum,
				"country":         country,
				"status":          status,
				"submitted_at":    submittedAt,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": list})
}

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

	if req.Status != "approved" && req.Status != "rejected" {
		writeErr(w, http.StatusBadRequest, "invalid_status")
		return
	}

	_, err := s.DB.Exec(`UPDATE kyc_documents SET status = $1, notes = $2, updated_at = NOW() WHERE id = $3`, req.Status, req.Notes, req.ID)
	if err != nil {
		log.Printf("Error updating KYC review: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "kyc_status_updated"})
}
