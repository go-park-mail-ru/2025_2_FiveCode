package apiutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		data     interface{}
		wantCode int
	}{
		{
			name:     "successful response",
			code:     http.StatusOK,
			data:     map[string]string{"message": "success"},
			wantCode: http.StatusOK,
		},
		{
			name:     "error response",
			code:     http.StatusBadRequest,
			data:     map[string]string{"error": "invalid request"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "user object",
			code:     http.StatusCreated,
			data:     map[string]interface{}{"id": 1, "email": "test@example.com"},
			wantCode: http.StatusCreated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteJSON(w, test.code, test.data)
			
			require.Equal(t, test.wantCode, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			var result map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err)
		})
	}
}
