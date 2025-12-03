package usecase

import (
	"context"
	"errors"
	"testing"

	"backend/notes_service/internal/models"
	"backend/notes_service/sharing/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) (*gomock.Controller, *mock.MockSharingRepository, *mock.MockNotesRepository, *SharingUsecase) {
	ctrl := gomock.NewController(t)
	mockSharingRepo := mock.NewMockSharingRepository(ctrl)
	mockNotesRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewSharingUsecase(mockSharingRepo, mockNotesRepo)
	return ctrl, mockSharingRepo, mockNotesRepo, usecase
}

func TestSharingUsecase_CheckNoteAccess(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)

		res, err := usecase.CheckNoteAccess(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, res.HasAccess)
	})

	t.Run("Error", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(nil, errors.New("repo error"))

		_, err := usecase.CheckNoteAccess(ctx, noteID, userID)
		assert.Error(t, err)
	})
}

func TestSharingUsecase_AddCollaborator(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	targetUserID := uint64(2)
	role := models.RoleEditor
	permission := &models.NotePermission{PermissionID: 5, GrantedTo: targetUserID, Role: role}

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, targetUserID).Return(false, nil)
		mockSharingRepo.EXPECT().CheckCollaboratorExists(ctx, noteID, targetUserID).Return(false, nil)
		mockSharingRepo.EXPECT().AddCollaborator(ctx, gomock.Any()).Return(permission, nil)
		mockSharingRepo.EXPECT().GetCollaboratorsByNoteID(ctx, noteID).Return([]*models.NotePermission{permission}, nil)
		mockSharingRepo.EXPECT().UpdateIsSharedFlag(ctx, noteID, true).Return(nil)

		res, err := usecase.AddCollaborator(ctx, noteID, currentUserID, targetUserID, role)
		assert.NoError(t, err)
		assert.Equal(t, permission, res)
	})

	t.Run("IsSubNote", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		parentID := uint64(99)
		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(&parentID, nil)

		_, err := usecase.AddCollaborator(ctx, noteID, currentUserID, targetUserID, role)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sub-notes cannot be shared")
	})

	t.Run("NotOwner", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(false, nil)

		_, err := usecase.AddCollaborator(ctx, noteID, currentUserID, targetUserID, role)
		assert.Error(t, err)
	})

	t.Run("TargetIsOwner", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, targetUserID).Return(true, nil)

		_, err := usecase.AddCollaborator(ctx, noteID, currentUserID, targetUserID, role)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot add note owner as collaborator")
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, targetUserID).Return(false, nil)
		mockSharingRepo.EXPECT().CheckCollaboratorExists(ctx, noteID, targetUserID).Return(true, nil)

		_, err := usecase.AddCollaborator(ctx, noteID, currentUserID, targetUserID, role)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already has access")
	})
}

func TestSharingUsecase_GetCollaborators(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	ownerID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, currentUserID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)
		mockSharingRepo.EXPECT().GetNoteOwnerID(ctx, noteID).Return(ownerID, nil)
		mockSharingRepo.EXPECT().GetCollaboratorsByNoteID(ctx, noteID).Return([]*models.NotePermission{}, nil)
		mockSharingRepo.EXPECT().GetPublicAccess(ctx, noteID).Return(nil, "", nil)

		oid, _, _, err := usecase.GetCollaborators(ctx, noteID, currentUserID)
		assert.NoError(t, err)
		assert.Equal(t, ownerID, oid)
	})

	t.Run("NoAccess", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, currentUserID).Return(&models.NoteAccessInfo{HasAccess: false}, nil)

		_, _, _, err := usecase.GetCollaborators(ctx, noteID, currentUserID)
		assert.Error(t, err)
	})
}

func TestSharingUsecase_UpdateCollaboratorRole(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	permissionID := uint64(5)
	newRole := models.RoleViewer
	permission := &models.NotePermission{PermissionID: permissionID, NoteID: noteID, Role: models.RoleEditor}

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().GetCollaboratorByID(ctx, permissionID).Return(permission, nil)
		mockSharingRepo.EXPECT().UpdateCollaboratorRole(ctx, permissionID, newRole).Return(nil)
		mockSharingRepo.EXPECT().GetCollaboratorByID(ctx, permissionID).Return(permission, nil)

		res, err := usecase.UpdateCollaboratorRole(ctx, noteID, currentUserID, permissionID, newRole)
		assert.NoError(t, err)
		assert.Equal(t, permission, res)
	})

	t.Run("WrongNoteID", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		wrongPermission := &models.NotePermission{PermissionID: permissionID, NoteID: 999}
		mockSharingRepo.EXPECT().GetCollaboratorByID(ctx, permissionID).Return(wrongPermission, nil)

		_, err := usecase.UpdateCollaboratorRole(ctx, noteID, currentUserID, permissionID, newRole)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to this note")
	})
}

func TestSharingUsecase_RemoveCollaborator(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	permissionID := uint64(5)
	permission := &models.NotePermission{PermissionID: permissionID, NoteID: noteID}

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().GetCollaboratorByID(ctx, permissionID).Return(permission, nil)
		mockSharingRepo.EXPECT().RemoveCollaborator(ctx, permissionID).Return(nil)
		mockSharingRepo.EXPECT().GetCollaboratorsByNoteID(ctx, noteID).Return([]*models.NotePermission{}, nil)
		mockSharingRepo.EXPECT().UpdateIsSharedFlag(ctx, noteID, false).Return(nil)

		err := usecase.RemoveCollaborator(ctx, noteID, currentUserID, permissionID)
		assert.NoError(t, err)
	})

	t.Run("WrongNoteID", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		wrongPermission := &models.NotePermission{PermissionID: permissionID, NoteID: 999}
		mockSharingRepo.EXPECT().GetCollaboratorByID(ctx, permissionID).Return(wrongPermission, nil)

		err := usecase.RemoveCollaborator(ctx, noteID, currentUserID, permissionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to this note")
	})
}

