package server

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/mock"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	notes := []models.Note{
		{ID: 1, Title: "Note 1", OwnerID: userID},
		{ID: 2, Title: "Note 2", OwnerID: userID},
	}

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().
			GetAllNotes(ctx, userID).
			Return(notes, nil)

		resp, err := server.GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID})
		assert.NoError(t, err)
		assert.Len(t, resp.Notes, 2)
		assert.Equal(t, "Note 1", resp.Notes[0].Title)
	})

	t.Run("Error", func(t *testing.T) {
		mockNoteUC.EXPECT().
			GetAllNotes(ctx, userID).
			Return(nil, errors.New("db error"))

		resp, err := server.GetAllNotes(ctx, &notePB.GetAllNotesRequest{UserId: userID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	title := "New Note"
	createdNote := &models.Note{ID: 1, Title: title, OwnerID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().
			CreateNote(ctx, userID, nil).
			Return(createdNote, nil)

		resp, err := server.CreateNote(ctx, &notePB.CreateNoteRequest{UserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, createdNote.ID, resp.Id)
	})

	t.Run("ParentNotFound", func(t *testing.T) {
		parentID := uint64(999)
		mockNoteUC.EXPECT().
			CreateNote(ctx, userID, &parentID).
			Return(nil, errors.New("parent note not found"))

		resp, err := server.CreateNote(ctx, &notePB.CreateNoteRequest{UserId: userID, ParentNoteId: &parentID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	note := &models.Note{ID: noteID, Title: "My Note", OwnerID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().
			GetNoteById(ctx, userID, noteID).
			Return(note, nil)

		resp, err := server.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
		assert.Equal(t, noteID, resp.Id)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockNoteUC.EXPECT().
			GetNoteById(ctx, userID, noteID).
			Return(nil, constants.ErrNotFound)

		_, err := server.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("AccessDenied", func(t *testing.T) {
		mockNoteUC.EXPECT().
			GetNoteById(ctx, userID, noteID).
			Return(nil, constants.ErrNoAccess)

		_, err := server.GetNoteById(ctx, &notePB.GetNoteByIdRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	title := "Updated Title"
	updatedNote := &models.Note{ID: noteID, Title: title, OwnerID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().
			UpdateNote(ctx, userID, noteID, &title, nil).
			Return(updatedNote, nil)

		resp, err := server.UpdateNote(ctx, &notePB.UpdateNoteRequest{UserId: userID, NoteId: noteID, Title: &title})
		assert.NoError(t, err)
		assert.Equal(t, title, resp.Title)
	})
}

func TestServer_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().
			DeleteNote(ctx, userID, noteID).
			Return(nil)

		_, err := server.DeleteNote(ctx, &notePB.DeleteNoteRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockNoteUC.EXPECT().DeleteNote(ctx, userID, noteID).Return(constants.ErrNotFound)
		_, err := server.DeleteNote(ctx, &notePB.DeleteNoteRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockNoteUC.EXPECT().DeleteNote(ctx, userID, noteID).Return(constants.ErrNoAccess)
		_, err := server.DeleteNote(ctx, &notePB.DeleteNoteRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	blocks := []models.Block{
		{
			BaseBlock: models.BaseBlock{
				ID:     1,
				NoteID: noteID,
				Type:   "text",
			},
			Content: models.TextContent{Text: "Hello"},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().
			GetBlocks(ctx, userID, noteID).
			Return(blocks, nil)

		resp, err := server.GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
		assert.Len(t, resp.Blocks, 1)
		assert.Equal(t, "text", resp.Blocks[0].Type)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().
			GetBlocks(ctx, userID, noteID).
			Return(nil, constants.ErrNoAccess)

		_, err := server.GetBlocks(ctx, &blockPB.GetBlocksRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_CreateTextBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{
		BaseBlock: models.BaseBlock{
			ID:     100,
			NoteID: noteID,
			Type:   "text",
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().
			CreateTextBlock(ctx, userID, noteID, nil).
			Return(block, nil)

		resp, err := server.CreateTextBlock(ctx, &blockPB.CreateTextBlockRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), resp.Id)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().
			CreateTextBlock(ctx, userID, noteID, nil).
			Return(nil, constants.ErrNoAccess)

		_, err := server.CreateTextBlock(ctx, &blockPB.CreateTextBlockRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_AddCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	noteID := uint64(10)
	currentUserID := uint64(1)
	targetUserID := uint64(2)
	role := models.RoleEditor
	permission := &models.NotePermission{PermissionID: 5, GrantedTo: targetUserID, Role: role, GrantedBy: currentUserID, CreatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().
			AddCollaborator(ctx, noteID, currentUserID, targetUserID, role).
			Return(permission, nil)

		resp, err := server.AddCollaborator(ctx, &sharePB.AddCollaboratorRequest{
			NoteId:        noteID,
			CurrentUserId: currentUserID,
			UserId:        targetUserID,
			Role:          sharePB.NoteRole_NOTE_ROLE_EDITOR,
		})
		assert.NoError(t, err)
		assert.Equal(t, uint64(5), resp.PermissionId)
	})

	t.Run("AccessDenied", func(t *testing.T) {
		mockSharingUC.EXPECT().
			AddCollaborator(ctx, noteID, currentUserID, targetUserID, role).
			Return(nil, constants.ErrNoAccess)

		_, err := server.AddCollaborator(ctx, &sharePB.AddCollaboratorRequest{
			NoteId:        noteID,
			CurrentUserId: currentUserID,
			UserId:        targetUserID,
			Role:          sharePB.NoteRole_NOTE_ROLE_EDITOR,
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_GetNoteByShareUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(mockNoteUC, mockBlockUC, mockSharingUC)
	ctx := context.Background()

	shareUUID := "uuid"
	note := &models.Note{ID: 10, Title: "Public Note", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().GetNoteByShareUUID(ctx, shareUUID).Return(note, nil)
		resp, err := server.GetNoteByShareUUID(ctx, &notePB.GetNoteByShareUUIDRequest{ShareUuid: shareUUID})
		assert.NoError(t, err)
		assert.Equal(t, note.ID, resp.Id)
	})
}

func TestServer_AddFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	server := NewServer(mockNoteUC, nil, nil)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().AddFavorite(ctx, userID, noteID).Return(nil)
		_, err := server.AddFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockNoteUC.EXPECT().AddFavorite(ctx, userID, noteID).Return(constants.ErrNotFound)
		_, err := server.AddFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_RemoveFavorite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNoteUC := mock.NewMockNoteUsecase(ctrl)
	server := NewServer(mockNoteUC, nil, nil)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mockNoteUC.EXPECT().RemoveFavorite(ctx, userID, noteID).Return(nil)
		_, err := server.RemoveFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockNoteUC.EXPECT().RemoveFavorite(ctx, userID, noteID).Return(constants.ErrNotFound)
		_, err := server.RemoveFavorite(ctx, &notePB.FavoriteRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	blockID := uint64(100)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: blockID, Type: "text"}}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().GetBlock(ctx, userID, blockID).Return(block, nil)
		resp, err := server.GetBlock(ctx, &blockPB.GetBlockRequest{UserId: userID, BlockId: blockID})
		assert.NoError(t, err)
		assert.Equal(t, blockID, resp.Id)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockBlockUC.EXPECT().GetBlock(ctx, userID, blockID).Return(nil, constants.ErrNotFound)
		_, err := server.GetBlock(ctx, &blockPB.GetBlockRequest{UserId: userID, BlockId: blockID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_CreateCodeBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: 100, Type: "code"}}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().CreateCodeBlock(ctx, userID, noteID, nil).Return(block, nil)
		resp, err := server.CreateCodeBlock(ctx, &blockPB.CreateCodeBlockRequest{UserId: userID, NoteId: noteID})
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), resp.Id)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().CreateCodeBlock(ctx, userID, noteID, nil).Return(nil, constants.ErrNoAccess)
		_, err := server.CreateCodeBlock(ctx, &blockPB.CreateCodeBlockRequest{UserId: userID, NoteId: noteID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_CreateAttachmentBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	fileID := uint64(999)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: 100, Type: "attachment"}}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().CreateAttachmentBlock(ctx, userID, noteID, nil, fileID).Return(block, nil)
		resp, err := server.CreateAttachmentBlock(ctx, &blockPB.CreateAttachmentBlockRequest{UserId: userID, NoteId: noteID, FileId: fileID})
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), resp.Id)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().CreateAttachmentBlock(ctx, userID, noteID, nil, fileID).Return(nil, constants.ErrNoAccess)
		_, err := server.CreateAttachmentBlock(ctx, &blockPB.CreateAttachmentBlockRequest{UserId: userID, NoteId: noteID, FileId: fileID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_UpdateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	blockID := uint64(100)
	text := "updated text"
	block := &models.Block{BaseBlock: models.BaseBlock{ID: blockID, Type: "text"}, Content: models.TextContent{Text: text}}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().UpdateBlock(ctx, userID, gomock.Any()).Return(block, nil)
		resp, err := server.UpdateBlock(ctx, &blockPB.UpdateBlockRequest{
			UserId:  userID,
			BlockId: blockID,
			Type:    "text",
			Content: &blockPB.UpdateBlockRequest_TextContent{TextContent: &blockPB.TextContent{Text: text}},
		})
		assert.NoError(t, err)
		assert.Equal(t, text, resp.GetTextContent().Text)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().UpdateBlock(ctx, userID, gomock.Any()).Return(nil, constants.ErrNoAccess)
		_, err := server.UpdateBlock(ctx, &blockPB.UpdateBlockRequest{
			UserId:  userID,
			BlockId: blockID,
			Type:    "text",
			Content: &blockPB.UpdateBlockRequest_TextContent{TextContent: &blockPB.TextContent{Text: text}},
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().DeleteBlock(ctx, userID, blockID).Return(nil)
		_, err := server.DeleteBlock(ctx, &blockPB.DeleteBlockRequest{UserId: userID, BlockId: blockID})
		assert.NoError(t, err)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().DeleteBlock(ctx, userID, blockID).Return(constants.ErrNoAccess)
		_, err := server.DeleteBlock(ctx, &blockPB.DeleteBlockRequest{UserId: userID, BlockId: blockID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_UpdateBlockPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlockUC := mock.NewMockBlocksUsecase(ctrl)
	server := NewServer(nil, mockBlockUC, nil)
	ctx := context.Background()

	userID := uint64(1)
	blockID := uint64(100)
	block := &models.Block{BaseBlock: models.BaseBlock{ID: blockID}}

	t.Run("Success", func(t *testing.T) {
		mockBlockUC.EXPECT().UpdateBlockPosition(ctx, userID, blockID, nil).Return(block, nil)
		resp, err := server.UpdateBlockPosition(ctx, &blockPB.UpdateBlockPositionRequest{UserId: userID, BlockId: blockID})
		assert.NoError(t, err)
		assert.Equal(t, blockID, resp.Id)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockBlockUC.EXPECT().UpdateBlockPosition(ctx, userID, blockID, nil).Return(nil, constants.ErrNoAccess)
		_, err := server.UpdateBlockPosition(ctx, &blockPB.UpdateBlockPositionRequest{UserId: userID, BlockId: blockID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_GetCollaborators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	collabs := []*models.NotePermission{{PermissionID: 1, Role: models.RoleViewer, CreatedAt: time.Now()}}
	ownerID := uint64(1)
	
	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().GetCollaborators(ctx, noteID, userID).Return(ownerID, collabs, nil, nil)
		resp, err := server.GetCollaborators(ctx, &sharePB.GetCollaboratorsRequest{NoteId: noteID, CurrentUserId: userID})
		assert.NoError(t, err)
		assert.Len(t, resp.Collaborators, 1)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingUC.EXPECT().GetCollaborators(ctx, noteID, userID).Return(uint64(0), nil, nil, constants.ErrNoAccess)
		_, err := server.GetCollaborators(ctx, &sharePB.GetCollaboratorsRequest{NoteId: noteID, CurrentUserId: userID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_UpdateCollaboratorRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	permID := uint64(5)
	perm := &models.NotePermission{PermissionID: permID, Role: models.RoleViewer}

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().UpdateCollaboratorRole(ctx, noteID, userID, permID, models.RoleViewer).Return(perm, nil)
		resp, err := server.UpdateCollaboratorRole(ctx, &sharePB.UpdateCollaboratorRoleRequest{
			NoteId:        noteID,
			CurrentUserId: userID,
			PermissionId:  permID,
			NewRole:       sharePB.NoteRole_NOTE_ROLE_VIEWER,
		})
		assert.NoError(t, err)
		assert.Equal(t, permID, resp.PermissionId)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingUC.EXPECT().UpdateCollaboratorRole(ctx, noteID, userID, permID, models.RoleViewer).Return(nil, constants.ErrNoAccess)
		_, err := server.UpdateCollaboratorRole(ctx, &sharePB.UpdateCollaboratorRoleRequest{
			NoteId:        noteID,
			CurrentUserId: userID,
			PermissionId:  permID,
			NewRole:       sharePB.NoteRole_NOTE_ROLE_VIEWER,
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_RemoveCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	permID := uint64(5)

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().RemoveCollaborator(ctx, noteID, userID, permID).Return(nil)
		_, err := server.RemoveCollaborator(ctx, &sharePB.RemoveCollaboratorRequest{NoteId: noteID, CurrentUserId: userID, PermissionId: permID})
		assert.NoError(t, err)
	})

	t.Run("NoAccess", func(t *testing.T) {
		mockSharingUC.EXPECT().RemoveCollaborator(ctx, noteID, userID, permID).Return(constants.ErrNoAccess)
		_, err := server.RemoveCollaborator(ctx, &sharePB.RemoveCollaboratorRequest{NoteId: noteID, CurrentUserId: userID, PermissionId: permID})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestServer_SetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	role := models.RoleViewer
	
	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().SetPublicAccess(ctx, noteID, userID, &role).Return(nil)
		mockSharingUC.EXPECT().GetPublicAccess(ctx, noteID, userID).Return(&role, nil)
		
		protoRole := sharePB.NoteRole_NOTE_ROLE_VIEWER
		resp, err := server.SetPublicAccess(ctx, &sharePB.SetPublicAccessRequest{
			NoteId:        noteID,
			CurrentUserId: userID,
			AccessLevel:   &protoRole,
		})
		assert.NoError(t, err)
		assert.Equal(t, sharePB.NoteRole_NOTE_ROLE_VIEWER, resp.GetAccessLevel())
		assert.Equal(t, "/notes/10", resp.ShareUrl)
	})
}

func TestServer_GetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	role := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().GetPublicAccess(ctx, noteID, userID).Return(&role, nil)
		resp, err := server.GetPublicAccess(ctx, &sharePB.GetPublicAccessRequest{NoteId: noteID, CurrentUserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, sharePB.NoteRole_NOTE_ROLE_VIEWER, resp.GetAccessLevel())
		assert.Equal(t, "/notes/10", resp.ShareUrl)
	})
}

func TestServer_GetSharingSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	settings := &models.SharingSettings{NoteID: noteID, IsOwner: true}

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().GetSharingSettings(ctx, noteID, userID).Return(settings, nil)
		resp, err := server.GetSharingSettings(ctx, &sharePB.GetSharingSettingsRequest{NoteId: noteID, CurrentUserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, noteID, resp.NoteId)
	})
}

func TestServer_ActivateAccessByLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	uuid := "uuid"
	res := &models.ActivateAccessResponse{NoteID: 10, AccessGranted: true}

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().ActivateAccessByLink(ctx, uuid, userID).Return(res, nil)
		resp, err := server.ActivateAccessByLink(ctx, &sharePB.ActivateAccessByLinkRequest{ShareUuid: uuid, UserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), resp.NoteId)
	})
}

func TestServer_CheckNoteAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSharingUC := mock.NewMockSharingUsecase(ctrl)
	server := NewServer(nil, nil, mockSharingUC)
	ctx := context.Background()

	userID := uint64(1)
	noteID := uint64(10)
	info := &models.NoteAccessInfo{HasAccess: true}

	t.Run("Success", func(t *testing.T) {
		mockSharingUC.EXPECT().CheckNoteAccess(ctx, noteID, userID).Return(info, nil)
		resp, err := server.CheckNoteAccess(ctx, &sharePB.CheckNoteAccessRequest{NoteId: noteID, UserId: userID})
		assert.NoError(t, err)
		assert.True(t, resp.HasAccess)
	})
}
