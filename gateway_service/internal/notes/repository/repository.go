package repository

import (
	fileModels "backend/gateway_service/internal/file/models"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/utils"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source=repository.go -destination=mock/mock_client.go -package=mock
type FileRepository interface {
	GetFileByID(ctx context.Context, fileID uint64) (*fileModels.File, error)
}

type NoteClient interface {
	GetAllNotes(ctx context.Context, in *notePB.GetAllNotesRequest, opts ...grpc.CallOption) (*notePB.GetAllNotesResponse, error)
	CreateNote(ctx context.Context, in *notePB.CreateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	GetNoteById(ctx context.Context, in *notePB.GetNoteByIdRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	UpdateNote(ctx context.Context, in *notePB.UpdateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	DeleteNote(ctx context.Context, in *notePB.DeleteNoteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AddFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RemoveFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SearchNotes(ctx context.Context, in *notePB.SearchNotesRequest, opts ...grpc.CallOption) (*notePB.SearchNotesResponse, error)
	SetIcon(ctx context.Context, in *notePB.SetIconRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	SetHeader(ctx context.Context, in *notePB.SetHeaderRequest, opts ...grpc.CallOption) (*notePB.Note, error)
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

type SharingClient interface {
	AddCollaborator(ctx context.Context, in *sharePB.AddCollaboratorRequest, opts ...grpc.CallOption) (*sharePB.CollaboratorResponse, error)
	GetCollaborators(ctx context.Context, in *sharePB.GetCollaboratorsRequest, opts ...grpc.CallOption) (*sharePB.GetCollaboratorsResponse, error)
	UpdateCollaboratorRole(ctx context.Context, in *sharePB.UpdateCollaboratorRoleRequest, opts ...grpc.CallOption) (*sharePB.CollaboratorResponse, error)
	RemoveCollaborator(ctx context.Context, in *sharePB.RemoveCollaboratorRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SetPublicAccess(ctx context.Context, in *sharePB.SetPublicAccessRequest, opts ...grpc.CallOption) (*sharePB.PublicAccessResponse, error)
	GetPublicAccess(ctx context.Context, in *sharePB.GetPublicAccessRequest, opts ...grpc.CallOption) (*sharePB.PublicAccessResponse, error)
	GetSharingSettings(ctx context.Context, in *sharePB.GetSharingSettingsRequest, opts ...grpc.CallOption) (*sharePB.SharingSettingsResponse, error)
	CheckNoteAccess(ctx context.Context, in *sharePB.CheckNoteAccessRequest, opts ...grpc.CallOption) (*sharePB.NoteAccessResponse, error)
	ActivateAccessByLink(ctx context.Context, in *sharePB.ActivateAccessByLinkRequest, opts ...grpc.CallOption) (*sharePB.ActivateAccessByLinkResponse, error)
}

type NotesRepository struct {
	noteClient    NoteClient
	blockClient   BlockClient
	sharingClient SharingClient
	fileRepo      FileRepository
}

func NewNotesRepository(n NoteClient, b BlockClient, s SharingClient, f FileRepository) *NotesRepository {
	return &NotesRepository{
		noteClient:    n,
		blockClient:   b,
		sharingClient: s,
		fileRepo:      f,
	}
}

func (r *NotesRepository) GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	resp, err := r.noteClient.GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	notes := make([]models.Note, len(resp.Notes))
	for i, pNote := range resp.Notes {
		note := utils.MapProtoToNote(pNote)
		r.enrichNoteWithIcon(ctx, note, pNote.IconFileId)
		r.enrichNoteWithHeader(ctx, note, pNote.HeaderFileId)
		notes[i] = *note
	}
	return notes, nil
}

func (r *NotesRepository) CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error) {
	req := &notePB.CreateNoteRequest{
		UserId:       userID,
		ParentNoteId: parentNoteID,
	}

	resp, err := r.noteClient.CreateNote(ctx, req)
	if err != nil {
		return nil, err
	}
	note := utils.MapProtoToNote(resp)
	r.enrichNoteWithIcon(ctx, note, resp.IconFileId)
	r.enrichNoteWithHeader(ctx, note, resp.HeaderFileId)
	return note, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	resp, err := r.noteClient.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
	if err != nil {
		return nil, err
	}
	note := utils.MapProtoToNote(resp)
	r.enrichNoteWithIcon(ctx, note, resp.IconFileId)
	r.enrichNoteWithHeader(ctx, note, resp.HeaderFileId)
	return note, nil
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
	note := utils.MapProtoToNote(resp)
	r.enrichNoteWithIcon(ctx, note, resp.IconFileId)
	r.enrichNoteWithHeader(ctx, note, resp.HeaderFileId)
	return note, nil
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

func (r *NotesRepository) SearchNotes(ctx context.Context, userID uint64, query string) (*models.SearchNotesResponse, error) {
	searchResult, err := r.noteClient.SearchNotes(ctx, &notePB.SearchNotesRequest{
		UserId: userID,
		Query:  query,
	})
	if err != nil {
		return nil, err
	}

	return utils.MapProtoToSearchNotesResponse(searchResult), nil
}

func (r *NotesRepository) SetIcon(ctx context.Context, userID, noteID, iconFileID uint64) (*models.Note, error) {
	resp, err := r.noteClient.SetIcon(ctx, &notePB.SetIconRequest{
		UserId:     userID,
		NoteId:     noteID,
		IconFileId: iconFileID,
	})
	if err != nil {
		return nil, err
	}

	note := utils.MapProtoToNote(resp)
	r.enrichNoteWithIcon(ctx, note, resp.IconFileId)
	r.enrichNoteWithHeader(ctx, note, resp.HeaderFileId)
	return note, nil
}

func (r *NotesRepository) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error) {
	resp, err := r.blockClient.GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID})
	if err != nil {
		return nil, err
	}

	blocks := make([]models.Block, len(resp.Blocks))
	for i, pbBlock := range resp.Blocks {
		blocks[i] = *utils.MapProtoToBlock(pbBlock)
		r.enrichBlockWithFile(ctx, &blocks[i])
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
	block := utils.MapProtoToBlock(resp)
	r.enrichBlockWithFile(ctx, block)
	return block, nil
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

func (r *NotesRepository) AddCollaborator(ctx context.Context, currentUserID, noteID, targetUserID uint64, role models.NoteRole) (*models.CollaboratorResponse, error) {
	resp, err := r.sharingClient.AddCollaborator(ctx, &sharePB.AddCollaboratorRequest{
		CurrentUserId: currentUserID,
		NoteId:        noteID,
		UserId:        targetUserID,
		Role:          utils.MapModelRoleToProto(role),
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToCollaboratorResponse(resp), nil
}

func (r *NotesRepository) GetCollaborators(ctx context.Context, currentUserID, noteID uint64) (*models.GetCollaboratorsResponse, error) {
	resp, err := r.sharingClient.GetCollaborators(ctx, &sharePB.GetCollaboratorsRequest{
		CurrentUserId: currentUserID,
		NoteId:        noteID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToGetCollaboratorsResponse(resp), nil
}

func (r *NotesRepository) UpdateCollaboratorRole(ctx context.Context, input *models.UpdateCollaboratorRoleInput) (*models.CollaboratorResponse, error) {
	resp, err := r.sharingClient.UpdateCollaboratorRole(ctx, &sharePB.UpdateCollaboratorRoleRequest{
		CurrentUserId: input.CurrentUserID,
		NoteId:        input.NoteID,
		PermissionId:  input.PermissionID,
		NewRole:       utils.MapModelRoleToProto(input.NewRole),
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToCollaboratorResponse(resp), nil
}

func (r *NotesRepository) RemoveCollaborator(ctx context.Context, currentUserID, noteID, permissionID uint64) error {
	_, err := r.sharingClient.RemoveCollaborator(ctx, &sharePB.RemoveCollaboratorRequest{
		CurrentUserId: currentUserID,
		NoteId:        noteID,
		PermissionId:  permissionID,
	})
	return err
}

func (r *NotesRepository) SetPublicAccess(ctx context.Context, input *models.SetPublicAccessInput) (*models.PublicAccessResponse, error) {
	var accessLevel *sharePB.NoteRole
	if input.AccessLevel != nil {
		level := utils.MapModelRoleToProto(*input.AccessLevel)
		accessLevel = &level
	}

	resp, err := r.sharingClient.SetPublicAccess(ctx, &sharePB.SetPublicAccessRequest{
		CurrentUserId: input.CurrentUserID,
		NoteId:        input.NoteID,
		AccessLevel:   accessLevel,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToPublicAccessResponse(resp), nil
}

func (r *NotesRepository) GetPublicAccess(ctx context.Context, currentUserID, noteID uint64) (*models.PublicAccessResponse, error) {
	resp, err := r.sharingClient.GetPublicAccess(ctx, &sharePB.GetPublicAccessRequest{
		CurrentUserId: currentUserID,
		NoteId:        noteID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToPublicAccessResponse(resp), nil
}

func (r *NotesRepository) GetSharingSettings(ctx context.Context, currentUserID, noteID uint64) (*models.SharingSettingsResponse, error) {
	resp, err := r.sharingClient.GetSharingSettings(ctx, &sharePB.GetSharingSettingsRequest{
		CurrentUserId: currentUserID,
		NoteId:        noteID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToSharingSettingsResponse(resp), nil
}

func (r *NotesRepository) ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error) {
	resp, err := r.sharingClient.ActivateAccessByLink(ctx, &sharePB.ActivateAccessByLinkRequest{
		ShareUuid: shareUUID,
		UserId:    userID,
	})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToActivateAccessResponse(resp), nil
}

func (r *NotesRepository) enrichBlockWithFile(ctx context.Context, block *models.Block) {
	if block.Type != models.BlockTypeAttachment {
		return
	}

	content, ok := block.Content.(models.AttachmentContent)
	if !ok {
		return
	}

	if strings.HasPrefix(content.URL, "file:") {
		idStr := strings.TrimPrefix(content.URL, "file:")
		fileID, err := strconv.ParseUint(idStr, 10, 64)
		if err == nil {
			file, err := r.fileRepo.GetFileByID(ctx, fileID)
			if err == nil {
				content.URL = utils.TransformMinioURL(file.URL)
				content.MimeType = file.MimeType
				content.SizeBytes = int(file.SizeBytes)
				content.Width = file.Width
				content.Height = file.Height
				block.Content = content
			}
		}
	}
}

func (r *NotesRepository) enrichNoteWithIcon(ctx context.Context, note *models.Note, iconFileID *uint64) {
	if iconFileID == nil {
		return
	}

	file, err := r.fileRepo.GetFileByID(ctx, *iconFileID)
	if err == nil {
		urlParts := strings.Split(file.URL, "/")
		iconName := urlParts[len(urlParts)-1]

		note.Icon = &models.Icon{
			ID:   file.ID,
			Name: iconName,
			URL:  utils.TransformMinioURL(file.URL),
		}
	}
}

func (r *NotesRepository) enrichNoteWithHeader(ctx context.Context, note *models.Note, headerFileID *uint64) {
	if headerFileID == nil {
		return
	}

	file, err := r.fileRepo.GetFileByID(ctx, *headerFileID)
	if err == nil {
		urlParts := strings.Split(file.URL, "/")
		headerName := urlParts[len(urlParts)-1]

		note.Header = &models.Header{
			ID:   file.ID,
			Name: headerName,
			URL:  utils.TransformMinioURL(file.URL),
		}
	}
}

func (r *NotesRepository) SetHeader(ctx context.Context, userID, noteID, headerFileID uint64) (*models.Note, error) {
	resp, err := r.noteClient.SetHeader(ctx, &notePB.SetHeaderRequest{
		UserId:       userID,
		NoteId:       noteID,
		HeaderFileId: headerFileID,
	})
	if err != nil {
		return nil, err
	}

	note := utils.MapProtoToNote(resp)
	r.enrichNoteWithIcon(ctx, note, resp.IconFileId)
	r.enrichNoteWithHeader(ctx, note, resp.HeaderFileId)
	return note, nil
}
