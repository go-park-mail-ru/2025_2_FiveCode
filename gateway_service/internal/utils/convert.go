package utils

import (
	noteModels "backend/gateway_service/internal/notes/models"
	shareModels "backend/gateway_service/internal/notes/models"
	userModels "backend/gateway_service/internal/user/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	sharePB "backend/notes_service/pkg/sharing/v1"
	userPB "backend/user_service/pkg/user/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note

func MapProtoToNote(p *notePB.Note) *noteModels.Note {
	if p == nil {
		return nil
	}
	m := &noteModels.Note{
		ID:         p.Id,
		OwnerID:    p.OwnerId,
		Title:      p.Title,
		IsFavorite: p.IsFavorite,
		IsArchived: p.IsArchived,
		IsShared:   p.IsShared,
		CreatedAt:  p.CreatedAt.AsTime(),
		UpdatedAt:  p.UpdatedAt.AsTime(),
	}
	if p.ParentNoteId != nil {
		m.ParentNoteID = p.ParentNoteId
	}
	if p.IconFileId != nil {
		m.IconFileID = p.IconFileId
	}
	if p.DeletedAt != nil {
		deletedAt := p.DeletedAt.AsTime()
		m.DeletedAt = &deletedAt
	}
	return m
}

// Block

func MapProtoToBlock(p *blockPB.Block) *noteModels.Block {
	if p == nil {
		return nil
	}
	b := &noteModels.Block{
		ID:        p.Id,
		NoteID:    p.NoteId,
		Type:      p.Type,
		Position:  p.Position,
		CreatedAt: p.CreatedAt.AsTime(),
		UpdatedAt: p.UpdatedAt.AsTime(),
	}

	switch c := p.Content.(type) {
	case *blockPB.Block_TextContent:
		b.Content = MapProtoTextContent(c.TextContent)
	case *blockPB.Block_CodeContent:
		b.Content = MapProtoCodeContent(c.CodeContent)
	case *blockPB.Block_AttachmentContent:
		b.Content = MapProtoAttachmentContent(c.AttachmentContent)
	}
	return b
}

func MapProtoTextContent(p *blockPB.TextContent) noteModels.TextContent {
	if p == nil {
		return noteModels.TextContent{}
	}
	formats := make([]noteModels.BlockTextFormat, len(p.Formats))
	for i, f := range p.Formats {
		formats[i] = noteModels.BlockTextFormat{
			ID:            f.Id,
			StartOffset:   int(f.StartOffset),
			EndOffset:     int(f.EndOffset),
			Bold:          f.Bold,
			Italic:        f.Italic,
			Underline:     f.Underline,
			Strikethrough: f.Strikethrough,
			Link:          f.Link,
			Font:          noteModels.TextFont(f.Font),
			Size:          int(f.Size),
		}
	}
	return noteModels.TextContent{Text: p.Text, Formats: formats}
}

func MapProtoCodeContent(p *blockPB.CodeContent) noteModels.CodeContent {
	if p == nil {
		return noteModels.CodeContent{}
	}
	return noteModels.CodeContent{Code: p.Code, Language: p.Language}
}

func MapProtoAttachmentContent(p *blockPB.AttachmentContent) noteModels.AttachmentContent {
	if p == nil {
		return noteModels.AttachmentContent{}
	}
	c := noteModels.AttachmentContent{
		URL:       p.Url,
		Caption:   p.Caption,
		MimeType:  p.MimeType,
		SizeBytes: int(p.SizeBytes),
	}
	if p.Width != nil {
		w := int(*p.Width)
		c.Width = &w
	}
	if p.Height != nil {
		h := int(*p.Height)
		c.Height = &h
	}
	return c
}

func MapModelTextContentToProto(m *noteModels.UpdateTextContent) *blockPB.TextContent {
	if m == nil {
		return nil
	}
	formats := make([]*blockPB.BlockTextFormat, len(m.Formats))
	for i, f := range m.Formats {
		formats[i] = &blockPB.BlockTextFormat{
			Id:            f.ID,
			StartOffset:   int32(f.StartOffset),
			EndOffset:     int32(f.EndOffset),
			Bold:          f.Bold,
			Italic:        f.Italic,
			Underline:     f.Underline,
			Strikethrough: f.Strikethrough,
			Link:          f.Link,
			Font:          string(f.Font),
			Size:          int32(f.Size),
		}
	}
	return &blockPB.TextContent{Text: m.Text, Formats: formats}
}

func MapModelCodeContentToProto(m *noteModels.UpdateCodeContent) *blockPB.CodeContent {
	if m == nil {
		return nil
	}
	return &blockPB.CodeContent{Code: m.Code, Language: m.Language}
}

// User

func MapProtoToUser(p *userPB.User) *userModels.User {
	if p == nil {
		return nil
	}
	u := &userModels.User{
		ID:        p.Id,
		Email:     p.Email,
		Username:  p.Username,
		CreatedAt: p.CreatedAt.AsTime(),
	}
	if p.AvatarFileId != nil {
		u.AvatarFileID = p.AvatarFileId
	}
	if p.UpdatedAt != nil {
		t := p.UpdatedAt.AsTime()
		u.UpdatedAt = &t
	}
	return u
}

func MapProtoRoleToModel(role sharePB.NoteRole) shareModels.NoteRole {
	switch role {
	case sharePB.NoteRole_NOTE_ROLE_OWNER:
		return shareModels.RoleOwner
	case sharePB.NoteRole_NOTE_ROLE_VIEWER:
		return shareModels.RoleViewer
	case sharePB.NoteRole_NOTE_ROLE_COMMENTER:
		return shareModels.RoleCommenter
	case sharePB.NoteRole_NOTE_ROLE_EDITOR:
		return shareModels.RoleEditor
	default:
		return shareModels.RoleViewer
	}
}

func MapModelRoleToProto(role shareModels.NoteRole) sharePB.NoteRole {
	switch role {
	case shareModels.RoleOwner:
		return sharePB.NoteRole_NOTE_ROLE_OWNER
	case shareModels.RoleViewer:
		return sharePB.NoteRole_NOTE_ROLE_VIEWER
	case shareModels.RoleCommenter:
		return sharePB.NoteRole_NOTE_ROLE_COMMENTER
	case shareModels.RoleEditor:
		return sharePB.NoteRole_NOTE_ROLE_EDITOR
	default:
		return sharePB.NoteRole_NOTE_ROLE_UNSPECIFIED
	}
}

func MapProtoToCollaborator(proto *sharePB.Collaborator) shareModels.Collaborator {
	return shareModels.Collaborator{
		PermissionID: proto.PermissionId,
		UserID:       proto.UserId,
		Role:         MapProtoRoleToModel(proto.Role),
		GrantedBy:    proto.GrantedBy,
		GrantedAt:    proto.GrantedAt.AsTime(),
	}
}

func MapProtoToCollaboratorResponse(proto *sharePB.CollaboratorResponse) *shareModels.CollaboratorResponse {
	return &shareModels.CollaboratorResponse{
		PermissionID: proto.PermissionId,
		Collaborator: MapProtoToCollaborator(proto.Collaborator),
	}
}

func MapProtoToGetCollaboratorsResponse(proto *sharePB.GetCollaboratorsResponse) *shareModels.GetCollaboratorsResponse {
	collaborators := make([]shareModels.Collaborator, len(proto.Collaborators))
	for i, c := range proto.Collaborators {
		collaborators[i] = MapProtoToCollaborator(c)
	}

	var publicAccessLevel *shareModels.NoteRole
	if proto.PublicAccessLevel != nil {
		role := MapProtoRoleToModel(*proto.PublicAccessLevel)
		publicAccessLevel = &role
	}

	return &shareModels.GetCollaboratorsResponse{
		NoteID:             proto.NoteId,
		OwnerID:            proto.OwnerId,
		Collaborators:      collaborators,
		PublicAccessLevel:  publicAccessLevel,
		TotalCollaborators: int(proto.TotalCollaborators),
	}
}

func MapProtoToPublicAccessResponse(proto *sharePB.PublicAccessResponse) *shareModels.PublicAccessResponse {
	var accessLevel *shareModels.NoteRole
	if proto.AccessLevel != nil {
		role := MapProtoRoleToModel(*proto.AccessLevel)
		accessLevel = &role
	}

	return &shareModels.PublicAccessResponse{
		NoteID:      proto.NoteId,
		AccessLevel: accessLevel,
		ShareURL:    proto.ShareUrl,
		UpdatedAt:   proto.UpdatedAt.AsTime(),
	}
}

func MapProtoToSharingSettingsResponse(proto *sharePB.SharingSettingsResponse) *shareModels.SharingSettingsResponse {
	collaborators := make([]shareModels.Collaborator, len(proto.Collaborators))
	for i, c := range proto.Collaborators {
		collaborators[i] = MapProtoToCollaborator(c)
	}

	var publicAccessLevel *shareModels.NoteRole
	if proto.PublicAccess.AccessLevel != nil {
		role := MapProtoRoleToModel(*proto.PublicAccess.AccessLevel)
		publicAccessLevel = &role
	}

	publicAccess := shareModels.PublicAccess{
		NoteID:      proto.PublicAccess.NoteId,
		AccessLevel: publicAccessLevel,
		ShareURL:    proto.PublicAccess.ShareUrl,
		UpdatedAt:   timestamppb.Now().AsTime(),
	}

	return &shareModels.SharingSettingsResponse{
		NoteID:             proto.NoteId,
		OwnerID:            proto.OwnerId,
		PublicAccess:       publicAccess,
		Collaborators:      collaborators,
		TotalCollaborators: int(proto.TotalCollaborators),
		IsOwner:            proto.IsOwner,
	}
}

func MapProtoToNoteAccessInfo(proto *sharePB.NoteAccessResponse) shareModels.NoteAccessInfo {
	return shareModels.NoteAccessInfo{
		HasAccess:  proto.HasAccess,
		Role:       MapProtoRoleToModel(proto.Role),
		IsOwner:    proto.IsOwner,
		CanEdit:    proto.CanEdit,
		CanComment: proto.CanComment,
	}
}

func MapProtoToActivateAccessResponse(proto *sharePB.ActivateAccessByLinkResponse) *shareModels.ActivateAccessResponse {
	return &shareModels.ActivateAccessResponse{
		NoteID:        proto.NoteId,
		AccessGranted: proto.AccessGranted,
		AccessInfo:    MapProtoToNoteAccessInfo(proto.AccessInfo),
	}
}
