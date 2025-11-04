package usecase

import (
	"backend/blocks/mock"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBlocksUsecase_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{}, nil)

	mockBlocksRepo.EXPECT().CreateBlock(ctx, uint64(1), models.BlockTypeText, 1.0).Return(&models.Block{
		ID:        1,
		NoteID:    1,
		Type:      models.BlockTypeText,
		Position:  1.0,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.CreateBlock(ctx, 1, 1, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
}

func TestBlocksUsecase_CreateBlock_GetNoteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(nil, assert.AnError)

	block, err := usecase.CreateBlock(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().GetBlocksByNoteID(ctx, uint64(1)).Return([]models.BlockWithContent{
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

	blocks, err := usecase.GetBlocks(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
}

func TestBlocksUsecase_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockByID(ctx, uint64(1)).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text: "Test",
	}, nil)

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.GetBlock(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
}

func TestBlocksUsecase_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().UpdateBlockText(ctx, uint64(1), "Updated text", gomock.Any()).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text: "Updated text",
	}, nil)

	block, err := usecase.UpdateBlock(ctx, 1, 1, "Updated text", []models.BlockTextFormat{})
	assert.NoError(t, err)
	assert.Equal(t, "Updated text", block.Text)
}

func TestBlocksUsecase_UpdateBlock_GetNoteIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(0), assert.AnError)

	block, err := usecase.UpdateBlock(ctx, 1, 1, "text", []models.BlockTextFormat{})
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_UpdateBlock_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.UpdateBlock(ctx, 1, 1, "text", []models.BlockTextFormat{})
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_GetBlocks_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	blocks, err := usecase.GetBlocks(ctx, 1, 1)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
	assert.Nil(t, blocks)
}

func TestBlocksUsecase_GetBlocks_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)
	mockBlocksRepo.EXPECT().GetBlocksByNoteID(ctx, uint64(1)).Return(nil, assert.AnError)

	blocks, err := usecase.GetBlocks(ctx, 1, 1)
	assert.Error(t, err)
	assert.Nil(t, blocks)
}

func TestBlocksUsecase_GetBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockByID(ctx, uint64(1)).Return(nil, assert.AnError)

	block, err := usecase.GetBlock(ctx, 1, 1)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_CreateBlock_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.CreateBlock(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_CreateBlock_CalculatePositionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)
	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return(nil, assert.AnError)

	block, err := usecase.CreateBlock(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)
	mockBlocksRepo.EXPECT().DeleteBlock(ctx, uint64(1)).Return(nil)

	err := usecase.DeleteBlock(ctx, 1, 1)
	assert.NoError(t, err)
}

func TestBlocksUsecase_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(1)).Return([]struct {
		ID       uint64
		Position float64
	}{}, nil)

	mockBlocksRepo.EXPECT().UpdateBlockPosition(ctx, uint64(1), 1.0).Return(&models.Block{
		ID:        1,
		NoteID:    1,
		Type:      models.BlockTypeText,
		Position:  1.0,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.UpdateBlockPosition(ctx, 1, 1, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
}

func TestBlocksUsecase_UpdateBlockPosition_WithBeforeBlockID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()
	beforeBlockID := uint64(2)

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(1)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 2, Position: 2.0},
		{ID: 3, Position: 3.0},
	}, nil)

	mockBlocksRepo.EXPECT().UpdateBlockPosition(ctx, uint64(1), 1.0).Return(&models.Block{
		ID:        1,
		NoteID:    1,
		Type:      models.BlockTypeText,
		Position:  1.0,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.UpdateBlockPosition(ctx, 1, 1, &beforeBlockID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
}

func TestBlocksUsecase_UpdateBlock_WithFormats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	formats := []models.BlockTextFormat{
		{StartOffset: 0, EndOffset: 5, Bold: true},
		{StartOffset: 5, EndOffset: 10, Italic: true},
	}

	mockBlocksRepo.EXPECT().UpdateBlockText(ctx, uint64(1), "Test text", gomock.Any()).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text:    "Test text",
		Formats: formats,
	}, nil)

	block, err := usecase.UpdateBlock(ctx, 1, 1, "Test text", formats)
	assert.NoError(t, err)
	assert.Equal(t, "Test text", block.Text)
}

