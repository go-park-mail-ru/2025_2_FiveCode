package delivery

import (
	"backend/blocks/mock"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
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

func TestBlocksDelivery_CreateTextBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)
	ctx := setupContext(t, 1)

	body := baseCreateBlockRequest{
		Type:          models.BlockTypeText,
		BeforeBlockID: nil,
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().CreateTextBlock(gomock.Any(), uint64(1), uint64(1), nil).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:       1,
			NoteID:   1,
			Type:     models.BlockTypeText,
			Position: 1.0,
		},
	}, nil)

	req := httptest.NewRequest("POST", "/notes/1/blocks", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.CreateBlock(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestBlocksDelivery_CreateBlock_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	body := baseCreateBlockRequest{Type: models.BlockTypeText}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/notes/1/blocks", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.CreateBlock(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBlocksDelivery_CreateTextBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)
	ctx := setupContext(t, 1)

	body := baseCreateBlockRequest{Type: models.BlockTypeText}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().CreateTextBlock(gomock.Any(), uint64(1), uint64(1), nil).Return(nil, assert.AnError)

	req := httptest.NewRequest("POST", "/notes/1/blocks", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.CreateBlock(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlocks(gomock.Any(), uint64(1), uint64(1)).Return([]models.BlockWithContent{
		{
			Block: models.Block{
				ID:        1,
				NoteID:    1,
				Type:      models.BlockTypeText,
				Position:  1.0,
				CreatedAt: time.Now(),
			},
			Text: "Test",
		},
	}, nil)

	req := httptest.NewRequest("GET", "/notes/1/blocks", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlocks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBlocksDelivery_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlock(gomock.Any(), uint64(1), uint64(1)).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text: "Test",
	}, nil)

	req := httptest.NewRequest("GET", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlock(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBlocksDelivery_GetBlock_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	req := httptest.NewRequest("GET", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlock(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBlocksDelivery_GetBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlock(gomock.Any(), uint64(1), uint64(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlock(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockRequest{
		Text:    "Updated text",
		Formats: []BlockTextFormatInput{},
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateBlock(gomock.Any(), uint64(1), uint64(1), "Updated text", gomock.Any()).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text: "Updated text",
	}, nil)

	req := httptest.NewRequest("PATCH", "/blocks/1", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlock(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBlocksDelivery_UpdateBlock_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	body := UpdateBlockRequest{Text: "text", Formats: []BlockTextFormatInput{}}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PATCH", "/blocks/1", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlock(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBlocksDelivery_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().DeleteBlock(gomock.Any(), uint64(1), uint64(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.DeleteBlock(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestBlocksDelivery_DeleteBlock_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	req := httptest.NewRequest("DELETE", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.DeleteBlock(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBlocksDelivery_DeleteBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().DeleteBlock(gomock.Any(), uint64(1), uint64(1)).Return(assert.AnError)

	req := httptest.NewRequest("DELETE", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.DeleteBlock(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockPositionRequest{
		BeforeBlockID: nil,
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateBlockPosition(gomock.Any(), uint64(1), uint64(1), nil).Return(&models.Block{
		ID:        1,
		NoteID:    1,
		Type:      models.BlockTypeText,
		Position:  2.0,
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest("PUT", "/blocks/1/position", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlockPosition(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBlocksDelivery_GetBlock_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlock(gomock.Any(), uint64(1), uint64(999)).Return(nil, namederrors.ErrNotFound)

	req := httptest.NewRequest("GET", "/blocks/999", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "999"})
	rr := httptest.NewRecorder()

	delivery.GetBlock(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBlocksDelivery_GetBlock_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlock(gomock.Any(), uint64(1), uint64(1)).Return(nil, namederrors.ErrNoAccess)

	req := httptest.NewRequest("GET", "/blocks/1", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlock(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestBlocksDelivery_DeleteBlock_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().DeleteBlock(gomock.Any(), uint64(1), uint64(999)).Return(namederrors.ErrNotFound)

	req := httptest.NewRequest("DELETE", "/blocks/999", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "999"})
	rr := httptest.NewRecorder()

	delivery.DeleteBlock(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBlocksDelivery_CreateBlock_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("POST", "/notes/1/blocks", bytes.NewBufferString("invalid json"))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.CreateBlock(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_GetBlocks_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("GET", "/notes/invalid/blocks", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
	rr := httptest.NewRecorder()

	delivery.GetBlocks(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_GetBlocks_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	mockUsecase.EXPECT().GetBlocks(gomock.Any(), uint64(1), uint64(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/notes/1/blocks", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlocks(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_UpdateBlock_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockRequest{Text: "text"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PATCH", "/blocks/invalid", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlock(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_UpdateBlock_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("PATCH", "/blocks/1", bytes.NewBufferString("invalid json"))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlock(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_UpdateBlockPosition_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockPositionRequest{BeforeBlockID: nil}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/blocks/invalid/position", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "invalid"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlockPosition(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_UpdateBlockPosition_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	req := httptest.NewRequest("PUT", "/blocks/1/position", bytes.NewBufferString("invalid json"))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlockPosition(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBlocksDelivery_UpdateBlockPosition_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockPositionRequest{BeforeBlockID: nil}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateBlockPosition(gomock.Any(), uint64(1), uint64(1), nil).Return(nil, assert.AnError)

	req := httptest.NewRequest("PUT", "/blocks/1/position", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlockPosition(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_UpdateBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := setupContext(t, 1)

	body := UpdateBlockRequest{Text: "text"}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateBlock(gomock.Any(), uint64(1), uint64(1), "text", gomock.Any()).Return(nil, assert.AnError)

	req := httptest.NewRequest("PATCH", "/blocks/1", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"block_id": "1"})
	rr := httptest.NewRecorder()

	delivery.UpdateBlock(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestBlocksDelivery_GetBlocks_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockBlocksUsecase(ctrl)
	delivery := NewBlocksDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	req := httptest.NewRequest("GET", "/notes/1/blocks", nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetBlocks(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBlocksDelivery_convertToFormats(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []BlockTextFormatInput
		wantLen int
	}{
		{
			name:    "empty formats",
			inputs:  []BlockTextFormatInput{},
			wantLen: 0,
		},
		{
			name: "with formats",
			inputs: []BlockTextFormatInput{
				{StartOffset: 0, EndOffset: 5, Bold: boolPtr(true)},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToFormats(tt.inputs)
			assert.Equal(t, tt.wantLen, len(result))
			if len(result) > 0 {
				assert.NotNil(t, result[0].Bold)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}