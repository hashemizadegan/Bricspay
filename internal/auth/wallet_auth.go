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

func GenerateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func BuildEIP191Message(domain, address, nonce, issuedAt, expiration string) string {
	return fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n%s\n\nURI: https://%s\nNonce: %s\nIssued At: %s\nExpiration Time: %s",
		domain, strings.ToLower(address), domain, nonce, issuedAt, expiration)
}

func RecoverSigner(message string, sigHex string) (string, error) {
	sigBytes, err := hexutil.Decode(sigHex)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(sigBytes))
	}
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	} else if sigBytes[64] > 1 {
		return "", errors.New("invalid signature V recovery byte")
	}
	msgPrefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256([]byte(msgPrefixed))
	pubKey, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		return "", fmt.Errorf("crypto.SigToPub recovery failed: %w", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return strings.ToLower(recoveredAddr.Hex()), nil
}
