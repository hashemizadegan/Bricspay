package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	maxKYCDocumentSize = 10 << 20 // 10 MiB
	maxKYCRequestSize  = maxKYCDocumentSize + 1<<20
)

func (s *Server) HandleKYCUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	userID, ok := principalUserID64(w, r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxKYCRequestSize)

	if err := r.ParseMultipartForm(maxKYCRequestSize); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_multipart_form")
		return
	}

	documentType := strings.TrimSpace(r.FormValue("document_type"))
	if documentType == "" {
		writeErr(w, http.StatusBadRequest, "document_type_required")
		return
	}

	if len(documentType) > 100 {
		writeErr(w, http.StatusBadRequest, "document_type_too_long")
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		file, header, err = r.FormFile("file")
	}
	if err != nil {
		writeErr(w, http.StatusBadRequest, "document_file_required")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxKYCDocumentSize+1))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "document_read_failed")
		return
	}

	if len(content) == 0 {
		writeErr(w, http.StatusBadRequest, "document_file_empty")
		return
	}

	if len(content) > maxKYCDocumentSize {
		writeErr(w, http.StatusRequestEntityTooLarge, "document_file_too_large")
		return
	}

	contentType := http.DetectContentType(content)
	switch contentType {
	case "application/pdf", "image/jpeg", "image/png":
	default:
		writeErr(w, http.StatusUnsupportedMediaType, "unsupported_document_type")
		return
	}

	documentID, err := newKYCDocumentID()
	if err != nil {
		log.Printf("kyc upload id error: %v", err)
		writeErr(w, http.StatusInternalServerError, "document_id_error")
		return
	}

	digest := sha256.Sum256(content)
	sha256Hex := hex.EncodeToString(digest[:])
	originalName := strings.TrimSpace(header.Filename)

	const query = `
INSERT INTO kyc_document_blobs (
    id,
    user_id,
    document_type,
    original_name,
    content_type,
    size_bytes,
    sha256,
    content
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

	_, err = s.DB.Exec(
		query,
		documentID,
		userID,
		documentType,
		originalName,
		contentType,
		len(content),
		sha256Hex,
		content,
	)
	if err != nil {
		log.Printf("kyc document upload database error: %v", err)
		writeErr(w, http.StatusInternalServerError, "database_error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":            documentID,
		"document_type": documentType,
		"file_name":     originalName,
		"content_type":  contentType,
		"size_bytes":    len(content),
		"sha256":        sha256Hex,
		"status":        "uploaded",
	})
}

func newKYCDocumentID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

var _ *sql.DB
