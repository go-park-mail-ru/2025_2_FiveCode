package utils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken(t *testing.T) {
	secretKey := []byte("12345678901234567890123456789012")
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		token, err := GenerateCSRFToken(sessionID, secretKey)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		_, err = base64.URLEncoding.DecodeString(token)
		assert.NoError(t, err)
	})

	t.Run("Invalid Secret Key", func(t *testing.T) {
		_, err := GenerateCSRFToken(sessionID, []byte("short"))
		assert.Equal(t, ErrInvalidSecretKey, err)
	})
}
