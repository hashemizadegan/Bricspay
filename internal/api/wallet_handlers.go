package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bricspay/internal/auth"
)

type walletNonceResponse struct {
	Nonce string `json:"nonce"`
}

type walletVerifyRequest struct {
	UserID        int64  `json:"user_id"`
	Address       string `json:"address"`
	WalletAddress string `json:"wallet_address"`
	Signature     string `json:"signature"`
	Nonce         string `json:"nonce"`
}

type walletVerifyResponse struct {
	OK      bool   `json:"ok"`
	Token   string `json:"token"`
	Address string `json:"address"`
	Signer  string `json:"signer"`
	UserID  int64  `json:"user_id"`
}

func (s *Server) HandleWalletNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nonce, err := auth.GenerateNonce()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to generate nonce")
		return
	}

	writeJSON(w, http.StatusOK, walletNonceResponse{
		Nonce: nonce,
	})
}

func (s *Server) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req walletVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}

	walletAddress := strings.TrimSpace(req.WalletAddress)
	if walletAddress == "" {
		walletAddress = strings.TrimSpace(req.Address)
	}

	if walletAddress == "" || strings.TrimSpace(req.Signature) == "" || strings.TrimSpace(req.Nonce) == "" {
		writeErr(w, http.StatusBadRequest, "wallet address, signature and nonce are required")
		return
	}

	message := auth.BuildEIP191Message(req.Nonce)

	signer, err := auth.RecoverSigner(message, req.Signature)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid wallet signature")
		return
	}

	if !strings.EqualFold(signer, walletAddress) {
		writeErr(w, http.StatusUnauthorized, "signature does not match wallet address")
		return
	}

	token, err := auth.IssueToken(req.UserID, "user")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	s.logAudit(req.UserID, walletAddress, "WALLET_VERIFY", "wallet_auth", 0, r)

	writeJSON(w, http.StatusOK, walletVerifyResponse{
		OK:      true,
		Token:   token,
		Address: walletAddress,
		Signer:  signer,
		UserID:  req.UserID,
	})
}
