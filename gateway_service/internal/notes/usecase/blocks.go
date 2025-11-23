package usecase

import (
	blockPB "backend/notes_service/pkg/block/v1"
	"context"
)

func (u *NotesUsecase) CreateTextBlock(ctx context.Context, req *blockPB.CreateTextBlockRequest) (*blockPB.Block, error) {
	return u.repo.CreateTextBlock(ctx, req)
}

func (u *NotesUsecase) CreateCodeBlock(ctx context.Context, req *blockPB.CreateCodeBlockRequest) (*blockPB.Block, error) {
	return u.repo.CreateCodeBlock(ctx, req)
}

func (u *NotesUsecase) CreateAttachmentBlock(ctx context.Context, req *blockPB.CreateAttachmentBlockRequest) (*blockPB.Block, error) {
	return u.repo.CreateAttachmentBlock(ctx, req)
}

func (u *NotesUsecase) GetBlocks(ctx context.Context, userID, noteID uint64) (*blockPB.GetBlocksResponse, error) {
	return u.repo.GetBlocks(ctx, userID, noteID)
}

func (u *NotesUsecase) GetBlock(ctx context.Context, userID, blockID uint64) (*blockPB.Block, error) {
	return u.repo.GetBlock(ctx, userID, blockID)
}

func (u *NotesUsecase) UpdateBlock(ctx context.Context, req *blockPB.UpdateBlockRequest) (*blockPB.Block, error) {
	return u.repo.UpdateBlock(ctx, req)
}

func (u *NotesUsecase) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	return u.repo.DeleteBlock(ctx, userID, blockID)
}

func (u *NotesUsecase) UpdateBlockPosition(ctx context.Context, req *blockPB.UpdateBlockPositionRequest) (*blockPB.Block, error) {
	return u.repo.UpdateBlockPosition(ctx, req)
}
