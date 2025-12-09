package utils

import (
	"testing"
	"time"

	noteModels "backend/gateway_service/internal/notes/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"
	userPB "backend/user_service/pkg/user/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapProtoToNote(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	t.Run("Nil input", func(t *testing.T) {
		assert.Nil(t, MapProtoToNote(nil))
	})

	t.Run("Full input", func(t *testing.T) {
		parentID := uint64(10)
		iconID := uint64(20)
		input := &notePB.Note{
			Id:           1,
			OwnerId:      2,
			Title:        "Title",
			IsFavorite:   true,
			IsArchived:   true,
			IsShared:     true,
			CreatedAt:    ts,
			UpdatedAt:    ts,
			DeletedAt:    ts,
			ParentNoteId: &parentID,
			IconFileId:   &iconID,
		}

		result := MapProtoToNote(input)
		assert.NotNil(t, result)
		assert.Equal(t, input.Id, result.ID)
		assert.Equal(t, input.OwnerId, result.OwnerID)
		assert.Equal(t, input.Title, result.Title)
		assert.True(t, result.IsFavorite)
		assert.True(t, result.IsArchived)
		assert.True(t, result.IsShared)
		assert.Equal(t, &parentID, result.ParentNoteID)
		assert.Equal(t, now.Unix(), result.CreatedAt.Unix())
		assert.Equal(t, now.Unix(), result.UpdatedAt.Unix())
		assert.Equal(t, now.Unix(), result.DeletedAt.Unix())
	})
}

func TestMapProtoToBlock(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	t.Run("Nil input", func(t *testing.T) {
		assert.Nil(t, MapProtoToBlock(nil))
	})

	t.Run("Text Block", func(t *testing.T) {
		input := &blockPB.Block{
			Id:        1,
			NoteId:    2,
			Type:      "text",
			Position:  1.5,
			CreatedAt: ts,
			UpdatedAt: ts,
			Content: &blockPB.Block_TextContent{
				TextContent: &blockPB.TextContent{
					Text: "hello",
				},
			},
		}
		result := MapProtoToBlock(input)
		assert.NotNil(t, result)
		assert.Equal(t, input.Id, result.ID)
		assert.IsType(t, noteModels.TextContent{}, result.Content)
		assert.Equal(t, "hello", result.Content.(noteModels.TextContent).Text)
	})

	t.Run("Code Block", func(t *testing.T) {
		input := &blockPB.Block{
			Id:   1,
			Type: "code",
			Content: &blockPB.Block_CodeContent{
				CodeContent: &blockPB.CodeContent{
					Code: "fmt.Println()",
				},
			},
			CreatedAt: ts,
			UpdatedAt: ts,
		}
		result := MapProtoToBlock(input)
		assert.NotNil(t, result)
		assert.IsType(t, noteModels.CodeContent{}, result.Content)
		assert.Equal(t, "fmt.Println()", result.Content.(noteModels.CodeContent).Code)
	})

	t.Run("Attachment Block", func(t *testing.T) {
		input := &blockPB.Block{
			Id:   1,
			Type: "image",
			Content: &blockPB.Block_AttachmentContent{
				AttachmentContent: &blockPB.AttachmentContent{
					Url: "http://example.com",
				},
			},
			CreatedAt: ts,
			UpdatedAt: ts,
		}
		result := MapProtoToBlock(input)
		assert.NotNil(t, result)
		assert.IsType(t, noteModels.AttachmentContent{}, result.Content)
		assert.Equal(t, "http://example.com", result.Content.(noteModels.AttachmentContent).URL)
	})
}

func TestMapProtoTextContent(t *testing.T) {
	t.Run("Nil input", func(t *testing.T) {
		result := MapProtoTextContent(nil)
		assert.Empty(t, result.Text)
	})

	t.Run("Full input", func(t *testing.T) {
		input := &blockPB.TextContent{
			Text: "hello world",
			Formats: []*blockPB.BlockTextFormat{
				{
					Id:          1,
					StartOffset: 0,
					EndOffset:   5,
					Bold:        true,
					Font:        "Arial",
					Size:        12,
				},
			},
		}
		result := MapProtoTextContent(input)
		assert.Equal(t, "hello world", result.Text)
		assert.Len(t, result.Formats, 1)
		assert.Equal(t, uint64(1), result.Formats[0].ID)
		assert.Equal(t, 0, result.Formats[0].StartOffset)
		assert.Equal(t, 5, result.Formats[0].EndOffset)
		assert.True(t, result.Formats[0].Bold)
		assert.Equal(t, noteModels.TextFont("Arial"), result.Formats[0].Font)
		assert.Equal(t, 12, result.Formats[0].Size)
	})
}

