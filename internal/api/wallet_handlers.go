package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"bricspay/internal/auth"
)

type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address"`
}

type ChallengeResponse struct {
	Nonce      string `json:"nonce"`
	Domain     string `json:"domain"`
	IssuedAt   string `json:"issued_at"`
	ExpiresAt  string `json:"expires_at"`
	Message    string `json:"message"`
	WalletAddr string `json:"wallet_address"`
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

// HandleWalletChallenge creates a nonce and the message the wallet must sign.
func (s *Server) HandleWalletChallenge(w http.ResponseWriter, r *http.Request) {
	var req ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	walletAddress := strings.ToLower(strings.TrimSpace(req.WalletAddress))
	if !common.IsHexAddress(walletAddress) ||
		!strings.HasPrefix(walletAddress, "0x") {
		writeErr(w, http.StatusBadRequest, "invalid_wallet_address")
		return
	}

	domain := walletRequestDomain(r)

	nonce, err := auth.GenerateNonce()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "nonce_generation_failed")
		return
	}

	issuedAt := time.Now().UTC().Truncate(time.Second)
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

	_, err = s.DB.ExecContext(
		r.Context(),
		`
		INSERT INTO wallet_challenges
			(address, nonce, expires_at, consumed, created_at)
		VALUES
			($1, $2, $3, FALSE, $4)
		`,
		walletAddress,
		nonce,
		expiresAt,
		issuedAt,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "challenge_storage_failed")
		return
	}

	writeJSON(w, http.StatusOK, ChallengeResponse{
		Nonce:      nonce,
		Domain:     domain,
		IssuedAt:   issuedAtString,
		ExpiresAt:  expiresAtString,
		Message:    message,
		WalletAddr: walletAddress,
	})
}

// HandleWalletVerify verifies the signature and issues a JWT.
func (s *Server) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	walletAddress := strings.ToLower(strings.TrimSpace(req.WalletAddress))
	nonce := strings.TrimSpace(req.Nonce)
	signature := strings.TrimSpace(req.Signature)

	if !common.IsHexAddress(walletAddress) ||
		!strings.HasPrefix(walletAddress, "0x") {
		writeErr(w, http.StatusBadRequest, "invalid_wallet_address")
		return
	}

	if nonce == "" || signature == "" {
		writeErr(
			w,
			http.StatusBadRequest,
			"wallet_address_nonce_and_signature_are_required",
		)
		return
	}

	var issuedAt time.Time
	var expiresAt time.Time
	var consumed bool

	err := s.DB.QueryRowContext(
		r.Context(),
		`
		SELECT created_at, expires_at, consumed
		FROM wallet_challenges
		WHERE address = $1
		  AND nonce = $2
		ORDER BY created_at DESC
		LIMIT 1
		`,
		walletAddress,
		nonce,
	).Scan(&issuedAt, &expiresAt, &consumed)

	if err != nil {
		if err == sql.ErrNoRows {
			writeErr(w, http.StatusUnauthorized, "challenge_not_found")
			return
		}

		writeErr(w, http.StatusInternalServerError, "challenge_lookup_failed")
		return
	}

	if consumed {
		writeErr(w, http.StatusUnauthorized, "challenge_already_used")
		return
	}

	if time.Now().UTC().After(expiresAt.UTC()) {
		writeErr(w, http.StatusUnauthorized, "challenge_expired")
		return
	}

	expectedMessage := auth.BuildEIP191Message(
		walletRequestDomain(r),
		walletAddress,
		nonce,
		issuedAt.UTC().Format(time.RFC3339),
		expiresAt.UTC().Format(time.RFC3339),
	)

	recoveredAddress, err := auth.RecoverSigner(expectedMessage, signature)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_wallet_signature")
		return
	}

	if strings.ToLower(recoveredAddress) != walletAddress {
		writeErr(w, http.StatusUnauthorized, "wallet_address_mismatch")
		return
	}

	result, err := s.DB.ExecContext(
		r.Context(),
		`
		UPDATE wallet_challenges
		SET consumed = TRUE
		WHERE address = $1
		  AND nonce = $2
		  AND consumed = FALSE
		  AND expires_at > $3
		`,
		walletAddress,
		nonce,
		time.Now().UTC(),
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "challenge_update_failed")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "challenge_update_failed")
		return
	}
	if rowsAffected != 1 {
		writeErr(w, http.StatusUnauthorized, "challenge_invalid_or_already_used")
		return
	}

	userID, userRole, err := walletFindOrCreateUser(
		r.Context(),
		s.DB,
		walletAddress,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "user_lookup_or_creation_failed")
		return
	}

	// userID از نوع int64 است و IssueToken در نسخهٔ فعلی string می‌گیرد.
	userIDStr := strconv.FormatInt(userID, 10)

	token, err := auth.IssueToken(userIDStr, userRole)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token_issue_failed")
		return
	}

	writeJSON(w, http.StatusOK, VerifyResponse{
		Token:         token,
		WalletAddress: walletAddress,
		UserID:        userIDStr,
		Role:          userRole,
		Status:        "ok",
	})
}

func walletRequestDomain(r *http.Request) string {
	domain := strings.TrimSpace(r.Host)
	if domain == "" {
		return "bricspay.local"
	}
	return domain
}

func walletFindOrCreateUser(
	ctx context.Context,
	db *sql.DB,
	walletAddress string,
) (int64, string, error) {
	userID, userRole, err := walletLookupUser(ctx, db, walletAddress)
	if err == nil {
		return userID, userRole, nil
	}
	if err != sql.ErrNoRows {
		return 0, "", err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT u.id, u.role
		FROM wallet_accounts AS wa
		JOIN users AS u ON u.id = wa.user_id
		WHERE wa.address = $1
		LIMIT 1
		`,
		walletAddress,
	).Scan(&userID, &userRole)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return 0, "", err
		}
		if userRole == "" {
			userRole = "user"
		}
		return userID, userRole, nil
	}
	if err != sql.ErrNoRows {
		return 0, "", err
	}

	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO users (role)
		VALUES ('user')
		RETURNING id, role
		`,
	).Scan(&userID, &userRole)
	if err != nil {
		return 0, "", err
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO wallet_accounts (user_id, address)
		VALUES ($1, $2)
		`,
		userID,
		walletAddress,
	)
	if err != nil {
		_ = tx.Rollback()

		existingID, existingRole, lookupErr := walletLookupUser(
			ctx,
			db,
			walletAddress,
		)
		if lookupErr == nil {
			return existingID, existingRole, nil
		}

		return 0, "", err
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}

	if userRole == "" {
		userRole = "user"
	}
	return userID, userRole, nil
}

func walletLookupUser(
	ctx context.Context,
	db *sql.DB,
	walletAddress string,
) (int64, string, error) {
	var userID int64
	var userRole string

	err := db.QueryRowContext(
		ctx,
		`
		SELECT u.id, u.role
		FROM wallet_accounts AS wa
		JOIN users AS u ON u.id = wa.user_id
		WHERE wa.address = $1
		LIMIT 1
		`,
		walletAddress,
	).Scan(&userID, &userRole)

	if err != nil {
		return 0, "", err
	}
	if userRole == "" {
		userRole = "user"
	}

	return userID, userRole, nil
}
