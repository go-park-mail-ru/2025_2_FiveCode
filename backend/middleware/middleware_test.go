package middleware

import (
	"backend/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeAuthMiddleware(t *testing.T) {
	s := store.NewStore()
	user, err := s.CreateUser("test@example.com", "password", "user")
	require.NoError(t, err)

	sessionID := s.CreateSession(user.ID)

	t.Run("valid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		rr := httptest.NewRecorder()

		handler := MakeAuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("no session cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler := MakeAuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
		rr := httptest.NewRecorder()

		handler := MakeAuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
