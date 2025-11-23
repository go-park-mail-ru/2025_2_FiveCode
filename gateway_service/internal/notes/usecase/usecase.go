package usecase

import (
	"context"

	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
)

type NotesRepository interface {
	// Notes methods
	GetAllNotes(ctx context.Context, userID uint64) (*notePB.GetAllNotesResponse, error)
	CreateNote(ctx context.Context, userID uint64) (*notePB.Note, error)
	GetNoteById(ctx context.Context, userID, noteID uint64) (*notePB.Note, error)
	UpdateNote(ctx context.Context, req *notePB.UpdateNoteRequest) (*notePB.Note, error)
	DeleteNote(ctx context.Context, userID, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error

	// Blocks methods
	GetBlocks(ctx context.Context, userID, noteID uint64) (*blockPB.GetBlocksResponse, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*blockPB.Block, error)
	CreateTextBlock(ctx context.Context, req *blockPB.CreateTextBlockRequest) (*blockPB.Block, error)
	CreateCodeBlock(ctx context.Context, req *blockPB.CreateCodeBlockRequest) (*blockPB.Block, error)
	CreateAttachmentBlock(ctx context.Context, req *blockPB.CreateAttachmentBlockRequest) (*blockPB.Block, error)
	UpdateBlock(ctx context.Context, req *blockPB.UpdateBlockRequest) (*blockPB.Block, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, req *blockPB.UpdateBlockPositionRequest) (*blockPB.Block, error)
}

type NotesUsecase struct {
	repo NotesRepository
}

func NewNotesUsecase(repo NotesRepository) *NotesUsecase {
	return &NotesUsecase{repo: repo}
}
