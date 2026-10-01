package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateNonce creates a secure cryptographically random hex nonce.
func GenerateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// BuildEIP191Message builds the deterministic personal sign challenge string.
func BuildEIP191Message(domain, address, nonce, issuedAt, expiration string) string {
	return fmt.Sprintf("%s wants you to sign in with your Ethereum account:
%s

URI: https://%s
Nonce: %s
Issued At: %s
Expiration Time: %s",
		domain, strings.ToLower(address), domain, nonce, issuedAt, expiration)
}

// RecoverSigner recovers the Ethereum address (0x...) from an EIP-191 personal_sign signature.
func RecoverSigner(message string, sigHex string) (string, error) {
	sigBytes, err := hexutil.Decode(sigHex)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(sigBytes))
	}

	// Normalize Ethereum signature V parameter (27/28 -> 0/1)
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	} else if sigBytes[64] > 1 {
		return "", errors.New("invalid signature V recovery byte")
	}

	// Personal sign prefix: "Ethereum Signed Message:
" + len(message) + message
	msgPrefixed := fmt.Sprintf("Ethereum Signed Message:
%d%s", len(message), message)
	hash := crypto.Keccak256([]byte(msgPrefixed))

	pubKey, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		return "", fmt.Errorf("crypto.SigToPub recovery failed: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return strings.ToLower(recoveredAddr.Hex()), nil
}