func TestSharingUsecase_SetPublicAccess(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	accessLevel := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().IsNoteOwner(ctx, noteID, currentUserID).Return(true, nil)
		mockSharingRepo.EXPECT().SetPublicAccess(ctx, noteID, &accessLevel).Return(nil)

		err := usecase.SetPublicAccess(ctx, noteID, currentUserID, &accessLevel)
		assert.NoError(t, err)
	})
}

func TestSharingUsecase_GetPublicAccess(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	accessLevel := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, currentUserID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)
		mockSharingRepo.EXPECT().GetPublicAccess(ctx, noteID).Return(&accessLevel, "uuid", nil)

		res, err := usecase.GetPublicAccess(ctx, noteID, currentUserID)
		assert.NoError(t, err)
		assert.Equal(t, &accessLevel, res)
	})
}

func TestSharingUsecase_GetSharingSettings(t *testing.T) {
	ctx := context.Background()
	noteID := uint64(10)
	currentUserID := uint64(1)
	ownerID := uint64(1)
	accessLevel := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		ctrl, mockSharingRepo, _, usecase := setupTest(t)
		defer ctrl.Finish()

		mockSharingRepo.EXPECT().CheckNoteAccess(ctx, noteID, currentUserID).Return(&models.NoteAccessInfo{HasAccess: true}, nil)
		mockSharingRepo.EXPECT().GetParentNoteID(ctx, noteID).Return(nil, nil)
		mockSharingRepo.EXPECT().GetNoteOwnerID(ctx, noteID).Return(ownerID, nil)
		mockSharingRepo.EXPECT().GetCollaboratorsByNoteID(ctx, noteID).Return([]*models.NotePermission{}, nil)
		mockSharingRepo.EXPECT().GetPublicAccess(ctx, noteID).Return(&accessLevel, "uuid", nil)

		settings, err := usecase.GetSharingSettings(ctx, noteID, currentUserID)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.True(t, settings.IsOwner)
		assert.Equal(t, &accessLevel, settings.PublicAccess.AccessLevel)
	})
}

func TestSharingUsecase_ActivateAccessByLink(t *testing.T) {
	ctx := context.Background()
	shareUUID := "uuid"
	userID := uint64(2)
	noteID := uint64(10)
	note := &models.Note{ID: noteID, OwnerID: 1}
	accessLevel := models.RoleViewer

	t.Run("Success_Public", func(t *testing.T) {
		ctrl, mockSharingRepo, mockNotesRepo, usecase := setupTest(t)
		defer ctrl.Finish()

		mockNotesRepo.EXPECT().GetNoteByShareUUID(ctx, shareUUID).Return(note, nil)
		mockSharingRepo.EXPECT().GetUserPermission(ctx, noteID, userID).Return(nil, nil)
		mockSharingRepo.EXPECT().GetPublicAccess(ctx, noteID).Return(&accessLevel, "", nil)
		mockSharingRepo.EXPECT().AddCollaborator(ctx, gomock.Any()).Return(&models.NotePermission{Role: accessLevel}, nil)
		mockSharingRepo.EXPECT().GetCollaboratorsByNoteID(ctx, noteID).Return([]*models.NotePermission{}, nil)
		mockSharingRepo.EXPECT().UpdateIsSharedFlag(ctx, noteID, false).Return(nil)

		res, err := usecase.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.True(t, res.AccessGranted)
	})

	t.Run("Owner", func(t *testing.T) {
		ctrl, _, mockNotesRepo, usecase := setupTest(t)
		defer ctrl.Finish()

		ownerNote := &models.Note{ID: noteID, OwnerID: userID}
		mockNotesRepo.EXPECT().GetNoteByShareUUID(ctx, shareUUID).Return(ownerNote, nil)

		res, err := usecase.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.True(t, res.AccessGranted)
		assert.True(t, res.AccessInfo.IsOwner)
	})

	t.Run("ExistingPermission", func(t *testing.T) {
		ctrl, mockSharingRepo, mockNotesRepo, usecase := setupTest(t)
		defer ctrl.Finish()

		mockNotesRepo.EXPECT().GetNoteByShareUUID(ctx, shareUUID).Return(note, nil)
		mockSharingRepo.EXPECT().GetUserPermission(ctx, noteID, userID).Return(&models.NotePermission{Role: models.RoleEditor}, nil)

		res, err := usecase.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.True(t, res.AccessGranted)
		assert.True(t, res.AccessInfo.CanEdit)
	})

	t.Run("NoPublicAccess", func(t *testing.T) {
		ctrl, mockSharingRepo, mockNotesRepo, usecase := setupTest(t)
		defer ctrl.Finish()

		mockNotesRepo.EXPECT().GetNoteByShareUUID(ctx, shareUUID).Return(note, nil)
		mockSharingRepo.EXPECT().GetUserPermission(ctx, noteID, userID).Return(nil, nil)
		mockSharingRepo.EXPECT().GetPublicAccess(ctx, noteID).Return(nil, "", nil)

		res, err := usecase.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.False(t, res.AccessGranted)
	})
}
