package usecase

import (
	"backend/gateway_service/internal/notes/models"
	userModels "backend/gateway_service/internal/user/models"
	"context"
)

type NotesRepository interface {
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, input *models.UpdateNoteInput) (*models.Note, error)
	DeleteNote(ctx context.Context, userID, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error

	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error)
	CreateTextBlock(ctx context.Context, input *models.CreateTextBlockInput) (*models.Block, error)
	CreateCodeBlock(ctx context.Context, input *models.CreateCodeBlockInput) (*models.Block, error)
	CreateAttachmentBlock(ctx context.Context, input *models.CreateAttachmentBlockInput) (*models.Block, error)
	UpdateBlock(ctx context.Context, userID uint64, input *models.UpdateBlockInput) (*models.Block, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error)

	AddCollaborator(ctx context.Context, currentUserID, noteID, targetUserID uint64, role models.NoteRole) (*models.CollaboratorResponse, error)
	GetCollaborators(ctx context.Context, currentUserID, noteID uint64) (*models.GetCollaboratorsResponse, error)
	UpdateCollaboratorRole(ctx context.Context, input *models.UpdateCollaboratorRoleInput) (*models.CollaboratorResponse, error)
	RemoveCollaborator(ctx context.Context, currentUserID, noteID, permissionID uint64) error
	SetPublicAccess(ctx context.Context, input *models.SetPublicAccessInput) (*models.PublicAccessResponse, error)
	GetPublicAccess(ctx context.Context, currentUserID, noteID uint64) (*models.PublicAccessResponse, error)
	GetSharingSettings(ctx context.Context, currentUserID, noteID uint64) (*models.SharingSettingsResponse, error)
	ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error)
}

type UserRepository interface {
	GetUserIDByEmail(ctx context.Context, email string) (uint64, error)
	GetUser(ctx context.Context, userID uint64) (*userModels.User, error)
}

type NotesUsecase struct {
	repo     NotesRepository
	userRepo UserRepository
}

func NewNotesUsecase(repo NotesRepository, userRepo UserRepository) *NotesUsecase {
	return &NotesUsecase{
		repo:     repo,
		userRepo: userRepo,
	}
}
