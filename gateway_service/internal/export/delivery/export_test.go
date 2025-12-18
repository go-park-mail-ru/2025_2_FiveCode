package delivery

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/gateway_service/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockExportUsecase struct {
	pdf   []byte
	title string
	err   error
}

func (m *mockExportUsecase) ExportNoteToPDF(ctx context.Context, userID, noteID uint64) ([]byte, string, error) {
	return m.pdf, m.title, m.err
}

func TestExportDelivery_ExportNoteToPDF(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pdf := []byte("fake pdf content")
		usecase := &mockExportUsecase{
			pdf:   pdf,
			title: "Test Note",
			err:   nil,
		}
		delivery := NewExportDelivery(usecase)

		req, _ := http.NewRequest("GET", "/export/1", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint64(1))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		delivery.ExportNoteToPDF(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/pdf", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
		assert.Equal(t, pdf, rr.Body.Bytes())
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		usecase := &mockExportUsecase{}
		delivery := NewExportDelivery(usecase)

		req, _ := http.NewRequest("GET", "/export/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint64(1))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		delivery.ExportNoteToPDF(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NoUserID", func(t *testing.T) {
		usecase := &mockExportUsecase{}
		delivery := NewExportDelivery(usecase)

		req, _ := http.NewRequest("GET", "/export/1", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "1"})

		rr := httptest.NewRecorder()
		delivery.ExportNoteToPDF(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		usecase := &mockExportUsecase{
			err: errors.New("usecase error"),
		}
		delivery := NewExportDelivery(usecase)

		req, _ := http.NewRequest("GET", "/export/1", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint64(1))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		delivery.ExportNoteToPDF(rr, req)

		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal", "My Note", "My Note"},
		{"WithSlash", "Note/Test", "Note_Test"},
		{"WithBackslash", "Note\\Test", "Note_Test"},
		{"WithColon", "Note:Test", "Note_Test"},
		{"WithSpecialChars", "Note*?\"<>|", "Note______"},
		{"TooLong", string(make([]byte, 150)), string(make([]byte, 100))},
		{"Empty", "", "note"},
		{"OnlySpecialChars", "///", "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if tt.name == "TooLong" {
				assert.Len(t, result, 100)
			} else if tt.name == "Empty" {
				assert.Equal(t, "note", result)
			} else if tt.name == "WithSpecialChars" {
				assert.Contains(t, result, "Note")
				assert.NotContains(t, result, "*")
				assert.NotContains(t, result, "?")
				assert.NotContains(t, result, "\"")
				assert.NotContains(t, result, "<")
				assert.NotContains(t, result, ">")
				assert.NotContains(t, result, "|")
			} else if tt.name == "OnlySpecialChars" {
				assert.Equal(t, "___", result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
