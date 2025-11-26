package delivery

import (
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/notes/delivery/mock"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/websocket"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNotesDelivery_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	notes := []models.Note{{ID: 1, Title: "Test Note"}}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetAllNotes(gomock.Any(), userID).Return(notes, nil)

		delivery.GetAllNotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		rr := httptest.NewRecorder()

		delivery.GetAllNotes(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetAllNotes(gomock.Any(), userID).Return(nil, errors.New("usecase error"))

		delivery.GetAllNotes(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	note := &models.Note{ID: 1, Title: "New Note"}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(reqBody))
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateNote(gomock.Any(), userID, nil).Return(note, nil)

		delivery.CreateNote(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(reqBody))
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateNote(gomock.Any(), userID, nil).Return(nil, errors.New("usecase error"))

		delivery.CreateNote(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	note := &models.Note{ID: noteID, Title: "Test Note"}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.GetNoteById(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(nil, errors.New("usecase error"))

		delivery.GetNoteById(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	title := "Updated Title"
	input := &models.UpdateNoteInput{ID: noteID, UserID: userID, Title: &title}
	note := &models.Note{ID: noteID, Title: title, UpdatedAt: time.Now()}
	blocks := []models.Block{}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"title": title,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateNote(gomock.Any(), input).Return(note, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return(blocks, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.UpdateNote(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"title": title,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateNote(gomock.Any(), input).Return(nil, errors.New("usecase error"))

		delivery.UpdateNote(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteNote(gomock.Any(), userID, noteID).Return(nil)

		delivery.DeleteNote(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteNote(gomock.Any(), userID, noteID).Return(errors.New("usecase error"))

		delivery.DeleteNote(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_AddFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/favorite", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().AddFavorite(gomock.Any(), userID, noteID).Return(nil)

		delivery.AddFavorite(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/favorite", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().AddFavorite(gomock.Any(), userID, noteID).Return(errors.New("usecase error"))

		delivery.AddFavorite(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_RemoveFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/favorite", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().RemoveFavorite(gomock.Any(), userID, noteID).Return(nil)

		delivery.RemoveFavorite(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/favorite", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().RemoveFavorite(gomock.Any(), userID, noteID).Return(errors.New("usecase error"))

		delivery.RemoveFavorite(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
