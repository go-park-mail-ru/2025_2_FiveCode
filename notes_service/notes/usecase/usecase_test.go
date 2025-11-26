package usecase

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/notes/mock"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNoteUsecase_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	expectedNotes := []models.Note{{ID: 1, Title: "Note 1"}}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNotes(ctx, userID).
			Return(expectedNotes, nil)

		notes, err := usecase.GetAllNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, notes)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNotes(ctx, userID).
			Return(nil, errors.New("db error"))

		notes, err := usecase.GetAllNotes(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, notes)
	})
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	note := &models.Note{ID: 1, Title: "New Note"}

	t.Run("Success_NoParent", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateNote(ctx, userID, nil).
			Return(note, nil)

		createdNote, err := usecase.CreateNote(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, note, createdNote)
	})

	t.Run("Success_WithParent", func(t *testing.T) {
		parentID := uint64(2)
		parentNote := &models.Note{ID: parentID}
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true, Role: models.RoleEditor}

		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, parentID, userID).
			Return(parentNote, nil)

		mockRepo.EXPECT().
			CreateNote(ctx, userID, &parentID).
			Return(note, nil)

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.NoError(t, err)
		assert.Equal(t, note, createdNote)
	})

	t.Run("Fail_NoAccessToParent", func(t *testing.T) {
		parentID := uint64(2)
		accessInfo := &models.NoteAccessInfo{HasAccess: false}

		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(accessInfo, nil)

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateNote(ctx, userID, nil).
			Return(nil, errors.New("db error"))

		createdNote, err := usecase.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
	})
}

func TestNoteUsecase_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)
	note := &models.Note{ID: noteID}

	t.Run("Success", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(note, nil)

		result, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})

	t.Run("NoAccess", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})
}

func TestNoteUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)
	title := "Updated Title"
	note := &models.Note{ID: noteID, Title: title}

	t.Run("Success", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			UpdateNote(ctx, noteID, &title, nil).
			Return(note, nil)

		result, err := usecase.UpdateNote(ctx, userID, noteID, &title, nil)
		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})

	t.Run("CannotEdit", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.UpdateNote(ctx, userID, noteID, &title, nil)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})
}

func TestNoteUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)

	t.Run("Success", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, IsOwner: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			DeleteNote(ctx, noteID).
			Return(nil)

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("NotOwner", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, IsOwner: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
	})
}

func TestNoteUsecase_AddFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(&models.NoteAccessInfo{HasAccess: true}, nil)

		mockRepo.EXPECT().
			AddFavorite(ctx, userID, noteID).
			Return(nil)

		err := usecase.AddFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(&models.NoteAccessInfo{HasAccess: false}, nil)

		err := usecase.AddFavorite(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
	})
}

func TestNoteUsecase_RemoveFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)

	t.Run("Success", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(&models.NoteAccessInfo{HasAccess: true}, nil)

		mockRepo.EXPECT().
			RemoveFavorite(ctx, userID, noteID).
			Return(nil)

		err := usecase.RemoveFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNoteUsecase_GetNoteByShareUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	shareUUID := "some-uuid"
	note := &models.Note{ID: 1, Title: "Shared Note"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNoteByShareUUID(ctx, shareUUID).
			Return(note, nil)

		result, err := usecase.GetNoteByShareUUID(ctx, shareUUID)
		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})
}
