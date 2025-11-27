package usecase

import (
	"backend/notes_service/blocks/mock"
	"backend/notes_service/blocks/repository"
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlocksUsecase_CreateTextBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{
		BaseBlock: models.BaseBlock{ID: 100, NoteID: noteID, Type: "text", Position: 1.0},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, noteID, uint64(0)).Return([]repository.BlockPositionInfo{}, nil)
		mockBlocksRepo.EXPECT().CreateTextBlock(ctx, noteID, 1.0, userID).Return(block, nil)

		res, err := usecase.CreateTextBlock(ctx, userID, noteID, nil)
		assert.NoError(t, err)
		assert.Equal(t, block, res)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: false}, nil)

		_, err := usecase.CreateTextBlock(ctx, userID, noteID, nil)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
	})
}

func TestBlocksUsecase_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blocks := []models.Block{
		{BaseBlock: models.BaseBlock{ID: 1, NoteID: noteID}},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)
		mockBlocksRepo.EXPECT().GetBlocksByNoteID(ctx, noteID).Return(blocks, nil)

		res, err := usecase.GetBlocks(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, blocks, res)
	})
}

func TestBlocksUsecase_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: blockID, NoteID: noteID}}

	t.Run("Success", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockByID(ctx, blockID).Return(block, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)

		res, err := usecase.GetBlock(ctx, userID, blockID)
		assert.NoError(t, err)
		assert.Equal(t, block, res)
	})
}

func TestBlocksUsecase_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, blockID).Return(noteID, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		mockBlocksRepo.EXPECT().DeleteBlock(ctx, blockID).Return(nil)

		err := usecase.DeleteBlock(ctx, userID, blockID)
		assert.NoError(t, err)
	})
}

func TestBlocksUsecase_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: blockID, NoteID: noteID, Position: 1.5}}

	t.Run("Success", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, blockID).Return(noteID, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, noteID, blockID).Return([]repository.BlockPositionInfo{
			{ID: 1, Position: 1.0},
			{ID: 2, Position: 2.0},
		}, nil)
		mockBlocksRepo.EXPECT().UpdateBlockPosition(ctx, blockID, 3.0).Return(block, nil)

		res, err := usecase.UpdateBlockPosition(ctx, userID, blockID, nil)
		assert.NoError(t, err)
		assert.Equal(t, block, res)
	})
}

func TestBlocksUsecase_CreateAttachmentBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	fileID := uint64(123)
	block := &models.Block{
		BaseBlock: models.BaseBlock{ID: 100, NoteID: noteID, Type: models.BlockTypeAttachment, Position: 1.0},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, noteID, uint64(0)).Return([]repository.BlockPositionInfo{}, nil)
		mockBlocksRepo.EXPECT().CreateAttachmentBlock(ctx, noteID, 1.0, fileID, userID).Return(block, nil)

		res, err := usecase.CreateAttachmentBlock(ctx, userID, noteID, nil, fileID)
		assert.NoError(t, err)
		assert.Equal(t, block, res)
	})

	t.Run("MissingFileID", func(t *testing.T) {
		_, err := usecase.CreateAttachmentBlock(ctx, userID, noteID, nil, 0)
		assert.Error(t, err)
	})
}

func TestBlocksUsecase_CreateCodeBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{
		BaseBlock: models.BaseBlock{ID: 100, NoteID: noteID, Type: models.BlockTypeCode, Position: 1.0},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		mockBlocksRepo.EXPECT().GetBlocksByNoteIDForPositionCalc(ctx, noteID, uint64(0)).Return([]repository.BlockPositionInfo{}, nil)
		mockBlocksRepo.EXPECT().CreateCodeBlock(ctx, noteID, 1.0, userID).Return(block, nil)

		res, err := usecase.CreateCodeBlock(ctx, userID, noteID, nil)
		assert.NoError(t, err)
		assert.Equal(t, block, res)
	})
}

func TestBlocksUsecase_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlocksRepo := mock.NewMockBlocksRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewBlocksUsecase(mockBlocksRepo, mockNotesRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)

	t.Run("UpdateTextBlock_Success", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, blockID).Return(noteID, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		
		existingBlock := &models.Block{
			BaseBlock: models.BaseBlock{ID: blockID, NoteID: noteID, Type: models.BlockTypeText},
		}
		mockBlocksRepo.EXPECT().GetBlockByID(ctx, blockID).Return(existingBlock, nil)

		content := models.UpdateTextContent{
			Text: "updated text",
			Formats: []models.BlockTextFormat{
				{StartOffset: 0, EndOffset: 5, Bold: true},
			},
		}
		req := &models.UpdateBlockRequest{
			BlockID: blockID,
			Content: content,
		}

		mockBlocksRepo.EXPECT().UpdateBlockText(ctx, blockID, content.Text, gomock.Any()).Return(existingBlock, nil)

		res, err := usecase.UpdateBlock(ctx, userID, req)
		assert.NoError(t, err)
		assert.Equal(t, existingBlock, res)
	})

	t.Run("UpdateCodeBlock_Success", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, blockID).Return(noteID, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		
		existingBlock := &models.Block{
			BaseBlock: models.BaseBlock{ID: blockID, NoteID: noteID, Type: models.BlockTypeCode},
		}
		mockBlocksRepo.EXPECT().GetBlockByID(ctx, blockID).Return(existingBlock, nil)

		content := models.UpdateCodeContent{
			Code:     "console.log()",
			Language: "javascript",
		}
		req := &models.UpdateBlockRequest{
			BlockID: blockID,
			Content: content,
		}

		mockBlocksRepo.EXPECT().UpdateCodeBlock(ctx, blockID, content.Language, content.Code).Return(existingBlock, nil)

		res, err := usecase.UpdateBlock(ctx, userID, req)
		assert.NoError(t, err)
		assert.Equal(t, existingBlock, res)
	})

	t.Run("UpdateAttachmentBlock_Fail", func(t *testing.T) {
		mockBlocksRepo.EXPECT().GetBlockNoteID(ctx, blockID).Return(noteID, nil)
		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true, CanEdit: true}, nil)
		
		existingBlock := &models.Block{
			BaseBlock: models.BaseBlock{ID: blockID, NoteID: noteID, Type: models.BlockTypeAttachment},
		}
		mockBlocksRepo.EXPECT().GetBlockByID(ctx, blockID).Return(existingBlock, nil)

		req := &models.UpdateBlockRequest{BlockID: blockID}

		_, err := usecase.UpdateBlock(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "updating attachment blocks is not supported")
	})
}

func TestOptimizeFormats(t *testing.T) {
	text := "hello world"
	formats := []models.BlockTextFormat{
		{StartOffset: 0, EndOffset: 5, Bold: true, Font: models.FontInter, Size: 12},
		{StartOffset: 0, EndOffset: 5, Italic: true, Font: models.FontInter, Size: 12},
		{StartOffset: 6, EndOffset: 11, Underline: true, Font: models.FontInter, Size: 12},
	}

	optimized := optimizeFormats(text, formats)
	assert.NotEmpty(t, optimized)
	
	assert.Equal(t, 0, optimized[0].StartOffset)
	assert.Equal(t, 5, optimized[0].EndOffset)
	assert.True(t, optimized[0].Bold)
	assert.True(t, optimized[0].Italic)

	assert.Equal(t, 6, optimized[1].StartOffset)
	assert.Equal(t, 11, optimized[1].EndOffset)
	assert.True(t, optimized[1].Underline)
}
