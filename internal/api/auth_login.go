package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"bricspay/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleLogin authenticates a user by email/password and issues a JWT.
func HandleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req loginRequest
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || req.Password == "" {
			writeJSONError(w, http.StatusBadRequest, "email and password are required")
			return
		}

		if db == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}

		var (
			userIDStr    string
			passwordHash string
		)
		err = db.QueryRow(
			"SELECT id, password_hash FROM bricspay_users WHERE email = $1",
			strings.ToLower(req.Email),
		).Scan(&userIDStr, &passwordHash)
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil && auth.CheckPassword(passwordHash, req.Password) != nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		token, err := auth.IssueToken(userIDStr, "user")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "login successful",
			"token":   token,
			"user_id": userIDStr,
		})
	}
}
