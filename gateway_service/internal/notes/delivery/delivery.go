package delivery

import (
	"backend/gateway_service/internal/notes/models"
	"context"
)

type NotesUsecase interface {
	// Notes
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, input *models.UpdateNoteInput) (*models.Note, error)
	DeleteNote(ctx context.Context, userID, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error

	// Blocks
	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error)
	CreateTextBlock(ctx context.Context, input *models.CreateTextBlockInput) (*models.Block, error)
	CreateCodeBlock(ctx context.Context, input *models.CreateCodeBlockInput) (*models.Block, error)
	CreateAttachmentBlock(ctx context.Context, input *models.CreateAttachmentBlockInput) (*models.Block, error)
	UpdateBlock(ctx context.Context, userID uint64, input *models.UpdateBlockInput) (*models.Block, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error)
}

type NotesDelivery struct {
	usecase NotesUsecase
}

func NewNotesDelivery(usecase NotesUsecase) *NotesDelivery {
	return &NotesDelivery{
		usecase: usecase,
	}
}
