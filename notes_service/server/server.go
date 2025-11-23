package server

import (
	"context"
	"errors"

	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -source=server.go -destination=../mock/mock_server.go -package=mock
type NoteUsecase interface {
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID uint64, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, userID uint64, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error
}

type BlocksUsecase interface {
	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error)
	UpdateBlock(ctx context.Context, userID uint64, req *models.UpdateBlockRequest) (*models.Block, error)
	CreateTextBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error)
	CreateCodeBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error)
	CreateAttachmentBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64, fileID uint64) (*models.Block, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error)
}

type Server struct {
	notePB.UnimplementedNoteServiceServer
	blockPB.UnimplementedBlockServiceServer

	noteUsecase   NoteUsecase
	blocksUsecase BlocksUsecase
}

func NewServer(noteUC NoteUsecase, blocksUC BlocksUsecase) *Server {
	return &Server{
		noteUsecase:   noteUC,
		blocksUsecase: blocksUC,
	}
}

func RegisterServices(grpcServer *grpc.Server, noteUC NoteUsecase, blocksUC BlocksUsecase) {
	server := NewServer(noteUC, blocksUC)
	notePB.RegisterNoteServiceServer(grpcServer, server)
	blockPB.RegisterBlockServiceServer(grpcServer, server)
}

func noteModelToProto(note *models.Note) *notePB.Note {
	if note == nil {
		return nil
	}

	protoNote := &notePB.Note{
		Id:         note.ID,
		OwnerId:    note.OwnerID,
		Title:      note.Title,
		IsFavorite: note.IsFavorite,
		IsArchived: note.IsArchived,
		IsShared:   note.IsShared,
		CreatedAt:  timestamppb.New(note.CreatedAt),
		UpdatedAt:  timestamppb.New(note.UpdatedAt),
	}

	if note.ParentNoteID != nil {
		protoNote.ParentNoteId = note.ParentNoteID
	}
	if note.IconFileID != nil {
		protoNote.IconFileId = note.IconFileID
	}
	if note.DeletedAt != nil {
		protoNote.DeletedAt = timestamppb.New(*note.DeletedAt)
	}

	return protoNote
}

func blockModelToProto(block *models.Block) *blockPB.Block {
	if block == nil {
		return nil
	}

	protoBlock := &blockPB.Block{
		Id:        block.ID,
		NoteId:    block.NoteID,
		Type:      block.Type,
		Position:  block.Position,
		CreatedAt: timestamppb.New(block.CreatedAt),
		UpdatedAt: timestamppb.New(block.UpdatedAt),
	}

	switch content := block.Content.(type) {
	case models.TextContent:
		protoBlock.Content = &blockPB.Block_TextContent{
			TextContent: textContentToProto(content),
		}
	case models.CodeContent:
		protoBlock.Content = &blockPB.Block_CodeContent{
			CodeContent: codeContentToProto(content),
		}
	case models.AttachmentContent:
		protoBlock.Content = &blockPB.Block_AttachmentContent{
			AttachmentContent: attachmentContentToProto(content),
		}
	}

	return protoBlock
}

func textContentToProto(content models.TextContent) *blockPB.TextContent {
	formats := make([]*blockPB.BlockTextFormat, len(content.Formats))
	for i, f := range content.Formats {
		formats[i] = &blockPB.BlockTextFormat{
			Id:            f.ID,
			StartOffset:   int32(f.StartOffset),
			EndOffset:     int32(f.EndOffset),
			Bold:          f.Bold,
			Italic:        f.Italic,
			Underline:     f.Underline,
			Strikethrough: f.Strikethrough,
			Font:          string(f.Font),
			Size:          int32(f.Size),
		}
		if f.Link != nil {
			formats[i].Link = f.Link
		}
	}

	return &blockPB.TextContent{
		Text:    content.Text,
		Formats: formats,
	}
}

func codeContentToProto(content models.CodeContent) *blockPB.CodeContent {
	return &blockPB.CodeContent{
		Code:     content.Code,
		Language: content.Language,
	}
}

func attachmentContentToProto(content models.AttachmentContent) *blockPB.AttachmentContent {
	protoContent := &blockPB.AttachmentContent{
		Url:       content.URL,
		MimeType:  content.MimeType,
		SizeBytes: int32(content.SizeBytes),
	}

	if content.Caption != nil {
		protoContent.Caption = content.Caption
	}
	if content.Width != nil {
		width := int32(*content.Width)
		protoContent.Width = &width
	}
	if content.Height != nil {
		height := int32(*content.Height)
		protoContent.Height = &height
	}

	return protoContent
}

