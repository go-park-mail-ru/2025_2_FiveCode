package repository

import (
	fileModels "backend/gateway_service/internal/file/models"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/notes/repository/mock"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNotesRepository_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	mockBlockClient := mock.NewMockBlockClient(ctrl)
	mockSharingClient := mock.NewMockSharingClient(ctrl)
	mockFileRepo := mock.NewMockFileRepository(ctrl)

	repo := NewNotesRepository(mockNoteClient, mockBlockClient, mockSharingClient, mockFileRepo)

	ctx := context.Background()
	userID := uint64(1)
	now := time.Now()

	protoNotes := []*notePB.Note{
		{
			Id:        1,
			OwnerId:   userID,
			Title:     "Note 1",
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		},
	}
	expectedNotes := []models.Note{
		{
			ID:        1,
			OwnerID:   userID,
			Title:     "Note 1",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID}).Return(&notePB.GetAllNotesResponse{Notes: protoNotes}, nil)

		notes, err := repo.GetAllNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, notes, 1)
		assert.Equal(t, expectedNotes[0].ID, notes[0].ID)
		assert.Equal(t, expectedNotes[0].Title, notes[0].Title)
	})

	t.Run("GrpcError", func(t *testing.T) {
		mockNoteClient.EXPECT().GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID}).Return(nil, status.Error(codes.Internal, "grpc error"))

		_, err := repo.GetAllNotes(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "grpc error")
	})
}

func TestNotesRepository_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	parentNoteID := uint64(5)

	protoNote := &notePB.Note{Id: 1, OwnerId: userID, ParentNoteId: &parentNoteID}

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().CreateNote(ctx, &notePB.CreateNoteRequest{UserId: userID, ParentNoteId: &parentNoteID}).Return(protoNote, nil)
		note, err := repo.CreateNote(ctx, userID, &parentNoteID)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), note.ID)
		assert.Equal(t, parentNoteID, *note.ParentNoteID)
	})
}

func TestNotesRepository_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	protoNote := &notePB.Note{Id: noteID, OwnerId: userID, Title: "Test Note"}

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID}).Return(protoNote, nil)
		note, err := repo.GetNoteById(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, noteID, note.ID)
		assert.Equal(t, "Test Note", note.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockNoteClient.EXPECT().GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID}).Return(nil, status.Error(codes.NotFound, "not found"))
		_, err := repo.GetNoteById(ctx, userID, noteID)
		assert.Error(t, err)
	})
}

func TestNotesRepository_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	title := "New Title"

	input := &models.UpdateNoteInput{UserID: userID, ID: noteID, Title: &title}
	protoNote := &notePB.Note{Id: noteID, OwnerId: userID, Title: title}

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().UpdateNote(ctx, &notePB.UpdateNoteRequest{UserId: userID, NoteId: noteID, Title: &title}).Return(protoNote, nil)
		note, err := repo.UpdateNote(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, title, note.Title)
	})
}

func TestNotesRepository_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().DeleteNote(ctx, &notePB.DeleteNoteRequest{UserId: userID, NoteId: noteID}).Return(&emptypb.Empty{}, nil)
		err := repo.DeleteNote(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_AddFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().AddFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID}).Return(&emptypb.Empty{}, nil)
		err := repo.AddFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_RemoveFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteClient := mock.NewMockNoteClient(ctrl)
	repo := NewNotesRepository(mockNoteClient, nil, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteClient.EXPECT().RemoveFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID}).Return(&emptypb.Empty{}, nil)
		err := repo.RemoveFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	mockFileRepo := mock.NewMockFileRepository(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, mockFileRepo)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success_WithAttachment", func(t *testing.T) {
		protoBlocks := []*blockPB.Block{
			{Id: 100, Type: "text", Content: &blockPB.Block_TextContent{TextContent: &blockPB.TextContent{Text: "Hello"}}},
			{Id: 101, Type: "attachment", Content: &blockPB.Block_AttachmentContent{AttachmentContent: &blockPB.AttachmentContent{Url: "file:999", MimeType: "image/png"}}},
		}

		mockBlockClient.EXPECT().GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID}).Return(&blockPB.GetBlocksResponse{Blocks: protoBlocks}, nil)

		width := 100
		height := 100
		file := &fileModels.File{ID: 999, URL: "http://minio/file.png", MimeType: "image/png", SizeBytes: 1024, Width: &width, Height: &height}
		mockFileRepo.EXPECT().GetFileByID(gomock.Any(), uint64(999)).Return(file, nil)

		blocks, err := repo.GetBlocks(ctx, userID, noteID)
		assert.NoError(t, err)
		assert.Len(t, blocks, 2)
		assert.Equal(t, "text", blocks[0].Type)
		assert.Equal(t, "attachment", blocks[1].Type)

		content := blocks[1].Content.(models.AttachmentContent)
		assert.NotEqual(t, "file:999", content.URL)
	})
}

