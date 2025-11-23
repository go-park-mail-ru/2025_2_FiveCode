package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"context"
)

func (u *NotesUsecase) CreateTextBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	return u.repo.CreateTextBlock(ctx, userID, noteID, beforeBlockID)
}

func (u *NotesUsecase) CreateCodeBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	return u.repo.CreateCodeBlock(ctx, userID, noteID, beforeBlockID)
}

func (u *NotesUsecase) CreateAttachmentBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64, fileID uint64) (*models.Block, error) {
	return u.repo.CreateAttachmentBlock(ctx, userID, noteID, beforeBlockID, fileID)
}

func (u *NotesUsecase) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error) {
	return u.repo.GetBlocks(ctx, userID, noteID)
}

func (u *NotesUsecase) GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error) {
	return u.repo.GetBlock(ctx, userID, blockID)
}

func (u *NotesUsecase) UpdateBlock(ctx context.Context, userID uint64, input *models.UpdateBlockInput) (*models.Block, error) {
	return u.repo.UpdateBlock(ctx, userID, input)
}

func (u *NotesUsecase) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	return u.repo.DeleteBlock(ctx, userID, blockID)
}

func (u *NotesUsecase) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	return u.repo.UpdateBlockPosition(ctx, userID, blockID, beforeBlockID)
}
