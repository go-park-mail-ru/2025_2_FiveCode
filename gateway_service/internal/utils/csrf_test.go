package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateToken(sessionID string, secretKey []byte, timestamp int64) string {
	sessionIDBytes := make([]byte, 16)
	copy(sessionIDBytes, []byte(sessionID))

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))

	nonce := make([]byte, 8)
	io.ReadFull(rand.Reader, nonce)

	plaintext := make([]byte, 32)
	copy(plaintext[0:16], sessionIDBytes)
	copy(plaintext[16:24], timestampBytes)
	copy(plaintext[24:32], nonce)

	block, _ := aes.NewCipher(secretKey)
	gcm, _ := cipher.NewGCM(block)

	gcmNonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, gcmNonce)

	ciphertext := gcm.Seal(gcmNonce, gcmNonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext)
}

func TestValidateCSRFToken_Full(t *testing.T) {
	secretKey := []byte("12345678901234567890123456789012")
	sessionID := "session_123"

	t.Run("Valid Token", func(t *testing.T) {
		token := generateToken(sessionID, secretKey, time.Now().Unix())
		err := ValidateCSRFToken(token, sessionID, secretKey, 60)
		assert.NoError(t, err)
	})

	t.Run("Expired Token", func(t *testing.T) {
		expiredTime := time.Now().Add(-61 * time.Minute).Unix()
		token := generateToken(sessionID, secretKey, expiredTime)
		err := ValidateCSRFToken(token, sessionID, secretKey, 60)
		assert.Equal(t, ErrTokenExpired, err)
	})

	t.Run("Session Mismatch", func(t *testing.T) {
		token := generateToken("other_session", secretKey, time.Now().Unix())
		err := ValidateCSRFToken(token, sessionID, secretKey, 60)
		assert.Equal(t, ErrSessionMismatch, err)
	})
}
