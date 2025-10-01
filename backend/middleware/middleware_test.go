package middleware

import (
	"backend/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	s := store.NewStore()
	user, err := s.CreateUser("test@example.com", "password")
	require.NoError(t, err)

	sessionID := s.CreateSession(user.ID)

	t.Run("valid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		rr := httptest.NewRecorder()

		handler := AuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			require.True(t, ok)
			require.Equal(t, user.ID, userID)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("no session cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler := AuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
		rr := httptest.NewRecorder()

		handler := AuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestUserAccessMiddleware(t *testing.T) {
	s := store.NewStore()
	user, err := s.CreateUser("test@example.com", "password")
	require.NoError(t, err)

	t.Run("valid access", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/user/{user_id}/notes", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}).Methods("GET")
		
		req := httptest.NewRequest("GET", "/user/1/notes", nil)
		req = req.WithContext(WithUserID(req.Context(), user.ID))
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
		handler := UserAccessMiddleware()(router)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("no user in context", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/user/{user_id}/notes", func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}).Methods("GET")
		
		req := httptest.NewRequest("GET", "/user/1/notes", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
		rr := httptest.NewRecorder()

		handler := UserAccessMiddleware()(router)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("access denied", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/user/{user_id}/notes", func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}).Methods("GET")
		
		req := httptest.NewRequest("GET", "/user/2/notes", nil)
		req = req.WithContext(WithUserID(req.Context(), user.ID))
		req = mux.SetURLVars(req, map[string]string{"user_id": "2"})
		rr := httptest.NewRecorder()

		handler := UserAccessMiddleware()(router)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("missing user id", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/user/{user_id}/notes", func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}).Methods("GET")
		
		req := httptest.NewRequest("GET", "/user//notes", nil)
		req = req.WithContext(WithUserID(req.Context(), user.ID))
		req = mux.SetURLVars(req, map[string]string{"user_id": ""})
		rr := httptest.NewRecorder()

		handler := UserAccessMiddleware()(router)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/user/{user_id}/notes", func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}).Methods("GET")
		
		req := httptest.NewRequest("GET", "/user/invalid/notes", nil)
		req = req.WithContext(WithUserID(req.Context(), user.ID))
		req = mux.SetURLVars(req, map[string]string{"user_id": "invalid"})
		rr := httptest.NewRecorder()

		handler := UserAccessMiddleware()(router)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
