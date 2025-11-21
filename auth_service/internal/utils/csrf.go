package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrInvalidSecretKey = errors.New("secret key must be 32 bytes for AES-256")
)

func GenerateCSRFToken(sessionID string, secretKey []byte) (string, error) {
	if len(secretKey) != 32 {
		return "", ErrInvalidSecretKey
	}

	sessionIDBytes := make([]byte, 16)
	copy(sessionIDBytes, []byte(sessionID))

	timestamp := time.Now().Unix()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))

	nonce := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	plaintext := make([]byte, 32)
	copy(plaintext[0:16], sessionIDBytes)
	copy(plaintext[16:24], timestampBytes)
	copy(plaintext[24:32], nonce)

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create gcm: %w", err)
	}

	gcmNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, gcmNonce); err != nil {
		return "", fmt.Errorf("failed to generate gcm nonce: %w", err)
	}

	ciphertext := gcm.Seal(gcmNonce, gcmNonce, plaintext, nil)

	token := base64.URLEncoding.EncodeToString(ciphertext)
	return token, nil
}
