package utils

import (
	noteModels "backend/gateway_service/internal/notes/models"
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

func MapProtoRoleToModel(role sharePB.NoteRole) noteModels.NoteRole {
	switch role {
	case sharePB.NoteRole_NOTE_ROLE_OWNER:
		return noteModels.RoleOwner
	case sharePB.NoteRole_NOTE_ROLE_VIEWER:
		return noteModels.RoleViewer
	case sharePB.NoteRole_NOTE_ROLE_COMMENTER:
		return noteModels.RoleCommenter
	case sharePB.NoteRole_NOTE_ROLE_EDITOR:
		return noteModels.RoleEditor
	default:
		return noteModels.RoleViewer
	}
}

func MapModelRoleToProto(role noteModels.NoteRole) sharePB.NoteRole {
	switch role {
	case noteModels.RoleOwner:
		return sharePB.NoteRole_NOTE_ROLE_OWNER
	case noteModels.RoleViewer:
		return sharePB.NoteRole_NOTE_ROLE_VIEWER
	case noteModels.RoleCommenter:
		return sharePB.NoteRole_NOTE_ROLE_COMMENTER
	case noteModels.RoleEditor:
		return sharePB.NoteRole_NOTE_ROLE_EDITOR
	default:
		return sharePB.NoteRole_NOTE_ROLE_UNSPECIFIED
	}
}

func MapProtoToCollaborator(proto *sharePB.Collaborator) noteModels.Collaborator {
	return noteModels.Collaborator{
		PermissionID: proto.PermissionId,
		UserID:       proto.UserId,
		Role:         MapProtoRoleToModel(proto.Role),
		GrantedBy:    proto.GrantedBy,
		GrantedAt:    proto.GrantedAt.AsTime(),
	}
}

func MapProtoToCollaboratorResponse(proto *sharePB.CollaboratorResponse) *noteModels.CollaboratorResponse {
	return &noteModels.CollaboratorResponse{
		PermissionID: proto.PermissionId,
		Collaborator: MapProtoToCollaborator(proto.Collaborator),
	}
}

func MapProtoToGetCollaboratorsResponse(proto *sharePB.GetCollaboratorsResponse) *noteModels.GetCollaboratorsResponse {
	collaborators := make([]noteModels.Collaborator, len(proto.Collaborators))
	for i, c := range proto.Collaborators {
		collaborators[i] = MapProtoToCollaborator(c)
	}

	var publicAccessLevel *noteModels.NoteRole
	if proto.PublicAccessLevel != nil {
		role := MapProtoRoleToModel(*proto.PublicAccessLevel)
		publicAccessLevel = &role
	}

	return &noteModels.GetCollaboratorsResponse{
		NoteID:             proto.NoteId,
		OwnerID:            proto.OwnerId,
		Collaborators:      collaborators,
		PublicAccessLevel:  publicAccessLevel,
		TotalCollaborators: int(proto.TotalCollaborators),
	}
}

func MapProtoToPublicAccessResponse(proto *sharePB.PublicAccessResponse) *noteModels.PublicAccessResponse {
	var accessLevel *noteModels.NoteRole
	if proto.AccessLevel != nil {
		role := MapProtoRoleToModel(*proto.AccessLevel)
		accessLevel = &role
	}

	return &noteModels.PublicAccessResponse{
		NoteID:      proto.NoteId,
		AccessLevel: accessLevel,
		ShareURL:    proto.ShareUrl,
		UpdatedAt:   proto.UpdatedAt.AsTime(),
	}
}

func MapProtoToSharingSettingsResponse(proto *sharePB.SharingSettingsResponse) *noteModels.SharingSettingsResponse {
	collaborators := make([]noteModels.Collaborator, len(proto.Collaborators))
	for i, c := range proto.Collaborators {
		collaborators[i] = MapProtoToCollaborator(c)
	}

	var publicAccessLevel *noteModels.NoteRole
	if proto.PublicAccess.AccessLevel != nil {
		role := MapProtoRoleToModel(*proto.PublicAccess.AccessLevel)
		publicAccessLevel = &role
	}

	publicAccess := noteModels.PublicAccess{
		NoteID:      proto.PublicAccess.NoteId,
		AccessLevel: publicAccessLevel,
		ShareURL:    proto.PublicAccess.ShareUrl,
		UpdatedAt:   timestamppb.Now().AsTime(),
	}

	return &noteModels.SharingSettingsResponse{
		NoteID:             proto.NoteId,
		OwnerID:            proto.OwnerId,
		PublicAccess:       publicAccess,
		Collaborators:      collaborators,
		TotalCollaborators: int(proto.TotalCollaborators),
		IsOwner:            proto.IsOwner,
	}
}

func MapProtoToNoteAccessInfo(proto *sharePB.NoteAccessResponse) noteModels.NoteAccessInfo {
	return noteModels.NoteAccessInfo{
		HasAccess:  proto.HasAccess,
		Role:       MapProtoRoleToModel(proto.Role),
		IsOwner:    proto.IsOwner,
		CanEdit:    proto.CanEdit,
		CanComment: proto.CanComment,
	}
}

func MapProtoToActivateAccessResponse(proto *sharePB.ActivateAccessByLinkResponse) *noteModels.ActivateAccessResponse {
	return &noteModels.ActivateAccessResponse{
		NoteID:        proto.NoteId,
		AccessGranted: proto.AccessGranted,
		AccessInfo:    MapProtoToNoteAccessInfo(proto.AccessInfo),
	}
}

func MapProtoToSearchResult(p *notePB.SearchResult) noteModels.SearchResult {
	if p == nil {
		return noteModels.SearchResult{}
	}
	return noteModels.SearchResult{
		NoteID:           p.NoteId,
		Title:            p.Title,
		HighlightedTitle: p.HighlightedTitle,
		ContentSnippet:   p.ContentSnippet,
		Rank:             p.Rank,
		UpdatedAt:        p.UpdatedAt.AsTime(),
	}
}

func MapProtoToSearchNotesResponse(p *notePB.SearchNotesResponse) *noteModels.SearchNotesResponse {
	if p == nil {
		return &noteModels.SearchNotesResponse{
			Results: []noteModels.SearchResult{},
			Count:   0,
		}
	}

	results := make([]noteModels.SearchResult, len(p.Results))
	for i, result := range p.Results {
		results[i] = MapProtoToSearchResult(result)
	}

	return &noteModels.SearchNotesResponse{
		Results: results,
		Count:   int(p.Count),
	}
}
