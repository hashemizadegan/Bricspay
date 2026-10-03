package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"bricspay/internal/auth"
)

type NonceRequest struct {
	WalletAddress string `json:"wallet_address"`
}

type NonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

type VerifyRequest struct {
	WalletAddress string `json:"wallet_address"`
	Signature     string `json:"signature"`
	Nonce         string `json:"nonce"`
}

type VerifyResponse struct {
	Token         string `json:"GAPGPTMASKTOKENlzswexhfolrX0X"`
	WalletAddress string `json:"wallet_address"`
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	Status        string `json:"status"`
}

// HandleWalletNonce handles requests for a cryptographic nonce
func (s *Server) HandleWalletNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req NonceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	req.WalletAddress = strings.TrimSpace(req.WalletAddress)
	if req.WalletAddress == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "wallet_address_required"})
		return
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "nonce_generation_failed"})
		return
	}
	nonce := hex.EncodeToString(bytes)
	message := fmt.Sprintf("Sign this message to authenticate with BRICS Pay: %s", nonce)

	expiresAt := time.Now().Add(10 * time.Minute)

	query := `
		INSERT INTO wallet_nonces (wallet_address, nonce, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (wallet_address)
		DO UPDATE SET nonce = $2, expires_at = $3, created_at = NOW();
	`
	_, err := s.DB.Exec(query, strings.ToLower(req.WalletAddress), nonce, expiresAt)
	if err != nil {
		log.Printf("DB error saving nonce: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}

	writeJSON(w, http.StatusOK, NonceResponse{
		Nonce:   nonce,
		Message: message,
	})
}

// HandleWalletVerify verifies the signed nonce and issues a JWT GAPGPTMASKTOKENlzswexhfolrX1X
func (s *Server) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	req.WalletAddress = strings.TrimSpace(strings.ToLower(req.WalletAddress))
	req.Signature = strings.TrimSpace(req.Signature)
	req.Nonce = strings.TrimSpace(req.Nonce)

	if req.WalletAddress == "" || req.Signature == "" || req.Nonce == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	var dbNonce string
	var expiresAt time.Time
	query := `SELECT nonce, expires_at FROM wallet_nonces WHERE wallet_address = $1`
	err := s.DB.QueryRow(query, req.WalletAddress).Scan(&dbNonce, &expiresAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "nonce_not_found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}

	if dbNonce != req.Nonce {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_nonce"})
		return
	}

	if time.Now().After(expiresAt) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "nonce_expired"})
		return
	}

	msg := fmt.Sprintf("Sign this message to authenticate with BRICS Pay: %s", req.Nonce)
	prefixedMsg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	msgHash := crypto.Keccak256([]byte(prefixedMsg))

	sigHex := strings.TrimPrefix(req.Signature, "0x")
	sig, err := hexutil.Decode("0x" + sigHex)
	if err != nil || len(sig) != 65 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_signature_format"})
		return
	}

	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	pubKey, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "signature_recovery_failed"})
		return
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey).Hex()
	if strings.ToLower(recoveredAddr) != req.WalletAddress {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "signature_verification_failed"})
		return
	}

	_, _ = s.DB.Exec(`DELETE FROM wallet_nonces WHERE wallet_address = $1`, req.WalletAddress)

	var userID int64
	var userRole string
	userQuery := `SELECT id, role FROM users WHERE wallet_address = $1`
	err = s.DB.QueryRow(userQuery, req.WalletAddress).Scan(&userID, &userRole)

	if err == sql.ErrNoRows {
		insertUser := `
			INSERT INTO users (wallet_address, role, created_at, updated_at)
			VALUES ($1, 'user', NOW(), NOW())
			RETURNING id, role
		`
		err = s.DB.QueryRow(insertUser, req.WalletAddress).Scan(&userID, &userRole)
		if err != nil {
			log.Printf("DB error creating user: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_create_user"})
			return
		}
	} else if err != nil {
		log.Printf("DB error querying user: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}

	tokenStr, err := auth.IssueToken(userID, userRole)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token_issue_failed"})
		return
	}

	writeJSON(w, http.StatusOK, VerifyResponse{
		Token:         tokenStr,
		WalletAddress: req.WalletAddress,
		UserID:        userID,
		Role:          userRole,
		Status:        "ok",
	})
}
