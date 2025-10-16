package apiutils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error string `json:"error"`
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