func TestNotesRepository_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	mockFileRepo := mock.NewMockFileRepository(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, mockFileRepo)

	ctx := context.Background()
	userID := uint64(1)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		protoBlock := &blockPB.Block{Id: blockID, Type: "text", Content: &blockPB.Block_TextContent{TextContent: &blockPB.TextContent{Text: "Hello"}}}
		mockBlockClient.EXPECT().GetBlock(ctx, &blockPB.GetBlockRequest{UserId: userID, BlockId: blockID}).Return(protoBlock, nil)

		block, err := repo.GetBlock(ctx, userID, blockID)
		assert.NoError(t, err)
		assert.Equal(t, blockID, block.ID)
	})
}

func TestNotesRepository_CreateTextBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	input := &models.CreateTextBlockInput{UserID: userID, NoteID: noteID}
	protoBlock := &blockPB.Block{Id: 100, Type: "text"}

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().CreateTextBlock(ctx, &blockPB.CreateTextBlockRequest{UserId: userID, NoteId: noteID}).Return(protoBlock, nil)
		block, err := repo.CreateTextBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), block.ID)
	})
}

func TestNotesRepository_CreateCodeBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	input := &models.CreateCodeBlockInput{UserID: userID, NoteID: noteID}
	protoBlock := &blockPB.Block{Id: 100, Type: "code"}

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().CreateCodeBlock(ctx, &blockPB.CreateCodeBlockRequest{UserId: userID, NoteId: noteID}).Return(protoBlock, nil)
		block, err := repo.CreateCodeBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), block.ID)
	})
}

func TestNotesRepository_CreateAttachmentBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	fileID := uint64(999)

	input := &models.CreateAttachmentBlockInput{UserID: userID, NoteID: noteID, FileID: fileID}
	protoBlock := &blockPB.Block{Id: 100, Type: "attachment"}

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().CreateAttachmentBlock(ctx, &blockPB.CreateAttachmentBlockRequest{UserId: userID, NoteId: noteID, FileId: fileID}).Return(protoBlock, nil)
		block, err := repo.CreateAttachmentBlock(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), block.ID)
	})
}

func TestNotesRepository_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	blockID := uint64(100)

	input := &models.UpdateBlockInput{
		BlockID: blockID,
		Type:    models.BlockTypeText,
		Content: models.UpdateTextContent{Text: "Updated"},
	}

	protoBlock := &blockPB.Block{Id: blockID, Type: "text", Content: &blockPB.Block_TextContent{TextContent: &blockPB.TextContent{Text: "Updated"}}}

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().UpdateBlock(ctx, gomock.Any()).Return(protoBlock, nil)

		block, err := repo.UpdateBlock(ctx, userID, input)
		assert.NoError(t, err)
		assert.Equal(t, "text", block.Type)
	})
}

func TestNotesRepository_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().DeleteBlock(ctx, &blockPB.DeleteBlockRequest{UserId: userID, BlockId: blockID}).Return(&emptypb.Empty{}, nil)
		err := repo.DeleteBlock(ctx, userID, blockID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockClient := mock.NewMockBlockClient(ctrl)
	repo := NewNotesRepository(nil, mockBlockClient, nil, nil)

	ctx := context.Background()
	userID := uint64(1)
	blockID := uint64(100)
	beforeBlockID := uint64(101)

	protoBlock := &blockPB.Block{Id: blockID, Type: "text"}

	t.Run("Success", func(t *testing.T) {
		mockBlockClient.EXPECT().UpdateBlockPosition(ctx, &blockPB.UpdateBlockPositionRequest{UserId: userID, BlockId: blockID, BeforeBlockId: &beforeBlockID}).Return(protoBlock, nil)
		block, err := repo.UpdateBlockPosition(ctx, userID, blockID, &beforeBlockID)
		assert.NoError(t, err)
		assert.Equal(t, blockID, block.ID)
	})
}

func TestNotesRepository_AddCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	targetUserID := uint64(2)
	noteID := uint64(10)
	role := models.RoleEditor

	protoResp := &sharePB.CollaboratorResponse{
		PermissionId: 5,
		Collaborator: &sharePB.Collaborator{UserId: targetUserID, Role: sharePB.NoteRole_NOTE_ROLE_EDITOR},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().AddCollaborator(ctx, gomock.Any()).Return(protoResp, nil)

		resp, err := repo.AddCollaborator(ctx, currentUserID, noteID, targetUserID, role)
		assert.NoError(t, err)
		assert.Equal(t, uint64(5), resp.PermissionID)
	})
}

