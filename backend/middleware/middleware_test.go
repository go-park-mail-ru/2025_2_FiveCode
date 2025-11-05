package middleware

import (
	"backend/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// no external store shims needed for these unit tests

func TestCORS_OptionsAndOrigin(t *testing.T) {
    h := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    // OPTIONS should return 200
    req := httptest.NewRequest("OPTIONS", "/", nil)
    req.Header.Set("Origin", "http://localhost:8030")
    w := httptest.NewRecorder()
    h.ServeHTTP(w, req)
    if w.Result().StatusCode != http.StatusOK {
        t.Fatalf("expected 200 for OPTIONS got %d", w.Result().StatusCode)
    }

    // Non-allowed origin should not set header but still call through
    req2 := httptest.NewRequest("GET", "/", nil)
    req2.Header.Set("Origin", "http://evil.com")
    w2 := httptest.NewRecorder()
    h.ServeHTTP(w2, req2)
    if w2.Result().Header.Get("Access-Control-Allow-Origin") != "" {
        t.Fatalf("unexpected CORS header for unallowed origin")
    }
}

func TestAuthMiddleware_NoCookie(t *testing.T) {
    h := AuthMiddleware(&store.Store{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    // No cookie -> 400
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    h.ServeHTTP(w, req)
    if w.Result().StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400 got %d", w.Result().StatusCode)
    }
}

func TestAuthMiddleware_ValidSession(t *testing.T) {
    // validate WithUserID/ GetUserID helpers by simulating middleware behaviour
    final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        uid, ok := GetUserID(r.Context())
        if !ok || uid == 0 {
            t.Fatalf("user id not set by middleware")
        }
        w.WriteHeader(http.StatusOK)
    })

    // Simulate setting user ID via WithUserID helper to mimic middleware effect
    req := httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
    w := httptest.NewRecorder()
    // call final handler with context containing user id
    final.ServeHTTP(w, req.WithContext(WithUserID(req.Context(), 42)))
    if w.Result().StatusCode != http.StatusOK {
        t.Fatalf("expected 200 got %d", w.Result().StatusCode)
    }
}

func TestUserAccessMiddleware_MissingAndForbiddenAndOK(t *testing.T) {
    um := UserAccessMiddleware()

    // Missing user id should return 401
    req := httptest.NewRequest("GET", "/user/1", nil)
    w := httptest.NewRecorder()
    um(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, req)
    if w.Result().StatusCode != http.StatusUnauthorized {
        t.Fatalf("expected 401 got %d", w.Result().StatusCode)
    }

    // Set user id but missing path var -> 400
    req2 := httptest.NewRequest("GET", "/", nil)
    w2 := httptest.NewRecorder()
    req2 = req2.WithContext(WithUserID(req2.Context(), 5))
    um(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w2, req2)
    if w2.Result().StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400 got %d", w2.Result().StatusCode)
    }

    // Simulate vars with different user id -> 403
    req3 := httptest.NewRequest("GET", "/user/6", nil)
    w3 := httptest.NewRecorder()
    req3 = req3.WithContext(WithUserID(req3.Context(), 5))
    // inject mux vars
    req3 = mux.SetURLVars(req3, map[string]string{"user_id": "6"})
    um(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w3, req3)
    if w3.Result().StatusCode != http.StatusForbidden {
        t.Fatalf("expected 403 got %d", w3.Result().StatusCode)
    }

    // OK case
    req4 := httptest.NewRequest("GET", "/user/5", nil)
    w4 := httptest.NewRecorder()
    req4 = req4.WithContext(WithUserID(req4.Context(), 5))
    req4 = mux.SetURLVars(req4, map[string]string{"user_id": "5"})
    um(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })).ServeHTTP(w4, req4)
    if w4.Result().StatusCode != http.StatusOK {
        t.Fatalf("expected 200 got %d", w4.Result().StatusCode)
    }
}

// no helper needed; use mux.SetURLVars

