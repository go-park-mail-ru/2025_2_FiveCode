package repository

import (
	"context"

	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Интерфейсы клиентов
type NoteClient interface {
	GetAllNotes(ctx context.Context, in *notePB.GetAllNotesRequest, opts ...grpc.CallOption) (*notePB.GetAllNotesResponse, error)
	CreateNote(ctx context.Context, in *notePB.CreateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	GetNoteById(ctx context.Context, in *notePB.GetNoteByIdRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	UpdateNote(ctx context.Context, in *notePB.UpdateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	DeleteNote(ctx context.Context, in *notePB.DeleteNoteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AddFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RemoveFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type BlockClient interface {
	GetBlocks(ctx context.Context, in *blockPB.GetBlocksRequest, opts ...grpc.CallOption) (*blockPB.GetBlocksResponse, error)
	GetBlock(ctx context.Context, in *blockPB.GetBlockRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
	CreateTextBlock(ctx context.Context, in *blockPB.CreateTextBlockRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
	CreateCodeBlock(ctx context.Context, in *blockPB.CreateCodeBlockRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
	CreateAttachmentBlock(ctx context.Context, in *blockPB.CreateAttachmentBlockRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
	UpdateBlock(ctx context.Context, in *blockPB.UpdateBlockRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
	DeleteBlock(ctx context.Context, in *blockPB.DeleteBlockRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UpdateBlockPosition(ctx context.Context, in *blockPB.UpdateBlockPositionRequest, opts ...grpc.CallOption) (*blockPB.Block, error)
}

type NotesRepository struct {
	noteClient  NoteClient
	blockClient BlockClient
}

func NewNotesRepository(n NoteClient, b BlockClient) *NotesRepository {
	return &NotesRepository{
		noteClient:  n,
		blockClient: b,
	}
}

// Note Methods

func (r *NotesRepository) GetAllNotes(ctx context.Context, userID uint64) (*notePB.GetAllNotesResponse, error) {
	return r.noteClient.GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID})
}

func (r *NotesRepository) CreateNote(ctx context.Context, userID uint64) (*notePB.Note, error) {
	return r.noteClient.CreateNote(ctx, &notePB.CreateNoteRequest{UserId: userID})
}

func (r *NotesRepository) GetNoteById(ctx context.Context, userID, noteID uint64) (*notePB.Note, error) {
	return r.noteClient.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
}

func (r *NotesRepository) UpdateNote(ctx context.Context, req *notePB.UpdateNoteRequest) (*notePB.Note, error) {
	return r.noteClient.UpdateNote(ctx, req)
}

func (r *NotesRepository) DeleteNote(ctx context.Context, userID, noteID uint64) error {
	_, err := r.noteClient.DeleteNote(ctx, &notePB.DeleteNoteRequest{UserId: userID, NoteId: noteID})
	return err
}

func (r *NotesRepository) AddFavorite(ctx context.Context, userID, noteID uint64) error {
	_, err := r.noteClient.AddFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
	return err
}

func (r *NotesRepository) RemoveFavorite(ctx context.Context, userID, noteID uint64) error {
	_, err := r.noteClient.RemoveFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
	return err
}

// Block Methods

func (r *NotesRepository) GetBlocks(ctx context.Context, userID, noteID uint64) (*blockPB.GetBlocksResponse, error) {
	return r.blockClient.GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID})
}

func (r *NotesRepository) GetBlock(ctx context.Context, userID, blockID uint64) (*blockPB.Block, error) {
	return r.blockClient.GetBlock(ctx, &blockPB.GetBlockRequest{UserId: userID, BlockId: blockID})
}

func (r *NotesRepository) CreateTextBlock(ctx context.Context, req *blockPB.CreateTextBlockRequest) (*blockPB.Block, error) {
	return r.blockClient.CreateTextBlock(ctx, req)
}

func (r *NotesRepository) CreateCodeBlock(ctx context.Context, req *blockPB.CreateCodeBlockRequest) (*blockPB.Block, error) {
	return r.blockClient.CreateCodeBlock(ctx, req)
}

func (r *NotesRepository) CreateAttachmentBlock(ctx context.Context, req *blockPB.CreateAttachmentBlockRequest) (*blockPB.Block, error) {
	return r.blockClient.CreateAttachmentBlock(ctx, req)
}

func (r *NotesRepository) UpdateBlock(ctx context.Context, req *blockPB.UpdateBlockRequest) (*blockPB.Block, error) {
	return r.blockClient.UpdateBlock(ctx, req)
}

func (r *NotesRepository) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	_, err := r.blockClient.DeleteBlock(ctx, &blockPB.DeleteBlockRequest{UserId: userID, BlockId: blockID})
	return err
}

func (r *NotesRepository) UpdateBlockPosition(ctx context.Context, req *blockPB.UpdateBlockPositionRequest) (*blockPB.Block, error) {
	return r.blockClient.UpdateBlockPosition(ctx, req)
}