func TestMapProtoCodeContent(t *testing.T) {
	t.Run("Nil input", func(t *testing.T) {
		result := MapProtoCodeContent(nil)
		assert.Empty(t, result.Code)
	})

	t.Run("Full input", func(t *testing.T) {
		input := &blockPB.CodeContent{
			Code:     "print('hello')",
			Language: "python",
		}
		result := MapProtoCodeContent(input)
		assert.Equal(t, "print('hello')", result.Code)
		assert.Equal(t, "python", result.Language)
	})
}

func TestMapProtoAttachmentContent(t *testing.T) {
	t.Run("Nil input", func(t *testing.T) {
		result := MapProtoAttachmentContent(nil)
		assert.Empty(t, result.URL)
	})

	t.Run("Full input", func(t *testing.T) {
		width := int32(100)
		height := int32(200)
		caption := "Image"
		input := &blockPB.AttachmentContent{
			Url:       "http://example.com/img.jpg",
			Caption:   &caption,
			MimeType:  "image/jpeg",
			SizeBytes: 1024,
			Width:     &width,
			Height:    &height,
		}
		result := MapProtoAttachmentContent(input)
		assert.Equal(t, "http://example.com/img.jpg", result.URL)
		assert.Equal(t, "Image", *result.Caption)
		assert.Equal(t, "image/jpeg", result.MimeType)
		assert.Equal(t, 1024, result.SizeBytes)
		assert.Equal(t, 100, *result.Width)
		assert.Equal(t, 200, *result.Height)
	})
}

func TestMapModelTextContentToProto(t *testing.T) {
	t.Run("Nil input", func(t *testing.T) {
		assert.Nil(t, MapModelTextContentToProto(nil))
	})

	t.Run("Full input", func(t *testing.T) {
		input := &noteModels.UpdateTextContent{
			Text: "text",
			Formats: []noteModels.BlockTextFormat{
				{
					ID:          1,
					StartOffset: 0,
					EndOffset:   5,
					Bold:        true,
					Font:        "Arial",
					Size:        12,
				},
			},
		}
		result := MapModelTextContentToProto(input)
		assert.Equal(t, "text", result.Text)
		assert.Len(t, result.Formats, 1)
		assert.Equal(t, uint64(1), result.Formats[0].Id)
		assert.True(t, result.Formats[0].Bold)
	})
}

func TestMapModelCodeContentToProto(t *testing.T) {
	t.Run("Nil input", func(t *testing.T) {
		assert.Nil(t, MapModelCodeContentToProto(nil))
	})

	t.Run("Full input", func(t *testing.T) {
		input := &noteModels.UpdateCodeContent{
			Code:     "code",
			Language: "go",
		}
		result := MapModelCodeContentToProto(input)
		assert.Equal(t, "code", result.Code)
		assert.Equal(t, "go", result.Language)
	})
}

func TestMapProtoToUser(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	t.Run("Nil input", func(t *testing.T) {
		assert.Nil(t, MapProtoToUser(nil))
	})

	t.Run("Full input", func(t *testing.T) {
		avatarID := uint64(10)
		input := &userPB.User{
			Id:           1,
			Email:        "test@example.com",
			Username:     "user",
			CreatedAt:    ts,
			UpdatedAt:    ts,
			AvatarFileId: &avatarID,
		}
		result := MapProtoToUser(input)

		assert.Equal(t, input.Id, result.ID)
		assert.Equal(t, input.Email, result.Email)
		assert.Equal(t, &avatarID, result.AvatarFileID)
		assert.Equal(t, now.Unix(), result.CreatedAt.Unix())
		assert.Equal(t, now.Unix(), result.UpdatedAt.Unix())
	})
}

func TestMapProtoRoleToModel(t *testing.T) {
	tests := []struct {
		input    sharePB.NoteRole
		expected noteModels.NoteRole
	}{
		{sharePB.NoteRole_NOTE_ROLE_VIEWER, noteModels.RoleViewer},
		{sharePB.NoteRole_NOTE_ROLE_COMMENTER, noteModels.RoleCommenter},
		{sharePB.NoteRole_NOTE_ROLE_EDITOR, noteModels.RoleEditor},
		{sharePB.NoteRole_NOTE_ROLE_UNSPECIFIED, noteModels.RoleViewer},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, MapProtoRoleToModel(tt.input))
	}
}

func TestMapModelRoleToProto(t *testing.T) {
	tests := []struct {
		input    noteModels.NoteRole
		expected sharePB.NoteRole
	}{
		{noteModels.RoleViewer, sharePB.NoteRole_NOTE_ROLE_VIEWER},
		{noteModels.RoleCommenter, sharePB.NoteRole_NOTE_ROLE_COMMENTER},
		{noteModels.RoleEditor, sharePB.NoteRole_NOTE_ROLE_EDITOR},
		{"unknown", sharePB.NoteRole_NOTE_ROLE_UNSPECIFIED},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, MapModelRoleToProto(tt.input))
	}
}