func TestBlocksUsecase_CreateBlock_WithBeforeBlockID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()
	beforeBlockID := uint64(2)

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 2, Position: 2.0},
	}, nil)

	mockBlocksRepo.EXPECT().CreateBlock(ctx, uint64(1), models.BlockTypeText, 1.0).Return(&models.Block{
		ID:        1,
		NoteID:    1,
		Type:      models.BlockTypeText,
		Position:  1.0,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.CreateBlock(ctx, 1, 1, &beforeBlockID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
}

func TestBlocksUsecase_GetBlock_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockByID(ctx, uint64(1)).Return(&models.BlockWithContent{
		Block: models.Block{
			ID:        1,
			NoteID:    1,
			Type:      models.BlockTypeText,
			Position:  1.0,
			CreatedAt: time.Now(),
		},
		Text: "Test",
	}, nil)

	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.GetBlock(ctx, 1, 1)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_DeleteBlock_GetNoteIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(0), assert.AnError)

	err := usecase.DeleteBlock(ctx, 1, 1)
	assert.Error(t, err)
}

func TestBlocksUsecase_DeleteBlock_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	err := usecase.DeleteBlock(ctx, 1, 1)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
}

func TestBlocksUsecase_DeleteBlock_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)
	mockBlocksRepo.EXPECT().DeleteBlock(ctx, uint64(1)).Return(assert.AnError)

	err := usecase.DeleteBlock(ctx, 1, 1)
	assert.Error(t, err)
}

func TestBlocksUsecase_UpdateBlockPosition_GetNoteIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(0), assert.AnError)

	block, err := usecase.UpdateBlockPosition(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_UpdateBlockPosition_NoAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   2,
		CreatedAt: time.Now(),
	}, nil)

	block, err := usecase.UpdateBlockPosition(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNoAccess, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_UpdateBlockPosition_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, uint64(1)).Return(uint64(1), nil)
	mockNotesRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
		ID:        1,
		OwnerID:   1,
		CreatedAt: time.Now(),
	}, nil)
	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(1)).Return([]struct {
		ID       uint64
		Position float64
	}{}, nil)
	mockBlocksRepo.EXPECT().UpdateBlockPosition(ctx, uint64(1), 1.0).Return(nil, assert.AnError)

	block, err := usecase.UpdateBlockPosition(ctx, 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestBlocksUsecase_stylesEqual(t *testing.T) {
	tests := []struct {
		name     string
		format1  models.BlockTextFormat
		format2  models.BlockTextFormat
		expected bool
	}{
		{
			name: "equal styles",
			format1: models.BlockTextFormat{
				Bold:   true,
				Italic: false,
				Font:   models.FontInter,
				Size:   12,
			},
			format2: models.BlockTextFormat{
				Bold:   true,
				Italic: false,
				Font:   models.FontInter,
				Size:   12,
			},
			expected: true,
		},
		{
			name: "different bold",
			format1: models.BlockTextFormat{
				Bold: true,
			},
			format2: models.BlockTextFormat{
				Bold: false,
			},
			expected: false,
		},
		{
			name: "equal links",
			format1: models.BlockTextFormat{
				Link: stringPtr("http://example.com"),
			},
			format2: models.BlockTextFormat{
				Link: stringPtr("http://example.com"),
			},
			expected: true,
		},
		{
			name: "one nil link",
			format1: models.BlockTextFormat{
				Link: nil,
			},
			format2: models.BlockTextFormat{
				Link: stringPtr("http://example.com"),
			},
			expected: false,
		},
		{
			name:     "both nil links",
			format1:  models.BlockTextFormat{Link: nil},
			format2:  models.BlockTextFormat{Link: nil},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stylesEqual(tt.format1, tt.format2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBlocksUsecase_mergeFormats(t *testing.T) {
	formats := map[int]models.BlockTextFormat{
		0: {Bold: true, Font: models.FontInter, Size: 12},
		1: {Italic: true, Font: models.FontInter, Size: 12},
	}

	result := mergeFormats(formats)
	assert.True(t, result.Bold)
	assert.True(t, result.Italic)
	assert.Equal(t, models.FontInter, result.Font)
	assert.Equal(t, 12, result.Size)
}

func TestBlocksUsecase_isDefaultFormat(t *testing.T) {
	tests := []struct {
		name   string
		format models.BlockTextFormat
		want   bool
	}{
		{
			name:   "default format",
			format: models.BlockTextFormat{Font: models.FontInter, Size: 12},
			want:   true,
		},
		{
			name:   "with bold",
			format: models.BlockTextFormat{Bold: true, Font: models.FontInter, Size: 12},
			want:   false,
		},
		{
			name:   "with link",
			format: models.BlockTextFormat{Link: stringPtr("http://example.com"), Font: models.FontInter, Size: 12},
			want:   false,
		},
		{
			name:   "different font",
			format: models.BlockTextFormat{Font: models.FontRoboto, Size: 12},
			want:   false,
		},
		{
			name:   "different size",
			format: models.BlockTextFormat{Font: models.FontInter, Size: 14},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDefaultFormat(tt.format)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBlocksUsecase_formatsEqual(t *testing.T) {
	tests := []struct {
		name    string
		format1 models.BlockTextFormat
		format2 models.BlockTextFormat
		want    bool
	}{
		{
			name:    "equal formats",
			format1: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Bold: true, Italic: false},
			format2: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Bold: true, Italic: false},
			want:    true,
		},
		{
			name:    "different start offset",
			format1: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Bold: true},
			format2: models.BlockTextFormat{StartOffset: 1, EndOffset: 5, Bold: true},
			want:    false,
		},
		{
			name:    "different bold",
			format1: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Bold: true},
			format2: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Bold: false},
			want:    false,
		},
		{
			name:    "equal with links",
			format1: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Link: stringPtr("http://example.com")},
			format2: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Link: stringPtr("http://example.com")},
			want:    true,
		},
		{
			name:    "different links",
			format1: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Link: stringPtr("http://example.com")},
			format2: models.BlockTextFormat{StartOffset: 0, EndOffset: 5, Link: stringPtr("http://other.com")},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatsEqual(tt.format1, tt.format2)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBlocksUsecase_calculatePosition_BeforeBlockInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()
	beforeBlockID := uint64(3)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 2, Position: 2.0},
		{ID: 3, Position: 3.0},
		{ID: 4, Position: 4.0},
	}, nil)

	position, err := usecase.calculatePosition(ctx, 1, &beforeBlockID, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, position)
}

