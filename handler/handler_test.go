// tests/auth_handlers_test.go
package handler

import (
	"backend/store"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doRequest(router http.Handler, method, url string, body []byte, cookies []*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, url, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	return responseRecorder
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name     string
		preSetup func(s *store.Store)
		content  interface{}
		wantHTTP int
	}{
		{
			name:     "valid registration",
			content:  map[string]string{"email": "u1@example.com", "password": "secret123"},
			wantHTTP: http.StatusCreated,
		},
		{
			name: "duplicate email",
			preSetup: func(s *store.Store) {
				_, _ = s.CreateUser("dup@example.com", "pass1234")
			},
			content:  map[string]string{"email": "dup@example.com", "password": "pass1234"},
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			content:  "not json",
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "missing email",
			content:  map[string]string{"password": "secret123"},
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "missing password",
			content:  map[string]string{"email": "u1@example.com"},
			wantHTTP: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := store.NewStore()
			if test.preSetup != nil {
				test.preSetup(s)
			}
			router := NewRouter(s)

			var body []byte
			switch contentType := test.content.(type) {
			case string:
				body = []byte(contentType)
			default:
				b, _ := json.Marshal(contentType)
				body = b
			}

			responseRecorder := doRequest(router, "POST", "/api/register", body, nil)
			if responseRecorder.Code != test.wantHTTP {
				t.Fatalf("want status %d, got %d; body: %s", test.wantHTTP, responseRecorder.Code, responseRecorder.Body.String())
			}

			if test.wantHTTP == http.StatusCreated {
				var got map[string]interface{}
				if err := json.Unmarshal(responseRecorder.Body.Bytes(), &got); err != nil {
					t.Fatalf("cannot decode response json: %contentType", err)
				}

				email, ok := got["email"].(string)
				if !ok || email == "" {
					t.Fatalf("expected email in response")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name          string
		preSetup      func(s *store.Store)
		content       interface{}
		wantHTTP      int
		wantSetCookie bool
	}{
		{
			name: "login success",
			preSetup: func(s *store.Store) {
				_, _ = s.CreateUser("login1@example.com", "mypassword")
			},
			content:       map[string]string{"email": "login1@example.com", "password": "mypassword"},
			wantHTTP:      http.StatusOK,
			wantSetCookie: true,
		},
		{
			name: "wrong password",
			preSetup: func(s *store.Store) {
				_, _ = s.CreateUser("login2@example.com", "rightpass")
			},
			content:       map[string]string{"email": "login2@example.com", "password": "wrong"},
			wantHTTP:      http.StatusUnauthorized,
			wantSetCookie: false,
		},
		{
			name: "user not found",
			preSetup: func(s *store.Store) {
				_, _ = s.CreateUser("login1@example.com", "mypassword")
			},
			content:  map[string]string{"email": "noexist@example.com", "password": "mypassword"},
			wantHTTP: http.StatusUnauthorized,
		},
		{
			name:     "invalid json",
			content:  "not json",
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "missing email",
			content:  map[string]string{"password": "secret123"},
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "missing password",
			content:  map[string]string{"email": "u1@example.com"},
			wantHTTP: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := store.NewStore()
			if test.preSetup != nil {
				test.preSetup(s)
			}
			router := NewRouter(s)

			var body []byte
			switch contentType := test.content.(type) {
			case string:
				body = []byte(contentType)
			default:
				b, _ := json.Marshal(contentType)
				body = b
			}

			responseRecorder := doRequest(router, "POST", "/api/login", body, nil)
			if responseRecorder.Code != test.wantHTTP {
				t.Fatalf("want status %d, got %d; body: %s", test.wantHTTP, responseRecorder.Code, responseRecorder.Body.String())
			}
			cookies := responseRecorder.Result().Cookies()
			if test.wantSetCookie && len(cookies) == 0 {
				t.Fatalf("expected cookie to be set on login success")
			}
			if !test.wantSetCookie && len(cookies) > 0 {
				t.Fatalf("did not expect cookie but got one")
			}
		})
	}
}

func TestLogout(t *testing.T) {
	tests := []struct {
		name              string
		setupCookie       func(s *store.Store) *http.Cookie
		wantHTTP          int
		expectSessionGone bool
	}{
		{
			name: "logout with valid session",
			setupCookie: func(s *store.Store) *http.Cookie {
				u, _ := s.CreateUser("lgout@example.com", "pw12345")
				sid := s.CreateSession(u.ID)
				return &http.Cookie{Name: "session_id", Value: sid, Path: "/"}
			},
			wantHTTP:          http.StatusOK,
			expectSessionGone: true,
		},
		{
			name:              "logout without cookie",
			setupCookie:       func(s *store.Store) *http.Cookie { return nil },
			wantHTTP:          http.StatusOK,
			expectSessionGone: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := store.NewStore()
			router := NewRouter(s)

			var cookie *http.Cookie
			if test.setupCookie != nil {
				cookie = test.setupCookie(s)
			}

			cookies := []*http.Cookie{}
			if cookie != nil {
				cookies = append(cookies, cookie)
			}

			responseRecorder := doRequest(router, "POST", "/api/logout", nil, cookies)
			if responseRecorder.Code != test.wantHTTP {
				t.Fatalf("want %d got %d body: %s", test.wantHTTP, responseRecorder.Code, responseRecorder.Body.String())
			}

			if test.expectSessionGone && cookie != nil {
				if _, ok := s.GetUserBySession(cookie.Value); ok {
					t.Fatalf("expected session to be deleted after logout")
				}
			}
		})
	}
}
