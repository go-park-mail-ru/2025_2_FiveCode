package Repository

import (
	"context"
	"errors"

	"backend/models"
	"backend/store"
)

type BlocksRepository struct {
	Store *store.Store
}

func NewBlocksRepository(store *store.Store) *BlocksRepository {
	return &BlocksRepository{
		Store: store,
	}
}

func (r *BlocksRepository) CreateBlock(ctx context.Context, noteID uint64, blockType models.BlockType, beforeBlockID *uint64) (*models.Block, error) {
	block, err := r.Store.CreateBlock(noteID, blockType, beforeBlockID)
	if err != nil {
		return nil, errors.New("failed to create block: " + err.Error())
	}
	return block, nil
}

func (r *BlocksRepository) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error) {
	blocks, err := r.Store.GetBlocksByNoteID(noteID)
	if err != nil {
		return nil, errors.New("failed to get blocks: " + err.Error())
	}
	return blocks, nil
}

func (r *BlocksRepository) GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error) {
	block, err := r.Store.GetBlockByID(blockID)
	if err != nil {
		return nil, errors.New("failed to get block: " + err.Error())
	}
	return block, nil
}

func (r *BlocksRepository) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
	block, err := r.Store.UpdateBlockText(blockID, text, formats)
	if err != nil {
		return nil, errors.New("failed to update block text: " + err.Error())
	}
	return block, nil
}

func (r *BlocksRepository) DeleteBlock(ctx context.Context, blockID uint64) error {
	if err := r.Store.DeleteBlock(blockID); err != nil {
		return errors.New("failed to delete block: " + err.Error())
	}
	return nil
}

func (r *BlocksRepository) UpdateBlockPosition(ctx context.Context, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	block, err := r.Store.UpdateBlockPosition(blockID, beforeBlockID)
	if err != nil {
		return nil, errors.New("failed to update position: " + err.Error())
	}
	return block, nil
}

func (r *BlocksRepository) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) {
	noteID, err := r.Store.GetBlockNoteID(blockID)
	if err != nil {
		return 0, errors.New("failed to get block note id: " + err.Error())
	}
	return noteID, nil
}
