package router

import (
	"backend/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	s := store.NewStore()
	router := NewRouter(s)
	require.NotNil(t, router, "router should not be nil")

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{
			name:   "register endpoint exists",
			method: "POST",
			path:   "/api/register",
		},
		{
			name:   "login endpoint exists",
			method: "POST",
			path:   "/api/login",
		},
		{
			name:   "logout endpoint exists",
			method: "POST",
			path:   "/api/logout",
		},
		{
			name:     "notes endpoint requires auth",
			method:   "GET",
			path:     "/api/user/1/notes",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "non-existent endpoint returns 404",
			method:   "GET",
			path:     "/api/none",
			wantCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if test.wantCode == 0 {
				require.NotEqual(t, http.StatusNotFound, rr.Code, "endpoint should exist")
			} else {
				require.Equal(t, test.wantCode, rr.Code, "should return expected status code")
			}
		})
	}
}
