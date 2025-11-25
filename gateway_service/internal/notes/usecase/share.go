package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"context"
	"fmt"
)

func (u *NotesUsecase) AddCollaborator(ctx context.Context, input *models.AddCollaboratorInput) (*models.CollaboratorResponse, error) {
	targetUserID, err := u.userRepo.GetUserIDByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return u.repo.AddCollaborator(ctx, input.CurrentUserID, input.NoteID, targetUserID, input.Role)
}

func (u *NotesUsecase) GetCollaborators(ctx context.Context, currentUserID, noteID uint64) (*models.GetCollaboratorsResponse, error) {
	return u.repo.GetCollaborators(ctx, currentUserID, noteID)
}

func (u *NotesUsecase) UpdateCollaboratorRole(ctx context.Context, input *models.UpdateCollaboratorRoleInput) (*models.CollaboratorResponse, error) {
	return u.repo.UpdateCollaboratorRole(ctx, input)
}

func (u *NotesUsecase) RemoveCollaborator(ctx context.Context, currentUserID, noteID, permissionID uint64) error {
	return u.repo.RemoveCollaborator(ctx, currentUserID, noteID, permissionID)
}

func (u *NotesUsecase) SetPublicAccess(ctx context.Context, input *models.SetPublicAccessInput) (*models.PublicAccessResponse, error) {
	return u.repo.SetPublicAccess(ctx, input)
}

func (u *NotesUsecase) GetPublicAccess(ctx context.Context, currentUserID, noteID uint64) (*models.PublicAccessResponse, error) {
	return u.repo.GetPublicAccess(ctx, currentUserID, noteID)
}

func (u *NotesUsecase) GetSharingSettings(ctx context.Context, currentUserID, noteID uint64) (*models.SharingSettingsResponse, error) {
	return u.repo.GetSharingSettings(ctx, currentUserID, noteID)
}

func (u *NotesUsecase) ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error) {
	return u.repo.ActivateAccessByLink(ctx, shareUUID, userID)
}
