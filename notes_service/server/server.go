package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"

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
	GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error)
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

type SharingUsecase interface {
	AddCollaborator(ctx context.Context, noteID, currentUserID, targetUserID uint64, role models.NoteRole) (*models.NotePermission, error)
	GetCollaborators(ctx context.Context, noteID, currentUserID uint64) (uint64, []*models.NotePermission, *models.NoteRole, error)
	UpdateCollaboratorRole(ctx context.Context, noteID, currentUserID, permissionID uint64, newRole models.NoteRole) (*models.NotePermission, error)
	RemoveCollaborator(ctx context.Context, noteID, currentUserID, permissionID uint64) error
	SetPublicAccess(ctx context.Context, noteID, currentUserID uint64, accessLevel *models.NoteRole) error
	GetPublicAccess(ctx context.Context, noteID, currentUserID uint64) (*models.NoteRole, error)
	GetSharingSettings(ctx context.Context, noteID, currentUserID uint64) (*models.SharingSettings, error)
	ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error)
	CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error)
}

type Server struct {
	notePB.UnimplementedNoteServiceServer
	blockPB.UnimplementedBlockServiceServer
	sharePB.UnimplementedSharingServiceServer

	noteUsecase    NoteUsecase
	blocksUsecase  BlocksUsecase
	sharingUsecase SharingUsecase
}

func NewServer(noteUC NoteUsecase, blocksUC BlocksUsecase, sharingUC SharingUsecase) *Server {
	return &Server{
		noteUsecase:    noteUC,
		blocksUsecase:  blocksUC,
		sharingUsecase: sharingUC,
	}
}

