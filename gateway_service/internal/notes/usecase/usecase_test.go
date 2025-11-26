package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/notes/usecase/mock"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNotesUsecase_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	expectedNotes := []models.Note{{ID: 1, Title: "Test Note", OwnerID: userID}}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetAllNotes(ctx, userID).Return(expectedNotes, nil)

		notes, err := usecase.GetAllNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, notes)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().GetAllNotes(ctx, userID).Return(nil, errors.New("repo error"))

		_, err := usecase.GetAllNotes(ctx, userID)
		assert.Error(t, err)
	})
}

func TestNotesUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	parentNoteID := uint64(5)
	expectedNote := &models.Note{ID: 1, Title: "New Note", OwnerID: userID, ParentNoteID: &parentNoteID}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetNoteById(ctx, userID, parentNoteID).Return(&models.Note{ID: parentNoteID}, nil)

		mockRepo.EXPECT().CreateNote(ctx, userID, &parentNoteID).Return(expectedNote, nil)

		note, err := usecase.CreateNote(ctx, userID, &parentNoteID)
		assert.NoError(t, err)
		assert.Equal(t, expectedNote, note)
	})

	t.Run("MaxNestingLevel", func(t *testing.T) {
		grandParentID := uint64(2)
		parentNote := &models.Note{ID: parentNoteID, ParentNoteID: &grandParentID}

		mockRepo.EXPECT().GetNoteById(ctx, userID, parentNoteID).Return(parentNote, nil)

		_, err := usecase.CreateNote(ctx, userID, &parentNoteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum nesting level")
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().GetNoteById(ctx, userID, parentNoteID).Return(&models.Note{ID: parentNoteID}, nil)
		mockRepo.EXPECT().CreateNote(ctx, userID, &parentNoteID).Return(nil, errors.New("repo error"))

		_, err := usecase.CreateNote(ctx, userID, &parentNoteID)
		assert.Error(t, err)
	})
}

func TestNotesUsecase_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	expectedNote := &models.Note{ID: noteID, Title: "Test Note", OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetNoteById(ctx, userID, noteID).Return(expectedNote, nil)

		note, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, expectedNote, note)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().GetNoteById(ctx, userID, noteID).Return(nil, errors.New("repo error"))

		_, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.Error(t, err)
	})
}

func TestNotesUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	title := "Updated Title"
	input := &models.UpdateNoteInput{UserID: userID, ID: noteID, Title: &title}
	expectedNote := &models.Note{ID: noteID, Title: title, OwnerID: userID}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().UpdateNote(ctx, input).Return(expectedNote, nil)

		note, err := usecase.UpdateNote(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedNote, note)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().UpdateNote(ctx, input).Return(nil, errors.New("repo error"))

		_, err := usecase.UpdateNote(ctx, input)
		assert.Error(t, err)
	})
}

func TestNotesUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteNote(ctx, userID, noteID).Return(nil)

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().DeleteNote(ctx, userID, noteID).Return(errors.New("repo error"))

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.Error(t, err)
	})
}

func TestNotesUsecase_Favorites(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("AddFavorite_Success", func(t *testing.T) {
		mockRepo.EXPECT().AddFavorite(ctx, userID, noteID).Return(nil)
		err := usecase.AddFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("RemoveFavorite_Success", func(t *testing.T) {
		mockRepo.EXPECT().RemoveFavorite(ctx, userID, noteID).Return(nil)
		err := usecase.RemoveFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesUsecase_Blocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)
	now := time.Now()

	t.Run("GetBlocks", func(t *testing.T) {
		expectedBlocks := []models.Block{{ID: blockID, NoteID: noteID}}
		mockRepo.EXPECT().GetBlocks(ctx, userID, noteID).Return(expectedBlocks, nil)
		blocks, err := usecase.GetBlocks(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlocks, blocks)
	})

	t.Run("GetBlock", func(t *testing.T) {
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID}
		mockRepo.EXPECT().GetBlock(ctx, userID, blockID).Return(expectedBlock, nil)
		block, err := usecase.GetBlock(ctx, userID, blockID)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})

	t.Run("CreateTextBlock", func(t *testing.T) {
		input := &models.CreateTextBlockInput{UserID: userID, NoteID: noteID}
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID, Type: "text"}
		mockRepo.EXPECT().CreateTextBlock(ctx, input).Return(expectedBlock, nil)
		block, err := usecase.CreateTextBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})

	t.Run("CreateCodeBlock", func(t *testing.T) {
		input := &models.CreateCodeBlockInput{UserID: userID, NoteID: noteID}
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID, Type: "code"}
		mockRepo.EXPECT().CreateCodeBlock(ctx, input).Return(expectedBlock, nil)
		block, err := usecase.CreateCodeBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})

	t.Run("CreateAttachmentBlock", func(t *testing.T) {
		input := &models.CreateAttachmentBlockInput{UserID: userID, NoteID: noteID}
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID, Type: "file"}
		mockRepo.EXPECT().CreateAttachmentBlock(ctx, input).Return(expectedBlock, nil)
		block, err := usecase.CreateAttachmentBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})

	t.Run("UpdateBlock", func(t *testing.T) {
		input := &models.UpdateBlockInput{BlockID: blockID}
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID, UpdatedAt: now}
		mockRepo.EXPECT().UpdateBlock(ctx, userID, input).Return(expectedBlock, nil)
		block, err := usecase.UpdateBlock(ctx, userID, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})

	t.Run("DeleteBlock", func(t *testing.T) {
		mockRepo.EXPECT().DeleteBlock(ctx, userID, blockID).Return(nil)
		err := usecase.DeleteBlock(ctx, userID, blockID)
		assert.NoError(t, err)
	})

	t.Run("UpdateBlockPosition", func(t *testing.T) {
		beforeBlockID := uint64(101)
		expectedBlock := &models.Block{ID: blockID, NoteID: noteID}
		mockRepo.EXPECT().UpdateBlockPosition(ctx, userID, blockID, &beforeBlockID).Return(expectedBlock, nil)
		block, err := usecase.UpdateBlockPosition(ctx, userID, blockID, &beforeBlockID)
		assert.NoError(t, err)
		assert.Equal(t, expectedBlock, block)
	})
}
