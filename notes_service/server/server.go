package server

import (
	"context"

	"backend/notes_service/internal/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -source=server.go -destination=../mock/mock_server.go -package=mock

type NoteUsecase interface {
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID uint64, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, userID uint64, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error
	GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error)
	SearchNotes(ctx context.Context, userID uint64, query string) (*models.SearchNotesResponse, error)
	SetIcon(ctx context.Context, userID, noteID, iconFileID uint64) (*models.Note, error)
	SetHeader(ctx context.Context, userID, noteID, headerFileID uint64) (*models.Note, error)
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
	if note.HeaderFileID != nil {
		protoNote.HeaderFileId = note.HeaderFileID
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

func noteRoleToProto(role models.NoteRole) sharePB.NoteRole {
	switch role {
	case models.RoleOwner:
		return sharePB.NoteRole_NOTE_ROLE_OWNER
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
	case sharePB.NoteRole_NOTE_ROLE_OWNER:
		return models.RoleOwner
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

func noteAccessInfoModelToProto(accessInfo *models.NoteAccessInfo) *sharePB.NoteAccessResponse {
	return &sharePB.NoteAccessResponse{
		HasAccess:  accessInfo.HasAccess,
		Role:       noteRoleToProto(accessInfo.Role),
		IsOwner:    accessInfo.IsOwner,
		CanEdit:    accessInfo.CanEdit,
		CanComment: accessInfo.Role == models.RoleCommenter || accessInfo.Role == models.RoleEditor,
	}
}

func searchResultModelToProto(result *models.SearchResult) *notePB.SearchResult {
	if result == nil {
		return nil
	}

	return &notePB.SearchResult{
		NoteId:           result.NoteID,
		Title:            result.Title,
		HighlightedTitle: result.HighlightedTitle,
		ContentSnippet:   result.ContentSnippet,
		Rank:             result.Rank,
		UpdatedAt:        timestamppb.New(result.UpdatedAt),
		IconFileId:       result.IconFileID,
	}
}

func searchResponseModelToProto(response *models.SearchNotesResponse) *notePB.SearchNotesResponse {
	if response == nil {
		return &notePB.SearchNotesResponse{
			Results: []*notePB.SearchResult{},
			Count:   0,
		}
	}

	protoResults := make([]*notePB.SearchResult, len(response.Results))
	for i := range response.Results {
		protoResults[i] = searchResultModelToProto(&response.Results[i])
	}

	return &notePB.SearchNotesResponse{
		Results: protoResults,
		Count:   int32(response.Count),
	}
}
