package apiutils

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"net/http"
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
	var ves validator.ValidationErrors
	ok := errors.As(err, &ves)
	if !ok {
		WriteError(w, code, err.Error())
	}

	out := make([]FieldError, 0, len(ves))
	for _, e := range ves {
		msg := humanizeError(e)
		out = append(out, FieldError{
			Field:   e.Field(),
			Message: msg,
		})
	}

	WriteValidationErrors(w, code, out)
}

func humanizeError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "minimum length is " + e.Param()
	default:
		return e.Error()
	}
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("json encode error")
	}
}
