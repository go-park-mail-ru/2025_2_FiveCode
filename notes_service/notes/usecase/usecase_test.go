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

	t.Run("Fail_CannotEditParent", func(t *testing.T) {
		parentID := uint64(2)
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: false}

		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(accessInfo, nil)

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
	})

	t.Run("Fail_ParentIsSubNote", func(t *testing.T) {
		parentID := uint64(2)
		parentNoteID := uint64(1)
		parentNote := &models.Note{ID: parentID, ParentNoteID: &parentNoteID}
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}

		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, parentID, userID).
			Return(parentNote, nil)

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
		assert.Contains(t, err.Error(), "cannot create sub-note of a sub-note")
	})

	t.Run("CheckAccessError", func(t *testing.T) {
		parentID := uint64(2)
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(nil, errors.New("access check error"))

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
		assert.Contains(t, err.Error(), "failed to check access")
	})

	t.Run("GetParentNoteError", func(t *testing.T) {
		parentID := uint64(2)
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}

		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, parentID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, parentID, userID).
			Return(nil, errors.New("not found"))

		createdNote, err := usecase.CreateNote(ctx, userID, &parentID)
		assert.Error(t, err)
		assert.Nil(t, createdNote)
		assert.Contains(t, err.Error(), "parent note not found")
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

	t.Run("CheckAccessError", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(nil, errors.New("access check error"))

		result, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check note access")
	})

	t.Run("GetNoteError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(nil, errors.New("not found"))

		result, err := usecase.GetNoteById(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get note")
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

	t.Run("NoAccess", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.UpdateNote(ctx, userID, noteID, &title, nil)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})

	t.Run("CheckAccessError", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(nil, errors.New("access check error"))

		result, err := usecase.UpdateNote(ctx, userID, noteID, &title, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check note access")
	})

	t.Run("UpdateError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			UpdateNote(ctx, noteID, &title, nil).
			Return(nil, errors.New("update error"))

		result, err := usecase.UpdateNote(ctx, userID, noteID, &title, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update note")
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

	t.Run("NoAccess", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
	})

	t.Run("CheckAccessError", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(nil, errors.New("access check error"))

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check note access")
	})

	t.Run("DeleteError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, IsOwner: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			DeleteNote(ctx, noteID).
			Return(errors.New("delete error"))

		err := usecase.DeleteNote(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete note")
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

	t.Run("CheckAccessError", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(nil, errors.New("access check error"))

		err := usecase.AddFavorite(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check note access")
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

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(&models.NoteAccessInfo{HasAccess: false}, nil)

		err := usecase.RemoveFavorite(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
	})

	t.Run("CheckAccessError", func(t *testing.T) {
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(nil, errors.New("access check error"))

		err := usecase.RemoveFavorite(ctx, userID, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check note access")
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

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNoteByShareUUID(ctx, shareUUID).
			Return(nil, errors.New("not found"))

		result, err := usecase.GetNoteByShareUUID(ctx, shareUUID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestNoteUsecase_SearchNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	query := "test query"
	searchResponse := &models.SearchNotesResponse{
		Results: []models.SearchResult{
			{NoteID: 1, Title: "Test Note 1", HighlightedTitle: "Test Note 1", ContentSnippet: "snippet 1", Rank: 0.5},
			{NoteID: 2, Title: "Test Note 2", HighlightedTitle: "Test Note 2", ContentSnippet: "snippet 2", Rank: 0.3},
		},
		Count: 2,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			SearchNotes(ctx, userID, query).
			Return(searchResponse, nil)

		result, err := usecase.SearchNotes(ctx, userID, query)
		assert.NoError(t, err)
		assert.Equal(t, searchResponse, result)
		assert.Equal(t, 2, result.Count)
	})

	t.Run("EmptyQuery", func(t *testing.T) {
		result, err := usecase.SearchNotes(ctx, userID, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query cannot be empty")
	})

	t.Run("QueryTooLong", func(t *testing.T) {
		longQuery := make([]byte, 201)
		for i := range longQuery {
			longQuery[i] = 'a'
		}

		result, err := usecase.SearchNotes(ctx, userID, string(longQuery))
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query too long")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo.EXPECT().
			SearchNotes(ctx, userID, query).
			Return(nil, errors.New("database error"))

		result, err := usecase.SearchNotes(ctx, userID, query)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to search notes")
	})
}

func TestNoteUsecase_SetIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)
	iconFileID := uint64(10)
	note := &models.Note{ID: noteID, Title: "Test Note"}

	t.Run("Success", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetIcon(ctx, noteID, iconFileID).
			Return(nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(note, nil)

		result, err := usecase.SetIcon(ctx, userID, noteID, iconFileID)
		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})

	t.Run("NoAccess", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.SetIcon(ctx, userID, noteID, iconFileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})

	t.Run("CannotEdit", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.SetIcon(ctx, userID, noteID, iconFileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})

	t.Run("SetIconError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetIcon(ctx, noteID, iconFileID).
			Return(errors.New("database error"))

		result, err := usecase.SetIcon(ctx, userID, noteID, iconFileID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to set icon")
	})

	t.Run("GetNoteError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetIcon(ctx, noteID, iconFileID).
			Return(nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(nil, errors.New("not found"))

		result, err := usecase.SetIcon(ctx, userID, noteID, iconFileID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get note")
	})
}

func TestNoteUsecase_SetHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockSharingRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(2)
	headerFileID := uint64(20)
	note := &models.Note{ID: noteID, Title: "Test Note"}

	t.Run("Success", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetHeader(ctx, noteID, headerFileID).
			Return(nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(note, nil)

		result, err := usecase.SetHeader(ctx, userID, noteID, headerFileID)
		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})

	t.Run("NoAccess", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.SetHeader(ctx, userID, noteID, headerFileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})

	t.Run("CannotEdit", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: false}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		result, err := usecase.SetHeader(ctx, userID, noteID, headerFileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNoAccess, err)
		assert.Nil(t, result)
	})

	t.Run("SetHeaderError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetHeader(ctx, noteID, headerFileID).
			Return(errors.New("database error"))

		result, err := usecase.SetHeader(ctx, userID, noteID, headerFileID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to set header")
	})

	t.Run("GetNoteError", func(t *testing.T) {
		accessInfo := &models.NoteAccessInfo{HasAccess: true, CanEdit: true}
		mockSharingRepo.EXPECT().
			CheckNoteAccess(ctx, noteID, userID).
			Return(accessInfo, nil)

		mockRepo.EXPECT().
			SetHeader(ctx, noteID, headerFileID).
			Return(nil)

		mockRepo.EXPECT().
			GetNoteById(ctx, noteID, userID).
			Return(nil, errors.New("not found"))

		result, err := usecase.SetHeader(ctx, userID, noteID, headerFileID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get note")
	})
}
