package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidToken     = errors.New("invalid csrf token")
	ErrTokenExpired     = errors.New("csrf token expired")
	ErrSessionMismatch  = errors.New("session id mismatch")
	ErrInvalidSecretKey = errors.New("secret key must be 32 bytes for AES-256")
)

func ValidateCSRFToken(token string, sessionID string, secretKey []byte, ttlMinutes int) error {
	if len(secretKey) != 32 {
		return ErrInvalidSecretKey
	}

	ciphertext, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("%w: invalid base64", ErrInvalidToken)
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return fmt.Errorf("%w: failed to create cipher", ErrInvalidToken)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("%w: failed to create gcm", ErrInvalidToken)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("%w: token too short", ErrInvalidToken)
	}

	gcmNonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, gcmNonce, encryptedData, nil)
	if err != nil {
		return fmt.Errorf("%w: decryption failed", ErrInvalidToken)
	}

	if len(plaintext) != 32 {
		return fmt.Errorf("%w: invalid plaintext length", ErrInvalidToken)
	}

	tokenSessionIDBytes := plaintext[0:16]
	timestampBytes := plaintext[16:24]

	sessionIDBytes := make([]byte, 16)
	copy(sessionIDBytes, []byte(sessionID))

	sessionIDLen := len(sessionID)
	if sessionIDLen > 16 {
		sessionIDLen = 16
	}

	for i := 0; i < sessionIDLen; i++ {
		if tokenSessionIDBytes[i] != sessionIDBytes[i] {
			return ErrSessionMismatch
		}
	}

	timestamp := int64(binary.BigEndian.Uint64(timestampBytes))
	now := time.Now().Unix()
	age := now - timestamp

	if age < 0 {
		return fmt.Errorf("%w: token from future", ErrInvalidToken)
	}

	if age > int64(ttlMinutes*60) {
		return ErrTokenExpired
	}

	return nil
}