func RegisterServices(grpcServer *grpc.Server, noteUC NoteUsecase, blocksUC BlocksUsecase, sharingUC SharingUsecase) {
	server := NewServer(noteUC, blocksUC, sharingUC)
	notePB.RegisterNoteServiceServer(grpcServer, server)
	blockPB.RegisterBlockServiceServer(grpcServer, server)
	sharePB.RegisterSharingServiceServer(grpcServer, server)
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

func (s *Server) GetNoteByShareUUID(ctx context.Context, req *notePB.GetNoteByShareUUIDRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.GetNoteByShareUUID(ctx, req.GetShareUuid())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		return nil, status.Error(codes.Internal, "failed to get note by share uuid")
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

func noteRoleToProto(role models.NoteRole) sharePB.NoteRole {
	switch role {
	case models.RoleViewer:
		return sharePB.NoteRole_NOTE_ROLE_VIEWER
	case models.RoleCommenter:
		return sharePB.NoteRole_NOTE_ROLE_COMMENTER
	case models.RoleEditor:
		return sharePB.NoteRole_NOTE_ROLE_EDITOR
	default:
		return sharePB.NoteRole_NOTE_ROLE_UNSPECIFIED
	}
}

func noteRoleFromProto(role sharePB.NoteRole) models.NoteRole {
	switch role {
	case sharePB.NoteRole_NOTE_ROLE_VIEWER:
		return models.RoleViewer
	case sharePB.NoteRole_NOTE_ROLE_COMMENTER:
		return models.RoleCommenter
	case sharePB.NoteRole_NOTE_ROLE_EDITOR:
		return models.RoleEditor
	default:
		return models.RoleViewer
	}
}

func collaboratorModelToProto(collab *models.Collaborator) *sharePB.Collaborator {
	return &sharePB.Collaborator{
		PermissionId: collab.PermissionID,
		UserId:       collab.UserID,
		Role:         noteRoleToProto(collab.Role),
		GrantedBy:    collab.GrantedBy,
		GrantedAt:    timestamppb.New(collab.GrantedAt),
	}
}

func publicAccessModelToProto(publicAccess *models.PublicAccess) *sharePB.PublicAccess {
	var accessLevel *sharePB.NoteRole
	if publicAccess.AccessLevel != nil {
		level := noteRoleToProto(*publicAccess.AccessLevel)
		accessLevel = &level
	}

	return &sharePB.PublicAccess{
		NoteId:      publicAccess.NoteID,
		AccessLevel: accessLevel,
		ShareUrl:    publicAccess.ShareURL,
	}
}

func noteAccessInfoModelToProto(accessInfo *models.NoteAccessInfo) *sharePB.NoteAccessResponse {
	return &sharePB.NoteAccessResponse{
		HasAccess:  accessInfo.HasAccess,
		Role:       noteRoleToProto(accessInfo.Role),
		IsOwner:    accessInfo.IsOwner,
		CanEdit:    accessInfo.CanEdit,
		CanComment: accessInfo.Role == models.RoleCommenter || accessInfo.Role == models.RoleEditor,
	}
}

func (s *Server) AddCollaborator(ctx context.Context, req *sharePB.AddCollaboratorRequest) (*sharePB.CollaboratorResponse, error) {
	permission, err := s.sharingUsecase.AddCollaborator(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetUserId(),
		noteRoleFromProto(req.GetRole()), // Конвертируем здесь
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to add collaborator")
	}

	collaborator := &models.Collaborator{
		PermissionID: permission.PermissionID,
		UserID:       permission.GrantedTo,
		Role:         permission.Role,
		GrantedBy:    permission.GrantedBy,
		GrantedAt:    permission.CreatedAt,
	}

	return &sharePB.CollaboratorResponse{
		PermissionId: permission.PermissionID,
		Collaborator: collaboratorModelToProto(collaborator),
	}, nil
}

func (s *Server) GetCollaborators(ctx context.Context, req *sharePB.GetCollaboratorsRequest) (*sharePB.GetCollaboratorsResponse, error) {
	ownerID, permissions, publicAccessLevel, err := s.sharingUsecase.GetCollaborators(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get collaborators")
	}

	collaborators := make([]*sharePB.Collaborator, len(permissions))
	for i, p := range permissions {
		collaborators[i] = &sharePB.Collaborator{
			PermissionId: p.PermissionID,
			UserId:       p.GrantedTo,
			Role:         noteRoleToProto(p.Role),
			GrantedBy:    p.GrantedBy,
			GrantedAt:    timestamppb.New(p.CreatedAt),
		}
	}

	var protoPublicAccessLevel *sharePB.NoteRole
	if publicAccessLevel != nil {
		level := noteRoleToProto(*publicAccessLevel)
		protoPublicAccessLevel = &level
	}

	return &sharePB.GetCollaboratorsResponse{
		NoteId:             req.GetNoteId(),
		OwnerId:            ownerID,
		Collaborators:      collaborators,
		PublicAccessLevel:  protoPublicAccessLevel,
		TotalCollaborators: int32(len(collaborators)),
	}, nil
}

func (s *Server) UpdateCollaboratorRole(ctx context.Context, req *sharePB.UpdateCollaboratorRoleRequest) (*sharePB.CollaboratorResponse, error) {
	permission, err := s.sharingUsecase.UpdateCollaboratorRole(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetPermissionId(),
		noteRoleFromProto(req.GetNewRole()),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "permission not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update collaborator role")
	}

	collaborator := &models.Collaborator{
		PermissionID: permission.PermissionID,
		UserID:       permission.GrantedTo,
		Role:         permission.Role,
		GrantedBy:    permission.GrantedBy,
		GrantedAt:    permission.CreatedAt,
	}

	return &sharePB.CollaboratorResponse{
		PermissionId: permission.PermissionID,
		Collaborator: collaboratorModelToProto(collaborator),
	}, nil
}

func (s *Server) SetPublicAccess(ctx context.Context, req *sharePB.SetPublicAccessRequest) (*sharePB.PublicAccessResponse, error) {
	var accessLevel *models.NoteRole
	if req.AccessLevel != nil {
		level := noteRoleFromProto(*req.AccessLevel)
		accessLevel = &level
	}

	err := s.sharingUsecase.SetPublicAccess(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		accessLevel,
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to set public access")
	}

	updatedAccess, err := s.sharingUsecase.GetPublicAccess(ctx, req.GetNoteId(), req.GetCurrentUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get updated public access")
	}

	var protoAccessLevel *sharePB.NoteRole
	if updatedAccess != nil {
		level := noteRoleToProto(*updatedAccess)
		protoAccessLevel = &level
	}

	return &sharePB.PublicAccessResponse{
		NoteId:      req.GetNoteId(),
		AccessLevel: protoAccessLevel,
		ShareUrl:    fmt.Sprintf("/notes/%d", req.GetNoteId()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (s *Server) GetPublicAccess(ctx context.Context, req *sharePB.GetPublicAccessRequest) (*sharePB.PublicAccessResponse, error) {
	accessLevel, err := s.sharingUsecase.GetPublicAccess(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get public access")
	}

	var protoAccessLevel *sharePB.NoteRole
	if accessLevel != nil {
		level := noteRoleToProto(*accessLevel)
		protoAccessLevel = &level
	}

	return &sharePB.PublicAccessResponse{
		NoteId:      req.GetNoteId(),
		AccessLevel: protoAccessLevel,
		ShareUrl:    fmt.Sprintf("/notes/%d", req.GetNoteId()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (s *Server) ActivateAccessByLink(ctx context.Context, req *sharePB.ActivateAccessByLinkRequest) (*sharePB.ActivateAccessByLinkResponse, error) {
	response, err := s.sharingUsecase.ActivateAccessByLink(
		ctx,
		req.GetShareUuid(),
		req.GetUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		return nil, status.Error(codes.Internal, "failed to activate access by link")
	}

	return &sharePB.ActivateAccessByLinkResponse{
		NoteId:        response.NoteID,
		AccessGranted: response.AccessGranted,
		AccessInfo:    noteAccessInfoModelToProto(&response.AccessInfo),
	}, nil
}

func (s *Server) RemoveCollaborator(ctx context.Context, req *sharePB.RemoveCollaboratorRequest) (*emptypb.Empty, error) {
	err := s.sharingUsecase.RemoveCollaborator(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetPermissionId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "permission not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to remove collaborator")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetSharingSettings(ctx context.Context, req *sharePB.GetSharingSettingsRequest) (*sharePB.SharingSettingsResponse, error) {
	settings, err := s.sharingUsecase.GetSharingSettings(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get sharing settings")
	}

	collaborators := make([]*sharePB.Collaborator, len(settings.Collaborators))
	for i, c := range settings.Collaborators {
		collaborators[i] = &sharePB.Collaborator{
			PermissionId: c.PermissionID,
			UserId:       c.UserID,
			Role:         noteRoleToProto(c.Role),
			GrantedBy:    c.GrantedBy,
			GrantedAt:    timestamppb.New(c.GrantedAt),
		}
	}

	var publicAccessLevel *sharePB.NoteRole
	if settings.PublicAccess.AccessLevel != nil {
		level := noteRoleToProto(*settings.PublicAccess.AccessLevel)
		publicAccessLevel = &level
	}

	publicAccess := &sharePB.PublicAccess{
		NoteId:      settings.PublicAccess.NoteID,
		AccessLevel: publicAccessLevel,
		ShareUrl:    settings.PublicAccess.ShareURL,
	}

	return &sharePB.SharingSettingsResponse{
		NoteId:             settings.NoteID,
		OwnerId:            settings.Owner.UserID,
		PublicAccess:       publicAccess,
		Collaborators:      collaborators,
		TotalCollaborators: int32(settings.TotalCount),
		IsOwner:            settings.IsOwner,
	}, nil
}

func (s *Server) CheckNoteAccess(ctx context.Context, req *sharePB.CheckNoteAccessRequest) (*sharePB.NoteAccessResponse, error) {
	accessInfo, err := s.sharingUsecase.CheckNoteAccess(
		ctx,
		req.GetNoteId(),
		req.GetUserId(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check note access")
	}

	return noteAccessInfoModelToProto(accessInfo), nil
}
