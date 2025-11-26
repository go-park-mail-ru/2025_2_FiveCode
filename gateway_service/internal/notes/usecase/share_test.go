package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/notes/usecase/mock"
	userModels "backend/gateway_service/internal/user/models"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNotesUsecase_AddCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	currentUserID := uint64(1)
	targetUserID := uint64(2)
	noteID := uint64(10)
	email := "target@example.com"
	role := models.RoleEditor

	input := &models.AddCollaboratorInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		Email:         email,
		Role:          role,
	}

	targetUser := &userModels.User{
		ID:       targetUserID,
		Email:    email,
		Username: "targetuser",
	}
	avatarID := uint64(55)
	targetUser.AvatarFileID = &avatarID

	t.Run("Success", func(t *testing.T) {
		repoResp := &models.CollaboratorResponse{
			PermissionID: 100,
			Collaborator: models.Collaborator{
				UserID:    targetUserID,
				Role:      role,
				GrantedBy: currentUserID,
			},
		}

		mockUserRepo.EXPECT().GetUserIDByEmail(ctx, email).Return(targetUserID, nil)
		mockRepo.EXPECT().AddCollaborator(ctx, currentUserID, noteID, targetUserID, role).Return(repoResp, nil)
		mockUserRepo.EXPECT().GetUser(ctx, targetUserID).Return(targetUser, nil)

		resp, err := usecase.AddCollaborator(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, email, resp.Collaborator.Email)
		assert.Equal(t, "targetuser", resp.Collaborator.Username)
		assert.Equal(t, &avatarID, resp.Collaborator.AvatarFileID)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserIDByEmail(ctx, email).Return(uint64(0), errors.New("user not found"))

		_, err := usecase.AddCollaborator(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user by email")
	})

	t.Run("RepoError", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserIDByEmail(ctx, email).Return(targetUserID, nil)
		mockRepo.EXPECT().AddCollaborator(ctx, currentUserID, noteID, targetUserID, role).Return(nil, errors.New("repo error"))

		_, err := usecase.AddCollaborator(ctx, input)
		assert.Error(t, err)
	})

	t.Run("GetUserError", func(t *testing.T) {
		repoResp := &models.CollaboratorResponse{
			PermissionID: 100,
			Collaborator: models.Collaborator{
				UserID:    targetUserID,
				Role:      role,
				GrantedBy: currentUserID,
			},
		}

		mockUserRepo.EXPECT().GetUserIDByEmail(ctx, email).Return(targetUserID, nil)
		mockRepo.EXPECT().AddCollaborator(ctx, currentUserID, noteID, targetUserID, role).Return(repoResp, nil)
		mockUserRepo.EXPECT().GetUser(ctx, targetUserID).Return(nil, errors.New("get user error"))

		resp, err := usecase.AddCollaborator(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, repoResp, resp)
		assert.Empty(t, resp.Collaborator.Email)
	})
}

func TestNotesUsecase_GetCollaborators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)
	ownerID := uint64(1)
	collabUserID := uint64(2)

	repoResp := &models.GetCollaboratorsResponse{
		OwnerID: ownerID,
		Collaborators: []models.Collaborator{
			{UserID: collabUserID, Role: models.RoleEditor},
		},
	}

	ownerUser := &userModels.User{ID: ownerID, Email: "owner@example.com", Username: "owner"}
	collabUser := &userModels.User{ID: collabUserID, Email: "collab@example.com", Username: "collab"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetCollaborators(ctx, currentUserID, noteID).Return(repoResp, nil)

		mockUserRepo.EXPECT().GetUser(ctx, collabUserID).Return(collabUser, nil)

		mockUserRepo.EXPECT().GetUser(ctx, ownerID).Return(ownerUser, nil)

		resp, err := usecase.GetCollaborators(ctx, currentUserID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, "collab@example.com", resp.Collaborators[0].Email)
		assert.Equal(t, "owner@example.com", resp.OwnerEmail)
	})
}

func TestNotesUsecase_UpdateCollaboratorRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	input := &models.UpdateCollaboratorRoleInput{PermissionID: 1, NewRole: models.RoleViewer}
	targetUserID := uint64(2)

	repoResp := &models.CollaboratorResponse{
		PermissionID: 1,
		Collaborator: models.Collaborator{UserID: targetUserID, Role: models.RoleViewer},
	}
	targetUser := &userModels.User{ID: targetUserID, Email: "target@example.com", Username: "target"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().UpdateCollaboratorRole(ctx, input).Return(repoResp, nil)
		mockUserRepo.EXPECT().GetUser(ctx, targetUserID).Return(targetUser, nil)

		resp, err := usecase.UpdateCollaboratorRole(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "target@example.com", resp.Collaborator.Email)
	})
}

func TestNotesUsecase_PublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)

	t.Run("SetPublicAccess", func(t *testing.T) {
		accessLevel := models.RoleViewer
		input := &models.SetPublicAccessInput{NoteID: noteID, AccessLevel: &accessLevel}
		expectedResp := &models.PublicAccessResponse{AccessLevel: &accessLevel, ShareURL: "some-url"}

		mockRepo.EXPECT().SetPublicAccess(ctx, input).Return(expectedResp, nil)

		resp, err := usecase.SetPublicAccess(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, &accessLevel, resp.AccessLevel)
	})

	t.Run("GetPublicAccess", func(t *testing.T) {
		accessLevel := models.RoleViewer
		expectedResp := &models.PublicAccessResponse{AccessLevel: &accessLevel, ShareURL: "some-url"}
		mockRepo.EXPECT().GetPublicAccess(ctx, currentUserID, noteID).Return(expectedResp, nil)

		resp, err := usecase.GetPublicAccess(ctx, currentUserID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, &accessLevel, resp.AccessLevel)
	})
}

func TestNotesUsecase_ActivateAccessByLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo, mockUserRepo)

	ctx := context.Background()
	shareUUID := "uuid"
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		expectedResp := &models.ActivateAccessResponse{
			NoteID:        10,
			AccessGranted: true,
			AccessInfo:    models.NoteAccessInfo{HasAccess: true, Role: models.RoleEditor},
		}
		mockRepo.EXPECT().ActivateAccessByLink(ctx, shareUUID, userID).Return(expectedResp, nil)

		resp, err := usecase.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})
}