func TestMapProtoToCollaborator(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)
	input := &sharePB.Collaborator{
		PermissionId: 1,
		UserId:       2,
		Role:         sharePB.NoteRole_NOTE_ROLE_EDITOR,
		GrantedBy:    3,
		GrantedAt:    ts,
	}
	result := MapProtoToCollaborator(input)
	assert.Equal(t, input.PermissionId, result.PermissionID)
	assert.Equal(t, input.UserId, result.UserID)
	assert.Equal(t, noteModels.RoleEditor, result.Role)
	assert.Equal(t, input.GrantedBy, result.GrantedBy)
	assert.Equal(t, now.Unix(), result.GrantedAt.Unix())
}

func TestMapProtoToCollaboratorResponse(t *testing.T) {
	input := &sharePB.CollaboratorResponse{
		PermissionId: 1,
		Collaborator: &sharePB.Collaborator{
			PermissionId: 1,
			UserId:       2,
			Role:         sharePB.NoteRole_NOTE_ROLE_EDITOR,
			GrantedAt:    timestamppb.Now(),
		},
	}
	result := MapProtoToCollaboratorResponse(input)
	assert.Equal(t, input.PermissionId, result.PermissionID)
	assert.Equal(t, input.Collaborator.UserId, result.Collaborator.UserID)
}

func TestMapProtoToGetCollaboratorsResponse(t *testing.T) {
	role := sharePB.NoteRole_NOTE_ROLE_VIEWER
	input := &sharePB.GetCollaboratorsResponse{
		NoteId:             1,
		OwnerId:            2,
		Collaborators:      []*sharePB.Collaborator{},
		PublicAccessLevel:  &role,
		TotalCollaborators: 0,
	}
	result := MapProtoToGetCollaboratorsResponse(input)
	assert.Equal(t, input.NoteId, result.NoteID)
	assert.Equal(t, noteModels.RoleViewer, *result.PublicAccessLevel)
}

func TestMapProtoToPublicAccessResponse(t *testing.T) {
	role := sharePB.NoteRole_NOTE_ROLE_VIEWER
	input := &sharePB.PublicAccessResponse{
		NoteId:      1,
		AccessLevel: &role,
		ShareUrl:    "http://url",
		UpdatedAt:   timestamppb.Now(),
	}
	result := MapProtoToPublicAccessResponse(input)
	assert.Equal(t, input.NoteId, result.NoteID)
	assert.Equal(t, noteModels.RoleViewer, *result.AccessLevel)
	assert.Equal(t, "http://url", result.ShareURL)
}

func TestMapProtoToSharingSettingsResponse(t *testing.T) {
	role := sharePB.NoteRole_NOTE_ROLE_VIEWER
	input := &sharePB.SharingSettingsResponse{
		NoteId:  1,
		OwnerId: 2,
		PublicAccess: &sharePB.PublicAccess{
			NoteId:      1,
			AccessLevel: &role,
			ShareUrl:    "http://url",
		},
		Collaborators:      []*sharePB.Collaborator{},
		TotalCollaborators: 0,
		IsOwner:            true,
	}
	result := MapProtoToSharingSettingsResponse(input)
	assert.Equal(t, input.NoteId, result.NoteID)
	assert.Equal(t, noteModels.RoleViewer, *result.PublicAccess.AccessLevel)
	assert.True(t, result.IsOwner)
}

func TestMapProtoToNoteAccessInfo(t *testing.T) {
	input := &sharePB.NoteAccessResponse{
		HasAccess:  true,
		Role:       sharePB.NoteRole_NOTE_ROLE_EDITOR,
		IsOwner:    false,
		CanEdit:    true,
		CanComment: true,
	}
	result := MapProtoToNoteAccessInfo(input)
	assert.True(t, result.HasAccess)
	assert.Equal(t, noteModels.RoleEditor, result.Role)
	assert.True(t, result.CanEdit)
}

func TestMapProtoToActivateAccessResponse(t *testing.T) {
	input := &sharePB.ActivateAccessByLinkResponse{
		NoteId:        1,
		AccessGranted: true,
		AccessInfo: &sharePB.NoteAccessResponse{
			HasAccess: true,
			Role:      sharePB.NoteRole_NOTE_ROLE_VIEWER,
		},
	}
	result := MapProtoToActivateAccessResponse(input)
	assert.Equal(t, input.NoteId, result.NoteID)
	assert.True(t, result.AccessGranted)
	assert.True(t, result.AccessInfo.HasAccess)
}
