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

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNotesDelivery_CreateBlock_Text(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{ID: 100, NoteID: noteID}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": models.BlockTypeText,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateTextBlock(gomock.Any(), gomock.Any()).Return(block, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{*block}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/notes/invalid/blocks", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidType", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": "unknown",
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": models.BlockTypeText,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateTextBlock(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_CreateBlock_Code(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{ID: 100, NoteID: noteID, Type: models.BlockTypeCode}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": models.BlockTypeCode,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateCodeBlock(gomock.Any(), gomock.Any()).Return(block, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{*block}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": models.BlockTypeCode,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateCodeBlock(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_CreateBlock_Attachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	fileID := uint64(123)
	block := &models.Block{ID: 100, NoteID: noteID, Type: models.BlockTypeAttachment}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type":    models.BlockTypeAttachment,
			"file_id": fileID,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateAttachmentBlock(gomock.Any(), gomock.Any()).Return(block, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{*block}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("MissingFileID", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": models.BlockTypeAttachment,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type":    models.BlockTypeAttachment,
			"file_id": fileID,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/blocks", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().CreateAttachmentBlock(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.CreateBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	noteID := uint64(10)
	blocks := []models.Block{{ID: 100}}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/blocks", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return(blocks, nil)

		delivery.GetBlocks(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes/invalid/blocks", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.GetBlocks(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/blocks", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.GetBlocks(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/blocks", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return(nil, errors.New("usecase error"))

		delivery.GetBlocks(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	blockID := uint64(100)
	block := &models.Block{ID: blockID}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlock(gomock.Any(), userID, blockID).Return(block, nil)

		delivery.GetBlock(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("InvalidBlockID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/blocks/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.GetBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		rr := httptest.NewRecorder()

		delivery.GetBlock(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlock(gomock.Any(), userID, blockID).Return(nil, errors.New("usecase error"))

		delivery.GetBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	blockID := uint64(100)
	noteID := uint64(10)
	block := &models.Block{ID: blockID, NoteID: noteID}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": "text",
			"content": map[string]interface{}{
				"text": "updated text",
			},
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateBlock(gomock.Any(), userID, gomock.Any()).Return(block, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{*block}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("InvalidBlockID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, "/blocks/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidTextContent", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type":    "text",
			"content": 123,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidCodeContent", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type":    "code",
			"content": 123,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": "unknown",
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"type": "text",
			"content": map[string]interface{}{
				"text": "updated text",
			},
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateBlock(gomock.Any(), userID, gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.UpdateBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	blockID := uint64(100)
	noteID := uint64(10)
	block := &models.Block{ID: blockID, NoteID: noteID}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlock(gomock.Any(), userID, blockID).Return(block, nil)
		mockUsecase.EXPECT().DeleteBlock(gomock.Any(), userID, blockID).Return(nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.DeleteBlock(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("InvalidBlockID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/blocks/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.DeleteBlock(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		rr := httptest.NewRecorder()

		delivery.DeleteBlock(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("GetBlockError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlock(gomock.Any(), userID, blockID).Return(nil, errors.New("usecase error"))

		delivery.DeleteBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("DeleteBlockError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/blocks/%d", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetBlock(gomock.Any(), userID, blockID).Return(block, nil)
		mockUsecase.EXPECT().DeleteBlock(gomock.Any(), userID, blockID).Return(errors.New("delete error"))

		delivery.DeleteBlock(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	blockID := uint64(100)
	noteID := uint64(10)
	beforeBlockID := uint64(101)
	block := &models.Block{ID: blockID, NoteID: noteID}
	note := &models.Note{ID: noteID, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"before_block_id": beforeBlockID,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d/position", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateBlockPosition(gomock.Any(), userID, blockID, &beforeBlockID).Return(block, nil)
		mockUsecase.EXPECT().GetBlocks(gomock.Any(), userID, noteID).Return([]models.Block{*block}, nil)
		mockUsecase.EXPECT().GetNoteById(gomock.Any(), userID, noteID).Return(note, nil)

		delivery.UpdateBlockPosition(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("InvalidBlockID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, "/blocks/invalid/position", nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.UpdateBlockPosition(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d/position", blockID), nil)
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		rr := httptest.NewRecorder()

		delivery.UpdateBlockPosition(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d/position", blockID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateBlockPosition(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"before_block_id": beforeBlockID,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/blocks/%d/position", blockID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"block_id": fmt.Sprintf("%d", blockID)})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateBlockPosition(gomock.Any(), userID, blockID, &beforeBlockID).Return(nil, errors.New("usecase error"))

		delivery.UpdateBlockPosition(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
