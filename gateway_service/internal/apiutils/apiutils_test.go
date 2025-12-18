package apiutils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "bad request")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad request"}`, w.Body.String())
}

func TestWriteErrorWithCode(t *testing.T) {
	w := httptest.NewRecorder()
	WriteErrorWithCode(w, http.StatusForbidden, "forbidden", "access_denied")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"forbidden", "error_code":"access_denied"}`, w.Body.String())
}

func TestWriteValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	errs := []FieldError{{Field: "email", Message: "invalid"}}
	WriteValidationErrors(w, http.StatusBadRequest, errs)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"errors":[{"field":"email", "message":"invalid"}]}`, w.Body.String())
}

func TestWriteValidationError(t *testing.T) {
	t.Run("SimpleError", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("simple error")
		WriteValidationError(w, http.StatusBadRequest, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"simple error"}`, w.Body.String())
	})

	t.Run("GovalidatorErrors", func(t *testing.T) {
		w := httptest.NewRecorder()
		errs := govalidator.Errors{
			errors.New("email: invalid format"),
			errors.New("password: too short"),
		}
		WriteValidationError(w, http.StatusBadRequest, errs)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "email")
		assert.Contains(t, w.Body.String(), "invalid format")
		assert.Contains(t, w.Body.String(), "password")
		assert.Contains(t, w.Body.String(), "too short")
	})
}

func TestGetters(t *testing.T) {
	t.Run("GetBool", func(t *testing.T) {
		val := true
		assert.True(t, GetBool(&val, false))
		assert.False(t, GetBool(nil, false))
	})

	t.Run("GetString", func(t *testing.T) {
		val := "test"
		assert.Equal(t, "test", GetString(&val, "default"))
		assert.Equal(t, "default", GetString(nil, "default"))
	})

	t.Run("GetInt", func(t *testing.T) {
		val := 123
		assert.Equal(t, 123, GetInt(&val, 456))
		assert.Equal(t, 456, GetInt(nil, 456))
	})
}

func TestStrictUnmarshal(t *testing.T) {
	type TestStruct struct {
		Field string `json:"field"`
	}

	t.Run("Success", func(t *testing.T) {
		var s TestStruct
		err := StrictUnmarshal([]byte(`{"field":"value"}`), &s)
		assert.NoError(t, err)
		assert.Equal(t, "value", s.Field)
	})

	t.Run("Unknown Field", func(t *testing.T) {
		var s TestStruct
		err := StrictUnmarshal([]byte(`{"field":"value", "unknown":"value"}`), &s)
		assert.Error(t, err)
	})
}

func TestHandleGrpcError(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"NotFound", status.Error(codes.NotFound, "not found"), http.StatusNotFound, "not found"},
		{"InvalidArgument", status.Error(codes.InvalidArgument, "bad args"), http.StatusBadRequest, "bad args"},
		{"Unauthenticated", status.Error(codes.Unauthenticated, "no auth"), http.StatusUnauthorized, "no auth"},
		{"AlreadyExists", status.Error(codes.AlreadyExists, "exists"), http.StatusConflict, "exists"},
		{"PermissionDenied", status.Error(codes.PermissionDenied, "denied"), http.StatusForbidden, "denied"},
		{"Internal", status.Error(codes.Internal, "internal"), http.StatusInternalServerError, "an unexpected error occurred"},
		{"Unimplemented", status.Error(codes.Unimplemented, "not implemented"), http.StatusInternalServerError, "an unexpected error occurred"},
		{"Unavailable", status.Error(codes.Unavailable, "unavailable"), http.StatusInternalServerError, "an unexpected error occurred"},
		{"NonGrpc", errors.New("std error"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleGrpcError(w, tt.err, logger)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestParseGovalidatorError(t *testing.T) {
	t.Run("WithField", func(t *testing.T) {
		field, msg := parseGovalidatorError("email: invalid format")
		assert.Equal(t, "email", field)
		assert.Equal(t, "invalid format", msg)
	})

	t.Run("WithoutField", func(t *testing.T) {
		field, msg := parseGovalidatorError("generic error")
		assert.Equal(t, "", field)
		assert.Equal(t, "generic error", msg)
	})

	t.Run("WithSpaces", func(t *testing.T) {
		field, msg := parseGovalidatorError("  email  :  invalid format  ")
		assert.Equal(t, "email", field)
		assert.Equal(t, "invalid format", msg)
	})
}

func TestWriteJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"key": "value"}
		WriteJSON(w, http.StatusOK, data)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "key")
		assert.Contains(t, w.Body.String(), "value")
	})

	t.Run("WithStruct", func(t *testing.T) {
		w := httptest.NewRecorder()
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		data := TestStruct{Name: "test", Value: 123}
		WriteJSON(w, http.StatusCreated, data)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "test")
		assert.Contains(t, w.Body.String(), "123")
	})
}
