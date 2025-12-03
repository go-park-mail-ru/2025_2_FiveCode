package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/gateway_service/internal/config"
	"backend/gateway_service/internal/middleware/mock"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	cfg := &config.Config{
		Cors: config.CorsConfig{
			AllowedOrigins: []string{"http://example.com"},
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(next, cfg)

	t.Run("Allowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("OPTIONS Request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mock.NewMockSessionValidator(ctrl)
	middleware := AuthMiddleware(mockValidator)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, uint64(123), userID)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session"})
		w := httptest.NewRecorder()

		mockValidator.EXPECT().
			ValidateSession(gomock.Any(), "valid_session").
			Return(uint64(123), nil)

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("No Cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid_session"})
		w := httptest.NewRecorder()

		mockValidator.EXPECT().
			ValidateSession(gomock.Any(), "invalid_session").
			Return(uint64(0), errors.New("invalid"))

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserAccessMiddleware(t *testing.T) {
	middleware := UserAccessMiddleware()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "123"})
		ctx := WithUserID(req.Context(), 123)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req.WithContext(ctx))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/456", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "456"})
		ctx := WithUserID(req.Context(), 123)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req.WithContext(ctx))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "123"})
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, 123)
	id, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, uint64(123), id)
}

func TestAccessLogMiddleware(t *testing.T) {
	middleware := AccessLogMiddleware

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func generateToken(sessionID string, secretKey []byte, ttl time.Duration) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	plaintext := make([]byte, 32)
	copy(plaintext[0:16], []byte(sessionID))
	binary.BigEndian.PutUint64(plaintext[16:24], uint64(time.Now().Unix()))

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func TestCSRFMiddleware(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	cfg := &config.Config{
		CSRF: config.CSRFConfig{
			SecretKey:       secretKey,
			TokenTTLMinutes: 60,
		},
	}
	middleware := CSRFMiddleware(cfg)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	sessionID := "session-id"
	validToken, _ := generateToken(sessionID, []byte(secretKey), 60*time.Minute)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		req.Header.Set("X-CSRF-Token", validToken)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Skip GET", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		req.Header.Set("X-CSRF-Token", "invalid")
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
