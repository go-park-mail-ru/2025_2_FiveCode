package apiutils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	t.Run("GovaleditorErrors", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("simple error")
		WriteValidationError(w, http.StatusBadRequest, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"simple error"}`, w.Body.String())
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
	}{
		{"NotFound", status.Error(codes.NotFound, "not found"), http.StatusNotFound},
		{"InvalidArgument", status.Error(codes.InvalidArgument, "bad args"), http.StatusBadRequest},
		{"Unauthenticated", status.Error(codes.Unauthenticated, "no auth"), http.StatusUnauthorized},
		{"AlreadyExists", status.Error(codes.AlreadyExists, "exists"), http.StatusConflict},
		{"PermissionDenied", status.Error(codes.PermissionDenied, "denied"), http.StatusForbidden},
		{"Internal", status.Error(codes.Internal, "internal"), http.StatusInternalServerError},
		{"NonGrpc", errors.New("std error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleGrpcError(w, tt.err, logger)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestParseGovalidatorError(t *testing.T) {
	field, msg := parseGovalidatorError("email: invalid format")
	assert.Equal(t, "email", field)
	assert.Equal(t, "invalid format", msg)

	field, msg = parseGovalidatorError("generic error")
	assert.Equal(t, "", field)
	assert.Equal(t, "generic error", msg)
}
