package api

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

	// سقف حجم مجاز: ۱۰ مگابایت
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "file_too_large_or_invalid_form")
		return
	}

	docType := strings.TrimSpace(r.FormValue("document_type"))
	if docType == "" || len(docType) > 64 {
		writeErr(w, http.StatusBadRequest, "invalid_document_type")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "file_missing")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil || len(content) == 0 {
		writeErr(w, http.StatusBadRequest, "empty_file_or_read_error")
		return
	}

	// ارزیابی نوع فایل از روی بایت‌های ابتدایی
	sniffBytes := content
	if len(sniffBytes) > 512 {
		sniffBytes = sniffBytes[:512]
	}
	detectedType := http.DetectContentType(sniffBytes)
	switch detectedType {
	case "application/pdf", "image/jpeg", "image/png":
	default:
		writeErr(w, http.StatusBadRequest, "unsupported_media_type")
		return
	}

	hasher := sha256.New()
	hasher.Write(content)
	sha256Hex := hex.EncodeToString(hasher.Sum(nil))

	blobID := uuid.New().String()
	var createdAt time.Time

	query := `
		INSERT INTO kyc_document_blobs (id, user_id, document_type, original_name, content_type, size_bytes, sha256, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`
	err = s.DB.QueryRowContext(r.Context(), query,
		blobID, userID, docType, header.Filename, detectedType, int64(len(content)), sha256Hex, content,
	).Scan(&createdAt)

	if err != nil {
		writeErr(w, http.StatusInternalServerError, "storage_failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":            blobID,
		"document_type": docType,
		"content_type":  detectedType,
		"size_bytes":    len(content),
		"sha256":        sha256Hex,
		"created_at":    createdAt,
	})
}
