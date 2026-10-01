package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"bricspay/internal/auth"
)

type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address"`
	Domain        string `json:"domain"`
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

// HandleWalletChallenge generates and stores an EIP-191 challenge for MetaMask.
func (s *Server) HandleWalletChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable"})
		return
	}

	var req ChallengeRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.WalletAddress = strings.ToLower(strings.TrimSpace(req.WalletAddress))
	if req.WalletAddress == "" || !strings.HasPrefix(req.WalletAddress, "0x") || len(req.WalletAddress) != 42 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid 0x ethereum wallet address required"})
		return
	}

	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = r.Host
		if domain == "" {
			domain = "bricspay.local"
		}
	}

	nonce, err := auth.GenerateNonce()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate nonce"})
		return
	}

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(5 * time.Minute)
	issuedAtStr := issuedAt.Format(time.RFC3339)
	expiresAtStr := expiresAt.Format(time.RFC3339)

	message := auth.BuildEIP191Message(domain, req.WalletAddress, nonce, issuedAtStr, expiresAtStr)

	_, err = s.DB.ExecContext(r.Context(),
		`INSERT INTO wallet_challenges (wallet_address, nonce, domain, message, expires_at, consumed)
		 VALUES ($1, $2, $3, $4, $5, FALSE)`,
		req.WalletAddress, nonce, domain, message, expiresAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist challenge: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ChallengeResponse{
		Nonce:      nonce,
		Domain:     domain,
		IssuedAt:   issuedAtStr,
		ExpiresAt:  expiresAtStr,
		Message:    message,
		WalletAddr: req.WalletAddress,
	})
}

// HandleWalletVerify verifies the MetaMask EIP-191 signature, consumes the challenge, and returns a JWT.
func (s *Server) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable"})
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.WalletAddress = strings.ToLower(strings.TrimSpace(req.WalletAddress))
	req.Nonce = strings.TrimSpace(req.Nonce)
	req.Signature = strings.TrimSpace(req.Signature)

	if req.WalletAddress == "" || req.Nonce == "" || req.Signature == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "wallet_address, nonce, and signature are required"})
		return
	}

	// 1. Fetch unconsumed challenge
	var challengeID string
	var expectedMsg string
	var expiresAt time.Time
	var consumed bool

	err := s.DB.QueryRowContext(r.Context(),
		`SELECT id, message, expires_at, consumed FROM wallet_challenges
		 WHERE wallet_address = $1 AND nonce = $2`,
		req.WalletAddress, req.Nonce).Scan(&challengeID, &expectedMsg, &expiresAt, &consumed)

	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "challenge not found or invalid"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db query error: " + err.Error()})
		return
	}

	if consumed {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "challenge has already been consumed (replay attack prevention)"})
		return
	}

	if time.Now().UTC().After(expiresAt) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "challenge expired"})
		return
	}

	// 2. Consume challenge immediately (single-use)
	_, _ = s.DB.ExecContext(r.Context(), `UPDATE wallet_challenges SET consumed = TRUE WHERE id = $1`, challengeID)

	// 3. Cryptographically recover signer address
	recoveredAddr, err := auth.RecoverSigner(expectedMsg, req.Signature)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "signature verification failed: " + err.Error()})
		return
	}

	if strings.ToLower(recoveredAddr) != req.WalletAddress {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "signature signer mismatch"})
		return
	}

	// 4. Link or create user and wallet_account
	var userID string
	var userRole string
	var userStatus string

	err = s.DB.QueryRowContext(r.Context(),
		`SELECT u.id, u.role, u.status
		 FROM wallet_accounts wa
		 JOIN users u ON u.id = wa.user_id
		 WHERE wa.wallet_address = $1`,
		req.WalletAddress).Scan(&userID, &userRole, &userStatus)

	if err == sql.ErrNoRows {
		// Auto-provision user account for wallet
		syntheticEmail := req.WalletAddress + "@wallet.bricspay.local"
		dummyPassHash, _ := auth.HashPassword("wallet_auto_provisioned_account_secure_pw")

		err = s.DB.QueryRowContext(r.Context(),
			`INSERT INTO users (email, password_hash, entity_type, role, status)
			 VALUES ($1, $2, 'CORPORATE', 'member', 'APPROVED')
			 ON CONFLICT (email) DO UPDATE SET updated_at = NOW()
			 RETURNING id, role, status`,
			syntheticEmail, dummyPassHash).Scan(&userID, &userRole, &userStatus)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to provision wallet user: " + err.Error()})
			return
		}

		_, err = s.DB.ExecContext(r.Context(),
			`INSERT INTO wallet_accounts (user_id, wallet_address, is_primary)
			 VALUES ($1, $2, TRUE)
			 ON CONFLICT (wallet_address) DO NOTHING`,
			userID, req.WalletAddress)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to link wallet account: " + err.Error()})
			return
		}
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "user lookup error: " + err.Error()})
		return
	}

	// 5. Issue JWT Token
	token, err := auth.IssueToken(userID, req.WalletAddress+"@wallet.bricspay.local", userRole)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate jwt: " + err.Error()})
		return
	}

	s.logAudit(userID, req.WalletAddress, "WALLET_LOGIN", "wallet_account", req.WalletAddress, r)

	writeJSON(w, http.StatusOK, VerifyResponse{
		Token:         token,
		WalletAddress: req.WalletAddress,
		UserID:        userID,
		Role:          userRole,
		Status:        userStatus,
	})
}
