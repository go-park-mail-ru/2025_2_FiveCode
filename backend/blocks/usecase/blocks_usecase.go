package blocksUsecase

import (
	"context"
	"fmt"

	"backend/models"
	namederrors "backend/named_errors"
)

type BlocksUsecase struct {
	BlocksRepo BlocksRepository
	NotesRepo  NotesRepository
}

type BlocksRepository interface {
	CreateBlock(ctx context.Context, noteID uint64, blockType models.BlockType, afterBlockID *uint64) (*models.Block, error)
	GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error)
	GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error)
	UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error)
	DeleteBlock(ctx context.Context, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, blockID uint64, afterBlockID *uint64) (*models.Block, error)
	GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error)
}

type NotesRepository interface {
	GetNoteById(ctx context.Context, noteID uint64) (*models.Note, error)
}

func NewBlocksUsecase(blocksRepo BlocksRepository, notesRepo NotesRepository) *BlocksUsecase {
	return &BlocksUsecase{
		BlocksRepo: blocksRepo,
		NotesRepo:  notesRepo,
	}
}

func (u *BlocksUsecase) CreateBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	block, err := u.BlocksRepo.CreateBlock(ctx, noteID, models.BlockTypeText, beforeBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.BlockWithContent, error) {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	blocks, err := u.BlocksRepo.GetBlocksByNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}

	return blocks, nil
}

func (u *BlocksUsecase) GetBlock(ctx context.Context, userID, blockID uint64) (*models.BlockWithContent, error) {
	block, err := u.BlocksRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if err := u.checkNoteAccess(ctx, userID, block.NoteID); err != nil {
		return nil, err
	}

	return block, nil
}

func (u *BlocksUsecase) UpdateBlock(ctx context.Context, userID, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	block, err := u.BlocksRepo.UpdateBlockText(ctx, blockID, text, formats)
	if err != nil {
		return nil, fmt.Errorf("failed to update block: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return err
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return err
	}

	if err := u.BlocksRepo.DeleteBlock(ctx, blockID); err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	return nil
}

func (u *BlocksUsecase) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	block, err := u.BlocksRepo.UpdateBlockPosition(ctx, blockID, beforeBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) checkNoteAccess(ctx context.Context, userID, noteID uint64) error {
	note, err := u.NotesRepo.GetNoteById(ctx, noteID)
	if err != nil {
		return err
	}

	if note.OwnerID != userID {
		return namederrors.ErrNoAccess
	}

	return nil
}
