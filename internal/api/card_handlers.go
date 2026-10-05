package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bricspay/internal/auth"
)

type bankCard struct {
	ID          int64  `json:"id"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand,omitempty"`
	HolderName  string `json:"holder_name,omitempty"`
	ExpiryMonth int    `json:"expiry_month,omitempty"`
	ExpiryYear  int    `json:"expiry_year,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Preferred   bool   `json:"preferred"`
}

func (s *Server) requireApprovedKYC(userID int64) (email, phone string, ok bool, err error) {
	var status string
	err = s.DB.QueryRow(`
SELECT u.email, u.phone, COALESCE(p.status::text,'PENDING')
  FROM users u
  LEFT JOIN kyc_profiles p ON p.user_id = u.id
 WHERE u.id = $1`, userID).Scan(&email, &phone, &status)
	if err != nil {
		return "", "", false, err
	}
	ok = status == "APPROVED" && strings.TrimSpace(email) != "" && strings.TrimSpace(phone) != ""
	return email, phone, ok, nil
}

// principalUserID64 شناسهٔ کاربر را از context می‌خواند و به int64 تبدیل می‌کند.
// چون Principal.UserID از نوع string است، تبدیل در مرز handler انجام می‌شود.
func principalUserID64(w http.ResponseWriter, r *http.Request) (int64, bool) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return 0, false
	}
	uid, err := strconv.ParseInt(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil || uid <= 0 {
		http.Error(w, `{"error":"invalid_user_id"}`, http.StatusUnauthorized)
		return 0, false
	}
	return uid, true
}

func (s *Server) HandleCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := principalUserID64(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.listCards(w, userID)
	case http.MethodPost:
		s.addCard(w, r, userID)
	default:
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleCardItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := principalUserID64(w, r)
	if !ok {
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/cards/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"bad_id"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"bad_id"}`, http.StatusBadRequest)
		return
	}

	if len(parts) == 2 && parts[1] == "preferred" {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		s.setPreferred(w, userID, id)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		s.renameCard(w, r, userID, id)
	case http.MethodDelete:
		s.removeCard(w, userID, id)
	default:
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) listCards(w http.ResponseWriter, userID int64) {
	rows, err := s.DB.Query(`
SELECT id, last4, COALESCE(brand,''), COALESCE(holder_name,''),
       COALESCE(expiry_month,0), COALESCE(expiry_year,0),
       COALESCE(display_name,''), is_preferred
  FROM bank_cards WHERE user_id = $1 ORDER BY is_preferred DESC, id`, userID)
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := make([]bankCard, 0, 3)
	for rows.Next() {
		var c bankCard
		if err := rows.Scan(&c.ID, &c.Last4, &c.Brand, &c.HolderName, &c.ExpiryMonth, &c.ExpiryYear, &c.DisplayName, &c.Preferred); err != nil {
			http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"cards": out, "max": 3})
}

func (s *Server) addCard(w http.ResponseWriter, r *http.Request, userID int64) {
	_, _, eligible, err := s.requireApprovedKYC(userID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"kyc_required"}`, http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	if !eligible {
		http.Error(w, `{"error":"kyc_or_profile_incomplete"}`, http.StatusForbidden)
		return
	}
	var req struct {
		Last4       string `json:"last4"`
		Brand       string `json:"brand"`
		HolderName  string `json:"holder_name"`
		ExpiryMonth int    `json:"expiry_month"`
		ExpiryYear  int    `json:"expiry_year"`
		DisplayName string `json:"display_name"`
		TokenRef    string `json:"token_ref"`
		Preferred   bool   `json:"preferred"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	req.Last4 = strings.TrimSpace(req.Last4)
	if len(req.Last4) != 4 {
		http.Error(w, `{"error":"invalid_last4"}`, http.StatusBadRequest)
		return
	}
	var n int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM bank_cards WHERE user_id=$1`, userID).Scan(&n); err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	if n >= 3 {
		http.Error(w, `{"error":"card_limit_exceeded"}`, http.StatusConflict)
		return
	}
	var id int64
	err = s.DB.QueryRow(`
INSERT INTO bank_cards (user_id, last4, brand, holder_name, expiry_month, expiry_year, display_name, token_ref, is_preferred)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		userID, req.Last4, req.Brand, req.HolderName, req.ExpiryMonth, req.ExpiryYear, req.DisplayName, req.TokenRef, req.Preferred || n == 0,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "card_limit_exceeded") {
			http.Error(w, `{"error":"card_limit_exceeded"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) setPreferred(w http.ResponseWriter, userID, cardID int64) {
	res, err := s.DB.Exec(`UPDATE bank_cards SET is_preferred = TRUE, updated_at = NOW() WHERE id=$1 AND user_id=$2`, cardID, userID)
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) renameCard(w http.ResponseWriter, r *http.Request, userID, cardID int64) {
	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	res, err := s.DB.Exec(`UPDATE bank_cards SET display_name=$1, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
		strings.TrimSpace(req.DisplayName), cardID, userID)
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) removeCard(w http.ResponseWriter, userID, cardID int64) {
	res, err := s.DB.Exec(`DELETE FROM bank_cards WHERE id=$1 AND user_id=$2`, cardID, userID)
	if err != nil {
		http.Error(w, `{"error":"db"}`, http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
