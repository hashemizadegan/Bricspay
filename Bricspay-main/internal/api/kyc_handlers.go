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

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password string }
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"}); return
	}
	var id, hash, role, status string
	err := s.DB.QueryRow(`SELECT id, password_hash, role, status FROM users WHERE email=$1`,
		strings.ToLower(strings.TrimSpace(req.Email))).Scan(&id, &hash, &role, &status)
	if err != nil || !auth.CheckPassword(hash, req.Password) {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"}); return
	}
	tok, _ := auth.IssueToken(id, req.Email, role)
	writeJSON(w, 200, map[string]string{"token": tok, "role": role, "status": status})
}

// ---------- Document Upload ----------

var allowedTypes = map[string]bool{
	"COMMERCIAL_REGISTRY": true, "STATUTES": true, "PASSPORT_SIGNATORY": true,
	"BANK_LICENSE": true, "TAX_CERTIFICATE": true, "OFAC_SCREENING": true,
}

func (s *Server) HandleKYCUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok { writeJSON(w, 401, map[string]string{"error": "unauthorized"}); return }

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeJSON(w, 400, map[string]string{"error": "multipart error or file > 10MB"}); return
	}
	docType := r.FormValue("document_type")
	if !allowedTypes[docType] {
		writeJSON(w, 400, map[string]string{"error": "invalid document_type"}); return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil { writeJSON(w, 400, map[string]string{"error": "file required"}); return }
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(hdr.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		writeJSON(w, 400, map[string]string{"error": "only pdf/jpg/png allowed"}); return
	}

	// پروفایل کاربر
	var profileID string
	err = s.DB.QueryRow(`SELECT id FROM kyc_profiles WHERE user_id=$1`, claims.UserID).Scan(&profileID)
	if err != nil { writeJSON(w, 404, map[string]string{"error": "profile not found"}); return }

	// هش و ذخیره
	h := sha256.New()
	tmp, _ := os.CreateTemp("", "kyc-*")
	defer os.Remove(tmp.Name())
	size, _ := io.Copy(io.MultiWriter(tmp, h), file)
	tmp.Close()
	sum := hex.EncodeToString(h.Sum(nil))

	dir := envOrStr("UPLOAD_DIR", "/data/uploads")
	os.MkdirAll(dir, 0o700)
	finalPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", profileID, sum[:16], ext))
	if err := os.Rename(tmp.Name(), finalPath); err != nil {
		// fallback: copy
		src, _ := os.Open(tmp.Name()); defer src.Close()
		dst, err2 := os.Create(finalPath)
		if err2 != nil { writeJSON(w, 500, map[string]string{"error": "storage error"}); return }
		io.Copy(dst, src); dst.Close()
	}

	var docID string
	err = s.DB.QueryRow(
		`INSERT INTO kyc_documents (profile_id, document_type, file_path, original_name, checksum_sha256, size_bytes, status)
		 VALUES ($1,$2,$3,$4,$5,$6,'SUBMITTED') RETURNING id`,
		profileID, docType, finalPath, hdr.Filename, sum, size).Scan(&docID)
	if err != nil { writeJSON(w, 500, map[string]string{"error": "db error"}); return }

	s.logAudit(claims.UserID, claims.Email, "UPLOAD_DOCUMENT", "kyc_document", docID, r)
	writeJSON(w, 201, map[string]string{"document_id": docID, "checksum": sum, "status": "SUBMITTED"})
}

// ---------- Admin ----------

func (s *Server) HandleAdminProfiles(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`
		SELECT p.id, u.email, p.legal_name, p.jurisdiction, p.status, p.risk_tier, p.created_at,
		       (SELECT count(*) FROM kyc_documents d WHERE d.profile_id = p.id) AS docs
		FROM kyc_profiles p JOIN users u ON u.id = p.user_id
		ORDER BY p.created_at DESC LIMIT 200`)
	if err != nil { writeJSON(w, 500, map[string]string{"error": "db error"}); return }
	defer rows.Close()
	type item struct {
		ID, Email, LegalName, Jurisdiction, Status, RiskTier string
		Docs int
		CreatedAt string
	}
	list := []item{}
	for rows.Next() {
		var it item
		rows.Scan(&it.ID, &it.Email, &it.LegalName, &it.Jurisdiction, &it.Status, &it.RiskTier, &it.CreatedAt, &it.Docs)
		list = append(list, it)
	}
	writeJSON(w, 200, map[string]interface{}{"profiles": list})
}

func (s *Server) HandleAdminDecision(w http.ResponseWriter, r *http.Request) {
	var req struct{ ProfileID, Decision, Notes string } // Decision: APPROVE | REJECT | UNDER_REVIEW
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"}); return
	}
	if req.Decision != "APPROVE" && req.Decision != "REJECT" && req.Decision != "UNDER_REVIEW" {
		writeJSON(w, 400, map[string]string{"error": "decision must be APPROVE|REJECT|UNDER_REVIEW"}); return
	}
	profileStatus := map[string]string{"APPROVE": "APPROVED", "REJECT": "REJECTED", "UNDER_REVIEW": "UNDER_REVIEW"}[req.Decision]
	var userID string
	err := s.DB.QueryRow(
		`UPDATE kyc_profiles SET status=$1, reviewer_notes=$2 WHERE id=$3 RETURNING user_id`,
		profileStatus, req.Notes, req.ProfileID).Scan(&userID)
	if err == sql.ErrNoRows { writeJSON(w, 404, map[string]string{"error": "profile not found"}); return }
	if err != nil { writeJSON(w, 500, map[string]string{"error": "db error"}); return }

	userStatus := map[string]string{"APPROVED": "APPROVED", "REJECTED": "REJECTED", "UNDER_REVIEW": "PENDING"}[profileStatus]
	s.DB.Exec(`UPDATE users SET status=$1 WHERE id=$2`, userStatus, userID)

	claims, _ := auth.FromContext(r.Context())
	s.logAudit(claims.UserID, claims.Email, "KYC_"+req.Decision, "kyc_profile", req.ProfileID, r)
	writeJSON(w, 200, map[string]string{"profile_id": req.ProfileID, "status": profileStatus})
}

func (s *Server) HandleAdminAudit(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT action, actor_email, target_entity, target_id, ip_address, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT 300`)
	if err != nil { writeJSON(w, 500, map[string]string{"error": "db error"}); return }
	defer rows.Close()
	type entry struct{ Action, Actor, Entity, TargetID, IP, Time string }
	logs := []entry{}
	for rows.Next() {
		var e entry
		rows.Scan(&e.Action, &e.Actor, &e.Entity, &e.TargetID, &e.IP, &e.Time)
		logs = append(logs, e)
	}
	writeJSON(w, 200, map[string]interface{}{"audit": logs})
}

// ---------- helpers ----------

func (s *Server) logAudit(actorID, actorEmail, action, entity, targetID string, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" { ip = r.RemoteAddr }
	s.DB.Exec(`INSERT INTO audit_logs (actor_id, actor_email, action, target_entity, target_id, ip_address)
		VALUES ($1,$2,$3,$4,$5,$6)`, actorID, actorEmail, action, entity, targetID, ip)
}

func envOrStr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" { return v }
	return d
}