func protoToTextContent(protoContent *blockPB.TextContent) models.UpdateTextContent {
	formats := make([]models.BlockTextFormat, len(protoContent.Formats))
	for i, f := range protoContent.Formats {
		formats[i] = models.BlockTextFormat{
			ID:            f.Id,
			StartOffset:   int(f.StartOffset),
			EndOffset:     int(f.EndOffset),
			Bold:          f.Bold,
			Italic:        f.Italic,
			Underline:     f.Underline,
			Strikethrough: f.Strikethrough,
			Font:          models.TextFont(f.Font),
			Size:          int(f.Size),
		}
		if f.Link != nil {
			formats[i].Link = f.Link
		}
	}

	return models.UpdateTextContent{
		Text:    protoContent.Text,
		Formats: formats,
	}
}

func protoToCodeContent(protoContent *blockPB.CodeContent) models.UpdateCodeContent {
	return models.UpdateCodeContent{
		Code:     protoContent.Code,
		Language: protoContent.Language,
	}
}

func (s *Server) GetAllNotes(ctx context.Context, req *notePB.GetAllNotesRequest) (*notePB.GetAllNotesResponse, error) {
	notes, err := s.noteUsecase.GetAllNotes(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notes")
	}

	protoNotes := make([]*notePB.Note, len(notes))
	for i := range notes {
		protoNotes[i] = noteModelToProto(&notes[i])
	}

	return &notePB.GetAllNotesResponse{
		Notes: protoNotes,
	}, nil
}

func (s *Server) CreateNote(ctx context.Context, req *notePB.CreateNoteRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.CreateNote(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) GetNoteById(ctx context.Context, req *notePB.GetNoteByIdRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.GetNoteById(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) UpdateNote(ctx context.Context, req *notePB.UpdateNoteRequest) (*notePB.Note, error) {
	var title *string
	var isArchived *bool

	if req.Title != nil {
		title = req.Title
	}
	if req.IsArchived != nil {
		isArchived = req.IsArchived
	}

	note, err := s.noteUsecase.UpdateNote(ctx, req.GetUserId(), req.GetNoteId(), title, isArchived)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) DeleteNote(ctx context.Context, req *notePB.DeleteNoteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.DeleteNote(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to delete note")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) AddFavorite(ctx context.Context, req *notePB.FavoriteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.AddFavorite(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to add favorite")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveFavorite(ctx context.Context, req *notePB.FavoriteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.RemoveFavorite(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to remove favorite")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetBlocks(ctx context.Context, req *blockPB.GetBlocksRequest) (*blockPB.GetBlocksResponse, error) {
	blocks, err := s.blocksUsecase.GetBlocks(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get blocks")
	}

	protoBlocks := make([]*blockPB.Block, len(blocks))
	for i := range blocks {
		protoBlocks[i] = blockModelToProto(&blocks[i])
	}

	return &blockPB.GetBlocksResponse{
		Blocks: protoBlocks,
	}, nil
}

func (s *Server) GetBlock(ctx context.Context, req *blockPB.GetBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.GetBlock(ctx, req.GetUserId(), req.GetBlockId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateTextBlock(ctx context.Context, req *blockPB.CreateTextBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateTextBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create text block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateCodeBlock(ctx context.Context, req *blockPB.CreateCodeBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateCodeBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create code block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateAttachmentBlock(ctx context.Context, req *blockPB.CreateAttachmentBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateAttachmentBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId, req.GetFileId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create attachment block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) UpdateBlock(ctx context.Context, req *blockPB.UpdateBlockRequest) (*blockPB.Block, error) {
	// Конвертируем proto request в models.UpdateBlockRequest
	updateReq := &models.UpdateBlockRequest{
		BlockID: req.GetBlockId(),
		Type:    req.GetType(),
	}

	// Конвертируем content
	switch content := req.Content.(type) {
	case *blockPB.UpdateBlockRequest_TextContent:
		updateReq.Content = protoToTextContent(content.TextContent)
	case *blockPB.UpdateBlockRequest_CodeContent:
		updateReq.Content = protoToCodeContent(content.CodeContent)
	}

	block, err := s.blocksUsecase.UpdateBlock(ctx, req.GetUserId(), updateReq)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) DeleteBlock(ctx context.Context, req *blockPB.DeleteBlockRequest) (*emptypb.Empty, error) {
	err := s.blocksUsecase.DeleteBlock(ctx, req.GetUserId(), req.GetBlockId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to delete block")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateBlockPosition(ctx context.Context, req *blockPB.UpdateBlockPositionRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.UpdateBlockPosition(ctx, req.GetUserId(), req.GetBlockId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update block position")
	}

	return blockModelToProto(block), nil
}
