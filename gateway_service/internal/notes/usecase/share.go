package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/utils"
	"context"
	"fmt"
)

func (u *NotesUsecase) AddCollaborator(ctx context.Context, input *models.AddCollaboratorInput) (*models.CollaboratorResponse, error) {
	targetUserID, err := u.userRepo.GetUserIDByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	resp, err := u.repo.AddCollaborator(ctx, input.CurrentUserID, input.NoteID, targetUserID, input.Role)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetUser(ctx, resp.Collaborator.UserID)
	if err != nil {
		return resp, nil
	}

	resp.Collaborator.Email = user.Email
	resp.Collaborator.Username = user.Username
	resp.Collaborator.AvatarFileID = user.AvatarFileID

	return resp, nil
}

func (u *NotesUsecase) GetCollaborators(ctx context.Context, currentUserID, noteID uint64) (*models.GetCollaboratorsResponse, error) {
	resp, err := u.repo.GetCollaborators(ctx, currentUserID, noteID)
	if err != nil {
		return nil, err
	}

	for i := range resp.Collaborators {
		user, err := u.userRepo.GetUser(ctx, resp.Collaborators[i].UserID)
		if err != nil {
			continue
		}

		resp.Collaborators[i].Email = user.Email
		resp.Collaborators[i].Username = user.Username
		resp.Collaborators[i].AvatarFileID = user.AvatarFileID
	}

	owner, err := u.userRepo.GetUser(ctx, resp.OwnerID)
	if err == nil {
		resp.OwnerEmail = owner.Email
		resp.OwnerUsername = owner.Username
		resp.OwnerAvatarFileID = owner.AvatarFileID
	}

	return resp, nil
}

func (u *NotesUsecase) UpdateCollaboratorRole(ctx context.Context, input *models.UpdateCollaboratorRoleInput) (*models.CollaboratorResponse, error) {
	resp, err := u.repo.UpdateCollaboratorRole(ctx, input)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetUser(ctx, resp.Collaborator.UserID)
	if err != nil {
		return resp, nil
	}

	resp.Collaborator.Email = user.Email
	resp.Collaborator.Username = user.Username
	resp.Collaborator.AvatarFileID = user.AvatarFileID

	return resp, nil
}

func (u *NotesUsecase) RemoveCollaborator(ctx context.Context, currentUserID, noteID, permissionID uint64) error {
	return u.repo.RemoveCollaborator(ctx, currentUserID, noteID, permissionID)
}

func (u *NotesUsecase) SetPublicAccess(ctx context.Context, input *models.SetPublicAccessInput) (*models.PublicAccessResponse, error) {
	resp, err := u.repo.SetPublicAccess(ctx, input)
	if err != nil {
		return nil, err
	}

	if resp.ShareURL != "" {
		resp.ShareURL = utils.TransformShareURL(resp.ShareURL)
	}

	return resp, nil
}

func (u *NotesUsecase) GetPublicAccess(ctx context.Context, currentUserID, noteID uint64) (*models.PublicAccessResponse, error) {
	resp, err := u.repo.GetPublicAccess(ctx, currentUserID, noteID)
	if err != nil {
		return nil, err
	}

	if resp.ShareURL != "" {
		resp.ShareURL = utils.TransformShareURL(resp.ShareURL)
	}

	return resp, nil
}

func (u *NotesUsecase) GetSharingSettings(ctx context.Context, currentUserID, noteID uint64) (*models.SharingSettingsResponse, error) {
	resp, err := u.repo.GetSharingSettings(ctx, currentUserID, noteID)
	if err != nil {
		return nil, err
	}

	for i := range resp.Collaborators {
		user, err := u.userRepo.GetUser(ctx, resp.Collaborators[i].UserID)
		if err != nil {
			continue
		}

		resp.Collaborators[i].Email = user.Email
		resp.Collaborators[i].Username = user.Username
		resp.Collaborators[i].AvatarFileID = user.AvatarFileID
	}

	owner, err := u.userRepo.GetUser(ctx, resp.OwnerID)
	if err == nil {
		resp.OwnerEmail = owner.Email
		resp.OwnerUsername = owner.Username
		resp.OwnerAvatarFileID = owner.AvatarFileID
	}

	if resp.PublicAccess.ShareURL != "" {
		resp.PublicAccess.ShareURL = utils.TransformShareURL(resp.PublicAccess.ShareURL)
	}

	return resp, nil
}

func (u *NotesUsecase) ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error) {
	return u.repo.ActivateAccessByLink(ctx, shareUUID, userID)
}