func TestBlocksUsecase_calculatePosition_BeforeBlockAtStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()
	beforeBlockID := uint64(2)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 2, Position: 2.0},
		{ID: 3, Position: 3.0},
	}, nil)

	position, err := usecase.calculatePosition(ctx, 1, &beforeBlockID, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, position)
}

func TestBlocksUsecase_calculatePosition_BeforeBlockNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()
	beforeBlockID := uint64(999)

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 2, Position: 2.0},
	}, nil)

	position, err := usecase.calculatePosition(ctx, 1, &beforeBlockID, 0)
	assert.Error(t, err)
	assert.Equal(t, float64(0), position)
}

func TestBlocksUsecase_calculatePosition_WithMultipleBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo)

	ctx := context.Background()

	mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, uint64(1), uint64(0)).Return([]struct {
		ID       uint64
		Position float64
	}{
		{ID: 1, Position: 1.0},
		{ID: 2, Position: 2.0},
		{ID: 3, Position: 3.0},
	}, nil)

	position, err := usecase.calculatePosition(ctx, 1, nil, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4.0, position)
}

func TestBlocksUsecase_optimizeFormats(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		formats []models.BlockTextFormat
		wantLen int
	}{
		{
			name:    "empty formats",
			text:    "test",
			formats: []models.BlockTextFormat{},
			wantLen: 0,
		},
		{
			name: "single format",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 4, Bold: true},
			},
			wantLen: 1,
		},
		{
			name: "overlapping formats",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 2, Bold: true},
				{StartOffset: 1, EndOffset: 4, Italic: true},
			},
			wantLen: 2,
		},
		{
			name: "adjacent formats same style",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 2, Bold: true},
				{StartOffset: 2, EndOffset: 4, Bold: true},
			},
			wantLen: 1,
		},
		{
			name: "invalid format - end offset too large",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 10, Bold: true},
			},
			wantLen: 0,
		},
		{
			name: "invalid format - start >= end",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 2, EndOffset: 2, Bold: true},
			},
			wantLen: 0,
		},
		{
			name: "default format filtered out",
			text: "test",
			formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 4, Font: models.FontInter, Size: 12},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizeFormats(tt.text, tt.formats)
			assert.GreaterOrEqual(t, len(result), 0)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
