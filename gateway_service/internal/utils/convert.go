package utils

import (
	noteModels "backend/gateway_service/internal/notes/models"
	userModels "backend/gateway_service/internal/user/models"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	userPB "backend/user_service/pkg/user/v1"
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
