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

	// Use whole seconds so the timestamps stored by Postgres can be used
	// to reconstruct the exact message during verification.
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

	// v11's wallet_challenges table uses "address"; it does not have
	// wallet_address, message, or domain columns.
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

	// v11 stores created_at and expires_at, not the message itself.
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

	// Rebuild the exact message using the stored timestamps. The challenge
	// and verification requests must use the same host/domain.
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

	// Atomically mark the challenge as used. This prevents the same
	// signature from being accepted more than once.
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

	token, err := auth.IssueToken(userID, userRole)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token_issue_failed")
		return
	}

	writeJSON(w, http.StatusOK, VerifyResponse{
		Token:         token,
		WalletAddress: walletAddress,
		UserID:        strconv.FormatInt(userID, 10),
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
	// First check whether this wallet is already linked to a user.
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

	// Check again inside the transaction in case another request linked
	// the wallet after the first lookup.
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

	// The v11 users table has a role column; wallet addresses are stored
	// separately in wallet_accounts.
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

		// If another request won a race to register this unique address,
		// return that account instead of failing.
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
