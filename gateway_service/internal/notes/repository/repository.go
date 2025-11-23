package repository

import (
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/utils"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

// --- Note Methods ---

func (r *NotesRepository) GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	resp, err := r.noteClient.GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	notes := make([]models.Note, len(resp.Notes))
	for i, pNote := range resp.Notes {
		notes[i] = *utils.MapProtoToNote(pNote)
	}
	return notes, nil
}

func (r *NotesRepository) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) {
	resp, err := r.noteClient.CreateNote(ctx, &notePB.CreateNoteRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToNote(resp), nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	resp, err := r.noteClient.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToNote(resp), nil
}

func (r *NotesRepository) UpdateNote(ctx context.Context, input *models.UpdateNoteInput) (*models.Note, error) {
	req := &notePB.UpdateNoteRequest{
		UserId: input.UserID,
		NoteId: input.ID,
	}
	if input.Title != nil {
		req.Title = input.Title
	}
	if input.IsArchived != nil {
		req.IsArchived = input.IsArchived
	}

	resp, err := r.noteClient.UpdateNote(ctx, req)
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToNote(resp), nil
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

// --- Block Methods ---

func (r *NotesRepository) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error) {
	resp, err := r.blockClient.GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID})
	if err != nil {
		return nil, err
	}

	blocks := make([]models.Block, len(resp.Blocks))
	for i, pbBlock := range resp.Blocks {
		blocks[i] = *utils.MapProtoToBlock(pbBlock)
	}
	return blocks, nil
}

func (r *NotesRepository) GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error) {
	resp, err := r.blockClient.GetBlock(ctx, &blockPB.GetBlockRequest{UserId: userID, BlockId: blockID})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}

func (r *NotesRepository) CreateTextBlock(ctx context.Context, input *models.CreateTextBlockInput) (*models.Block, error) {
	resp, err := r.blockClient.CreateTextBlock(ctx, &blockPB.CreateTextBlockRequest{
		UserId:        input.UserID,
		NoteId:        input.NoteID,
		BeforeBlockId: input.BeforeBlockID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}

func (r *NotesRepository) CreateCodeBlock(ctx context.Context, input *models.CreateCodeBlockInput) (*models.Block, error) {
	resp, err := r.blockClient.CreateCodeBlock(ctx, &blockPB.CreateCodeBlockRequest{
		UserId:        input.UserID,
		NoteId:        input.NoteID,
		BeforeBlockId: input.BeforeBlockID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}

func (r *NotesRepository) CreateAttachmentBlock(ctx context.Context, input *models.CreateAttachmentBlockInput) (*models.Block, error) {
	resp, err := r.blockClient.CreateAttachmentBlock(ctx, &blockPB.CreateAttachmentBlockRequest{
		UserId:        input.UserID,
		NoteId:        input.NoteID,
		BeforeBlockId: input.BeforeBlockID,
		FileId:        input.FileID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}

func (r *NotesRepository) UpdateBlock(ctx context.Context, userID uint64, input *models.UpdateBlockInput) (*models.Block, error) {
	req := &blockPB.UpdateBlockRequest{
		UserId:  userID,
		BlockId: input.BlockID,
		Type:    input.Type,
	}

	switch input.Type {
	case models.BlockTypeText:
		if content, ok := input.Content.(models.UpdateTextContent); ok {
			req.Content = &blockPB.UpdateBlockRequest_TextContent{
				TextContent: utils.MapModelTextContentToProto(&content),
			}
		}
	case models.BlockTypeCode:
		if content, ok := input.Content.(models.UpdateCodeContent); ok {
			req.Content = &blockPB.UpdateBlockRequest_CodeContent{
				CodeContent: utils.MapModelCodeContentToProto(&content),
			}
		}
	}

	resp, err := r.blockClient.UpdateBlock(ctx, req)
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}

func (r *NotesRepository) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	_, err := r.blockClient.DeleteBlock(ctx, &blockPB.DeleteBlockRequest{UserId: userID, BlockId: blockID})
	return err
}

func (r *NotesRepository) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	resp, err := r.blockClient.UpdateBlockPosition(ctx, &blockPB.UpdateBlockPositionRequest{
		UserId:        userID,
		BlockId:       blockID,
		BeforeBlockId: beforeBlockID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToBlock(resp), nil
}
