package apiutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ErrorResponseWithCode struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

func WriteError(w http.ResponseWriter, code int, errorText string) {
	WriteJSON(w, code, ErrorResponse{Error: errorText})
}

func WriteErrorWithCode(w http.ResponseWriter, code int, errorText string, errorCode string) {
	WriteJSON(w, code, ErrorResponseWithCode{
		Error:     errorText,
		ErrorCode: errorCode,
	})
}

func WriteValidationErrors(w http.ResponseWriter, code int, errs []FieldError) {
	WriteJSON(w, code, ValidationErrors{Errors: errs})
}

func WriteValidationError(w http.ResponseWriter, code int, err error) {
	if ge, ok := err.(govalidator.Errors); ok {
		out := make([]FieldError, 0, len(ge))
		for _, e := range ge.Errors() {
			field, msg := parseGovalidatorError(e.Error())
			out = append(out, FieldError{
				Field:   field,
				Message: msg,
			})
		}
		WriteValidationErrors(w, code, out)
		return
	}

	WriteError(w, code, err.Error())
}

func parseGovalidatorError(s string) (field, message string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", strings.TrimSpace(s)
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("json encode error")
	}
}

func GetBool(val *bool, def bool) bool {
	if val != nil {
		return *val
	}
	return def
}

func GetString(val *string, def string) string {
	if val != nil {
		return *val
	}
	return def
}

func GetInt(val *int, def int) int {
	if val != nil {
		return *val
	}
	return def
}

func StrictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func HandleGrpcError(w http.ResponseWriter, err error, log zerolog.Logger) {
	st, ok := status.FromError(err)
	if !ok {
		log.Error().Err(err).Msg("unhandled non-grpc error type")
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	switch st.Code() {
	case codes.NotFound:
		WriteError(w, http.StatusNotFound, st.Message())
	case codes.InvalidArgument:
		WriteError(w, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		WriteError(w, http.StatusUnauthorized, st.Message())
	case codes.AlreadyExists:
		WriteError(w, http.StatusConflict, st.Message())
	case codes.PermissionDenied:
		WriteError(w, http.StatusForbidden, st.Message())
	default:
		log.Error().Err(err).Str("grpc_code", st.Code().String()).Msg("unhandled gRPC error")
		WriteError(w, http.StatusInternalServerError, "an unexpected error occurred")
	}
}
