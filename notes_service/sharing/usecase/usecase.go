package usecase

import (
	"backend/notes_service/internal/constants"
	"context"
	"fmt"
	"time"

	"backend/notes_service/internal/models"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type SharingRepository interface {
	AddCollaborator(ctx context.Context, permission *models.NotePermission) (*models.NotePermission, error)
	GetCollaboratorsByNoteID(ctx context.Context, noteID uint64) ([]*models.NotePermission, error)
	GetCollaboratorByID(ctx context.Context, permissionID uint64) (*models.NotePermission, error)
	UpdateCollaboratorRole(ctx context.Context, permissionID uint64, role models.NoteRole) error
	RemoveCollaborator(ctx context.Context, permissionID uint64) error
	CheckCollaboratorExists(ctx context.Context, noteID, userID uint64) (bool, error)

	SetPublicAccess(ctx context.Context, noteID uint64, accessLevel *models.NoteRole) error
	GetPublicAccess(ctx context.Context, noteID uint64) (*models.NoteRole, string, error)

	GetNoteOwnerID(ctx context.Context, noteID uint64) (uint64, error)
	CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error)
	IsNoteOwner(ctx context.Context, noteID, userID uint64) (bool, error)
	GetUserPermission(ctx context.Context, noteID, userID uint64) (*models.NotePermission, error)
	CanUserShare(ctx context.Context, noteID, userID uint64) (bool, error)

	UpdateIsSharedFlag(ctx context.Context, noteID uint64, isShared bool) error

	GetParentNoteID(ctx context.Context, noteID uint64) (*uint64, error)
}

type NotesRepository interface {
	GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error)
}

type SharingUsecase struct {
	sharingRepo SharingRepository
	notesRepo   NotesRepository
}

func NewSharingUsecase(sharingRepo SharingRepository, notesRepo NotesRepository) *SharingUsecase {
	return &SharingUsecase{
		sharingRepo: sharingRepo,
		notesRepo:   notesRepo,
	}
}

func (uc *SharingUsecase) validateNoteOwnership(ctx context.Context, noteID, userID uint64) error {
	isOwner, err := uc.sharingRepo.IsNoteOwner(ctx, noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %w", err)
	}

	if !isOwner {
		return fmt.Errorf("access denied: user is not the note owner")
	}

	return nil
}

func (uc *SharingUsecase) validateNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error) {
	accessInfo, err := uc.sharingRepo.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		return nil, fmt.Errorf("access denied: user does not have access to this note")
	}

	return accessInfo, nil
}

func (uc *SharingUsecase) CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error) {
	accessInfo, err := uc.sharingRepo.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	return accessInfo, nil
}

func (uc *SharingUsecase) updateIsSharedFlag(ctx context.Context, noteID uint64) error {
	collaborators, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get collaborators count: %w", err)
	}

	isShared := len(collaborators) > 0

	if err := uc.sharingRepo.UpdateIsSharedFlag(ctx, noteID, isShared); err != nil {
		return fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return nil
}

func (uc *SharingUsecase) AddCollaborator(ctx context.Context, noteID, currentUserID, targetUserID uint64, role models.NoteRole) (*models.NotePermission, error) {
	parentNoteID, err := uc.sharingRepo.GetParentNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent note id: %w", err)
	}
	if parentNoteID != nil {
		return nil, constants.ErrSubNoteCannotBeShared
	}

	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return nil, err
	}

	isOwner, err := uc.sharingRepo.IsNoteOwner(ctx, noteID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if target user is owner: %w", err)
	}
	if isOwner {
		return nil, fmt.Errorf("cannot add note owner as collaborator")
	}

	exists, err := uc.sharingRepo.CheckCollaboratorExists(ctx, noteID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check collaborator exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has access to this note")
	}

	permission := &models.NotePermission{
		NoteID:    noteID,
		GrantedBy: currentUserID,
		GrantedTo: targetUserID,
		Role:      role,
	}

	createdPermission, err := uc.sharingRepo.AddCollaborator(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to add collaborator: %w", err)
	}

	if err := uc.updateIsSharedFlag(ctx, noteID); err != nil {
		return nil, fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return createdPermission, nil
}

func (uc *SharingUsecase) GetCollaborators(ctx context.Context, noteID, currentUserID uint64) (uint64, []*models.NotePermission, *models.NoteRole, error) {
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return 0, nil, nil, err
	}

	ownerID, err := uc.sharingRepo.GetNoteOwnerID(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get note owner: %w", err)
	}

	permissions, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	publicAccess, _, err := uc.sharingRepo.GetPublicAccess(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get public access: %w", err)
	}

	ownerPermission := &models.NotePermission{
		PermissionID: 0,
		NoteID:       noteID,
		GrantedTo:    ownerID,
		GrantedBy:    ownerID,
		Role:         models.RoleOwner,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	allPermissions := append([]*models.NotePermission{ownerPermission}, permissions...)

	return ownerID, allPermissions, publicAccess, nil
}

func (uc *SharingUsecase) UpdateCollaboratorRole(ctx context.Context, noteID, currentUserID, permissionID uint64, newRole models.NoteRole) (*models.NotePermission, error) {
	parentNoteID, err := uc.sharingRepo.GetParentNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent note id: %w", err)
	}
	if parentNoteID != nil {
		return nil, constants.ErrSubNoteCannotBeShared
	}

	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return nil, err
	}

	permission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}

	if permission.NoteID != noteID {
		return nil, fmt.Errorf("permission does not belong to this note")
	}

	if err := uc.sharingRepo.UpdateCollaboratorRole(ctx, permissionID, newRole); err != nil {
		return nil, fmt.Errorf("failed to update collaborator role: %w", err)
	}

	updatedPermission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated collaborator: %w", err)
	}

	return updatedPermission, nil
}

