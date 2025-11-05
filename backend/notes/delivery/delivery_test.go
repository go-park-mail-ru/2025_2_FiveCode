package delivery

import (
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"backend/notes/mock"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func setupContext(t *testing.T, userID uint64) context.Context {
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	return middleware.WithUserID(ctx, userID)
}

func TestNotesDelivery_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetAllNotes(gomock.Any(), uint64(1)).Return([]models.Note{
		{ID: 1, OwnerID: 1, Title: "Note 1"},
	}, nil)

	req := httptest.NewRequest("GET", "/notes", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.GetAllNotes(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNotesDelivery_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().CreateNote(gomock.Any(), uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		Title:     "New Note",
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest("POST", "/notes", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.CreateNote(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestNotesDelivery_CreateNote_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().CreateNote(gomock.Any(), uint64(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest("POST", "/notes", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.CreateNote(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNotesDelivery_CreateNote_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	// No user ID in context

	req := httptest.NewRequest("POST", "/notes", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.CreateNote(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNotesDelivery_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetNoteById(gomock.Any(), uint64(1), uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		Title:     "Test Note",
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest("GET", "/notes/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetNoteById(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNotesDelivery_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateNoteRequest{
		Title: stringPtr("Updated Title"),
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateNote(gomock.Any(), uint64(1), uint64(1), stringPtr("Updated Title"), nil).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		Title:     "Updated Title",
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest("PUT", "/notes/1", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateNote(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNotesDelivery_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().DeleteNote(gomock.Any(), uint64(1), uint64(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/notes/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.DeleteNote(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNotesDelivery_GetAllNotes_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetAllNotes(gomock.Any(), uint64(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/notes", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.GetAllNotes(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNotesDelivery_GetNoteById_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("GET", "/notes/invalid", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
	rr := httptest.NewRecorder()

	delivery.GetNoteById(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestNotesDelivery_UpdateNote_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("PUT", "/notes/1", bytes.NewBufferString("invalid json"))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateNote(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestNotesDelivery_UpdateNote_NoFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateNoteRequest{}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/notes/1", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateNote(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func stringPtr(s string) *string {
	return &s
}
