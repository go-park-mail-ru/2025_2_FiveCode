package apiutils

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// НОВОЕ: Ошибка с кодом
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

func TransformMinioURL(internalURL string) string {
	if internalURL == "" {
		return ""
	}

	internalEndpoint := viper.GetString("MINIO_ENDPOINT")
	publicEndpoint := viper.GetString("MINIO_PUBLIC_ENDPOINT")

	if internalEndpoint == "" || publicEndpoint == "" {
		return internalURL
	}

	url := internalURL

	normalizedInternal := strings.Replace(internalEndpoint, "http://", "", 1)
	normalizedInternal = strings.Replace(normalizedInternal, "https://", "", 1)

	normalizedPublic := strings.Replace(publicEndpoint, "http://", "", 1)
	normalizedPublic = strings.Replace(normalizedPublic, "https://", "", 1)

	url = strings.Replace(url, normalizedInternal, normalizedPublic, 1)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(publicEndpoint, "https://") {
			url = "https://" + url
		} else {
			url = "http://" + url
		}
	}

	return url
}