func (uc *SharingUsecase) RemoveCollaborator(ctx context.Context, noteID, currentUserID, permissionID uint64) error {
	parentNoteID, err := uc.sharingRepo.GetParentNoteID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get parent note id: %w", err)
	}
	if parentNoteID != nil {
		return constants.ErrSubNoteCannotBeShared
	}

	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return err
	}

	permission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get collaborator: %w", err)
	}

	if permission.NoteID != noteID {
		return fmt.Errorf("permission does not belong to this note")
	}

	if err := uc.sharingRepo.RemoveCollaborator(ctx, permissionID); err != nil {
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	if err := uc.updateIsSharedFlag(ctx, noteID); err != nil {
		return fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return nil
}

func (uc *SharingUsecase) SetPublicAccess(ctx context.Context, noteID, currentUserID uint64, accessLevel *models.NoteRole) error {
	parentNoteID, err := uc.sharingRepo.GetParentNoteID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get parent note id: %w", err)
	}
	if parentNoteID != nil {
		return constants.ErrSubNoteCannotBeShared
	}

	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return err
	}

	if err := uc.sharingRepo.SetPublicAccess(ctx, noteID, accessLevel); err != nil {
		return fmt.Errorf("failed to set public access: %w", err)
	}

	return nil
}

func (uc *SharingUsecase) GetPublicAccess(ctx context.Context, noteID, currentUserID uint64) (*models.NoteRole, error) {
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return nil, err
	}

	accessLevel, _, err := uc.sharingRepo.GetPublicAccess(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access: %w", err)
	}

	return accessLevel, nil
}

func (uc *SharingUsecase) GetSharingSettings(ctx context.Context, noteID, currentUserID uint64) (*models.SharingSettings, error) {
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return nil, err
	}

	parentNoteID, err := uc.sharingRepo.GetParentNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent note id: %w", err)
	}

	targetNoteID := noteID
	if parentNoteID != nil {
		targetNoteID = *parentNoteID
	}

	ownerID, err := uc.sharingRepo.GetNoteOwnerID(ctx, targetNoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note owner: %w", err)
	}

	permissions, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, targetNoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	publicAccessLevel, shareUUID, err := uc.sharingRepo.GetPublicAccess(ctx, targetNoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access: %w", err)
	}

	ownerCollaborator := models.Collaborator{
		PermissionID: 0,
		UserID:       ownerID,
		Role:         models.RoleOwner,
		GrantedBy:    ownerID,
		GrantedAt:    time.Now(),
	}

	collaborators := make([]models.Collaborator, 0, len(permissions)+1)
	collaborators = append(collaborators, ownerCollaborator)

	for _, p := range permissions {
		collaborators = append(collaborators, models.Collaborator{
			PermissionID: p.PermissionID,
			UserID:       p.GrantedTo,
			Role:         p.Role,
			GrantedBy:    p.GrantedBy,
			GrantedAt:    p.CreatedAt,
		})
	}

	owner := models.NoteOwner{
		UserID: ownerID,
	}

	publicAccess := models.PublicAccess{
		NoteID:      targetNoteID,
		AccessLevel: publicAccessLevel,
		ShareURL:    shareUUID,
	}

	isOwner := (currentUserID == ownerID)

	settings := &models.SharingSettings{
		NoteID:        noteID,
		Owner:         owner,
		PublicAccess:  publicAccess,
		Collaborators: collaborators,
		TotalCount:    len(collaborators),
		IsOwner:       isOwner,
	}

	return settings, nil
}

func (uc *SharingUsecase) ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error) {
	note, err := uc.notesRepo.GetNoteByShareUUID(ctx, shareUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note by share_uuid: %w", err)
	}

	if note.OwnerID == userID {
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: true,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   true,
				HasAccess: true,
				Role:      models.RoleOwner,
				CanEdit:   true,
			},
		}, nil
	}

	existingPermission, err := uc.sharingRepo.GetUserPermission(ctx, note.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}

	if existingPermission != nil {
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: true,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   false,
				HasAccess: true,
				Role:      existingPermission.Role,
				CanEdit:   existingPermission.Role == models.RoleEditor,
			},
		}, nil
	}

	publicAccessLevel, _, err := uc.sharingRepo.GetPublicAccess(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access level: %w", err)
	}

	if publicAccessLevel == nil {
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: false,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   false,
				HasAccess: false,
			},
		}, nil
	}

	permission := &models.NotePermission{
		NoteID:    note.ID,
		GrantedBy: note.OwnerID,
		GrantedTo: userID,
		Role:      *publicAccessLevel,
	}

	createdPermission, err := uc.sharingRepo.AddCollaborator(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission by link: %w", err)
	}

	if err := uc.updateIsSharedFlag(ctx, note.ID); err != nil {
		return nil, fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return &models.ActivateAccessResponse{
		NoteID:        note.ID,
		AccessGranted: true,
		AccessInfo: models.NoteAccessInfo{
			IsOwner:   false,
			HasAccess: true,
			Role:      createdPermission.Role,
			CanEdit:   createdPermission.Role == models.RoleEditor,
		},
	}, nil
}
