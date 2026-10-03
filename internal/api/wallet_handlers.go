package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bricspay/internal/auth"
)

type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address"`
}

type ChallengeResponse struct {
	Nonce       string `json:"nonce"`
	Domain      string `json:"domain"`
	IssuedAt    string `json:"issued_at"`
	ExpiresAt   string `json:"expires_at"`
	Message     string `json:"message"`
	WalletAddr  string `json:"wallet_address"`
}

type VerifyRequest struct {
	WalletAddress string `json:"wallet_address"`
	Nonce         string `json:"nonce"`
	Signature     string `json:"signature"`
}

type VerifyResponse struct {
	Token         string `json:"token"`
	WalletAddress string `json:"wallet_address"`
	UserID        string `json:"user_id"`
	Role          string `json:"role"`
	Status        string `json:"status"`
}

// HandleWalletChallenge creates a nonce and the exact message
// that the wallet must sign.
func (s *Server) HandleWalletChallenge(w http.ResponseWriter, r *http.Request) {
	var req ChallengeRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		walletWriteError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		walletWriteError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	walletAddress := strings.ToLower(strings.TrimSpace(req.WalletAddress))

	if len(walletAddress) != 42 ||
		!strings.HasPrefix(walletAddress, "0x") {
		walletWriteError(w, http.StatusBadRequest, "invalid_wallet_address")
		return
	}

	domain := strings.TrimSpace(r.Host)
	if domain == "" {
		domain = "bricspay.local"
	}

	nonce, err := auth.GenerateNonce()
	if err != nil {
		walletWriteError(w, http.StatusInternalServerError, "nonce_generation_failed")
		return
	}

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(5 * time.Minute)

	issuedAtString := issuedAt.Format(time.RFC3339)
	expiresAtString := expiresAt.Format(time.RFC3339)

	message := auth.BuildEIP191Message(
		domain,
		walletAddress,
		nonce,
		issuedAtString,
		expiresAtString,
	)

	_, err = s.db.ExecContext(
		r.Context(),
		`
		INSERT INTO wallet_challenges
			(wallet_address, nonce, domain, message, expires_at, consumed)
		VALUES
			($1, $2, $3, $4, $5, FALSE)
		`,
		walletAddress,
		nonce,
		domain,
		message,
		expiresAt,
	)
	if err != nil {
		walletWriteError(w, http.StatusInternalServerError, "challenge_storage_failed")
		return
	}

	walletWriteJSON(w, http.StatusOK, ChallengeResponse{
		Nonce:      nonce,
		Domain:     domain,
		IssuedAt:   issuedAtString,
		ExpiresAt:  expiresAtString,
		Message:    message,
		WalletAddr: walletAddress,
	})
}

// HandleWalletVerify verifies the signature against the exact message
// previously stored in wallet_challenges.
func (s *Server) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		walletWriteError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		walletWriteError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	walletAddress := strings.ToLower(strings.TrimSpace(req.WalletAddress))
	nonce := strings.TrimSpace(req.Nonce)
	signature := strings.TrimSpace(req.Signature)

	if len(walletAddress) != 42 ||
		!strings.HasPrefix(walletAddress, "0x") {
		walletWriteError(w, http.StatusBadRequest, "invalid_wallet_address")
		return
	}

	if nonce == "" || signature == "" {
		walletWriteError(
			w,
			http.StatusBadRequest,
			"wallet_address_nonce_and_signature_are_required",
		)
		return
	}

	// Retrieve the exact message created during the challenge step.
	var expectedMessage string
	var expiresAt time.Time
	var consumed bool

	err = s.db.QueryRowContext(
		r.Context(),
		`
		SELECT message, expires_at, consumed
		FROM wallet_challenges
		WHERE wallet_address = $1
		  AND nonce = $2
		ORDER BY expires_at DESC
		LIMIT 1
		`,
		walletAddress,
		nonce,
	).Scan(
		&expectedMessage,
		&expiresAt,
		&consumed,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			walletWriteError(w, http.StatusUnauthorized, "challenge_not_found")
			return
		}

		walletWriteError(
			w,
			http.StatusInternalServerError,
			"challenge_lookup_failed",
		)
		return
	}

	if consumed {
		walletWriteError(w, http.StatusUnauthorized, "challenge_already_used")
		return
	}

	if time.Now().UTC().After(expiresAt.UTC()) {
		walletWriteError(w, http.StatusUnauthorized, "challenge_expired")
		return
	}

	// Verify against the stored message. Do not rebuild the message here.
	recoveredAddress, err := auth.RecoverSigner(expectedMessage, signature)
	if err != nil {
		walletWriteError(w, http.StatusUnauthorized, "invalid_wallet_signature")
		return
	}

	if strings.ToLower(recoveredAddress) != walletAddress {
		walletWriteError(w, http.StatusUnauthorized, "wallet_address_mismatch")
		return
	}

	// Consume the challenge after successful signature verification.
	_, err = s.db.ExecContext(
		r.Context(),
		`
		UPDATE wallet_challenges
		SET consumed = TRUE
		WHERE wallet_address = $1
		  AND nonce = $2
		  AND consumed = FALSE
		`,
		walletAddress,
		nonce,
	)
	if err != nil {
		walletWriteError(
			w,
			http.StatusInternalServerError,
			"challenge_update_failed",
		)
		return
	}

	// Find the existing user.
	var userID int64
	var userRole string

	err = s.db.QueryRowContext(
		r.Context(),
		`
		SELECT id, role
		FROM users
		WHERE wallet_address = $1
		LIMIT 1
		`,
		walletAddress,
	).Scan(&userID, &userRole)

	// If the wallet has no account, create one.
	if err == sql.ErrNoRows {
		err = s.db.QueryRowContext(
			r.Context(),
			`
			INSERT INTO users
				(wallet_address, role, status)
			VALUES
				($1, 'user', 'APPROVED')
			RETURNING id, role
			`,
			walletAddress,
		).Scan(&userID, &userRole)

		if err != nil {
			walletWriteError(
				w,
				http.StatusInternalServerError,
				"user_creation_failed",
			)
			return
		}
	} else if err != nil {
		walletWriteError(
			w,
			http.StatusInternalServerError,
			"user_lookup_failed",
		)
		return
	}

	if userRole == "" {
		userRole = "user"
	}

	token, err := auth.IssueToken(userID, userRole)
	if err != nil {
		walletWriteError(
			w,
			http.StatusInternalServerError,
			"token_issue_failed",
		)
		return
	}

	walletWriteJSON(w, http.StatusOK, VerifyResponse{
		Token:         token,
		WalletAddress: walletAddress,
		UserID:        strconv.FormatInt(userID, 10),
		Role:          userRole,
		Status:        "ok",
	})
}

func walletWriteJSON(w http.ResponseWriter, statusCode int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(value)
}

func walletWriteError(w http.ResponseWriter, statusCode int, message string) {
	walletWriteJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