func TestNotesRepository_GetCollaborators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)

	protoResp := &sharePB.GetCollaboratorsResponse{
		Collaborators: []*sharePB.Collaborator{
			{PermissionId: 5, UserId: 2},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().GetCollaborators(ctx, &sharePB.GetCollaboratorsRequest{NoteId: noteID, CurrentUserId: currentUserID}).Return(protoResp, nil)

		resp, err := repo.GetCollaborators(ctx, currentUserID, noteID)
		assert.NoError(t, err)
		assert.Len(t, resp.Collaborators, 1)
	})
}

func TestNotesRepository_UpdateCollaboratorRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)
	permissionID := uint64(5)
	newRole := models.RoleViewer

	input := &models.UpdateCollaboratorRoleInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		PermissionID:  permissionID,
		NewRole:       newRole,
	}

	protoResp := &sharePB.CollaboratorResponse{
		PermissionId: permissionID,
		Collaborator: &sharePB.Collaborator{Role: sharePB.NoteRole_NOTE_ROLE_VIEWER},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().UpdateCollaboratorRole(ctx, gomock.Any()).Return(protoResp, nil)

		resp, err := repo.UpdateCollaboratorRole(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, permissionID, resp.PermissionID)
	})
}

func TestNotesRepository_RemoveCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)
	permissionID := uint64(5)

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().RemoveCollaborator(ctx, &sharePB.RemoveCollaboratorRequest{NoteId: noteID, CurrentUserId: currentUserID, PermissionId: permissionID}).Return(&emptypb.Empty{}, nil)
		err := repo.RemoveCollaborator(ctx, currentUserID, noteID, permissionID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_SetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)
	role := models.RoleViewer

	input := &models.SetPublicAccessInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		AccessLevel:   &role,
	}

	protoRole := sharePB.NoteRole_NOTE_ROLE_VIEWER
	protoResp := &sharePB.PublicAccessResponse{
		AccessLevel: &protoRole,
		ShareUrl:    "http://link",
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().SetPublicAccess(ctx, gomock.Any()).Return(protoResp, nil)

		resp, err := repo.SetPublicAccess(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "http://link", resp.ShareURL)
	})
}

func TestNotesRepository_GetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)

	protoRole := sharePB.NoteRole_NOTE_ROLE_VIEWER
	protoResp := &sharePB.PublicAccessResponse{
		AccessLevel: &protoRole,
		ShareUrl:    "http://link",
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().GetPublicAccess(ctx, &sharePB.GetPublicAccessRequest{NoteId: noteID, CurrentUserId: currentUserID}).Return(protoResp, nil)

		resp, err := repo.GetPublicAccess(ctx, currentUserID, noteID)
		assert.NoError(t, err)
		assert.Equal(t, "http://link", resp.ShareURL)
	})
}

func TestNotesRepository_GetSharingSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	currentUserID := uint64(1)
	noteID := uint64(10)

	protoRole := sharePB.NoteRole_NOTE_ROLE_VIEWER
	protoResp := &sharePB.SharingSettingsResponse{
		PublicAccess: &sharePB.PublicAccess{AccessLevel: &protoRole},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().GetSharingSettings(ctx, &sharePB.GetSharingSettingsRequest{NoteId: noteID, CurrentUserId: currentUserID}).Return(protoResp, nil)

		resp, err := repo.GetSharingSettings(ctx, currentUserID, noteID)
		assert.NoError(t, err)
		assert.NotNil(t, resp.PublicAccess.AccessLevel)
	})
}

func TestNotesRepository_ActivateAccessByLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingClient := mock.NewMockSharingClient(ctrl)
	repo := NewNotesRepository(nil, nil, mockSharingClient, nil)

	ctx := context.Background()
	userID := uint64(1)
	shareUUID := "uuid"

	protoResp := &sharePB.ActivateAccessByLinkResponse{
		NoteId:     10,
		AccessInfo: &sharePB.NoteAccessResponse{Role: sharePB.NoteRole_NOTE_ROLE_VIEWER},
	}

	t.Run("Success", func(t *testing.T) {
		mockSharingClient.EXPECT().ActivateAccessByLink(ctx, &sharePB.ActivateAccessByLinkRequest{ShareUuid: shareUUID, UserId: userID}).Return(protoResp, nil)

		resp, err := repo.ActivateAccessByLink(ctx, shareUUID, userID)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), resp.NoteID)
	})
}
