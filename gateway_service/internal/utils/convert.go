package utils

import (
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	noteModel "backend/notes_service/models"
	userPB "backend/user_service/pkg/user/v1"
	userModel "backend/user_service/models"
)

func ProtoUserToModel(p *userPB.User) *userModel.User {
	if p == nil {
		return nil
	}
	m := &userModel.User{
		ID:        p.Id,
		Email:     p.Email,
		Username:  p.Username,
		CreatedAt: p.CreatedAt.AsTime(),
	}
	if p.UpdatedAt != nil && p.UpdatedAt.IsValid() {
		updatedTime := p.UpdatedAt.AsTime()
		m.UpdatedAt = &updatedTime
	}
	if p.AvatarFileId != nil {
		m.AvatarFileID = p.AvatarFileId
	}
	return m
}

func ProtoNoteToModel(protoNote *notePB.Note) *noteModel.Note {
	if protoNote == nil {
		return nil
	}

	note := &noteModel.Note{
		ID:         protoNote.Id,
		OwnerID:    protoNote.OwnerId,
		Title:      protoNote.Title,
		IsFavorite: protoNote.IsFavorite,
		IsArchived: protoNote.IsArchived,
		IsShared:   protoNote.IsShared,
		CreatedAt:  protoNote.CreatedAt.AsTime(),
		UpdatedAt:  protoNote.UpdatedAt.AsTime(),
	}

	if protoNote.ParentNoteId != nil {
		note.ParentNoteID = protoNote.ParentNoteId
	}

	if protoNote.IconFileId != nil {
		note.IconFileID = protoNote.IconFileId
	}

	if protoNote.DeletedAt != nil {
		deletedAt := protoNote.DeletedAt.AsTime()
		note.DeletedAt = &deletedAt
	}

	return note
}

func ProtoNotesToModels(protoNotes []*notePB.Note) []noteModel.Note {
	if protoNotes == nil {
		return []noteModel.Note{}
	}

	notes := make([]noteModel.Note, len(protoNotes))
	for i, protoNote := range protoNotes {
		if protoNote != nil {
			notes[i] = *ProtoNoteToModel(protoNote)
		}
	}

	return notes
}

func ProtoBlockToModel(protoBlock *blockPB.Block) *noteModel.Block {
	if protoBlock == nil {
		return nil
	}

	block := &noteModel.Block{
		BaseBlock: noteModel.BaseBlock{
			ID:        protoBlock.Id,
			NoteID:    protoBlock.NoteId,
			Type:      protoBlock.Type,
			Position:  protoBlock.Position,
			CreatedAt: protoBlock.CreatedAt.AsTime(),
			UpdatedAt: protoBlock.UpdatedAt.AsTime(),
		},
	}

	switch content := protoBlock.Content.(type) {
	case *blockPB.Block_TextContent:
		block.Content = protoTextContentToModel(content.TextContent)
	case *blockPB.Block_CodeContent:
		block.Content = protoCodeContentToModel(content.CodeContent)
	case *blockPB.Block_AttachmentContent:
		block.Content = protoAttachmentContentToModel(content.AttachmentContent)
	}

	return block
}

func ProtoBlocksToModels(protoBlocks []*blockPB.Block) []noteModel.Block {
	if protoBlocks == nil {
		return []noteModel.Block{}
	}

	blocks := make([]noteModel.Block, len(protoBlocks))
	for i, protoBlock := range protoBlocks {
		if protoBlock != nil {
			blocks[i] = *ProtoBlockToModel(protoBlock)
		}
	}

	return blocks
}

func protoTextContentToModel(protoContent *blockPB.TextContent) noteModel.TextContent {
	if protoContent == nil {
		return noteModel.TextContent{}
	}

	formats := make([]noteModel.BlockTextFormat, len(protoContent.Formats))
	for i, f := range protoContent.Formats {
		formats[i] = noteModel.BlockTextFormat{
			ID:            f.Id,
			StartOffset:   int(f.StartOffset),
			EndOffset:     int(f.EndOffset),
			Bold:          f.Bold,
			Italic:        f.Italic,
			Underline:     f.Underline,
			Strikethrough: f.Strikethrough,
			Font:          noteModel.TextFont(f.Font),
			Size:          int(f.Size),
		}
		if f.Link != nil {
			formats[i].Link = f.Link
		}
	}

	return noteModel.TextContent{
		Text:    protoContent.Text,
		Formats: formats,
	}
}

func protoCodeContentToModel(protoContent *blockPB.CodeContent) noteModel.CodeContent {
	if protoContent == nil {
		return noteModel.CodeContent{}
	}

	return noteModel.CodeContent{
		Code:     protoContent.Code,
		Language: protoContent.Language,
	}
}

func protoAttachmentContentToModel(protoContent *blockPB.AttachmentContent) noteModel.AttachmentContent {
	if protoContent == nil {
		return noteModel.AttachmentContent{}
	}

	content := noteModel.AttachmentContent{
		URL:       protoContent.Url,
		MimeType:  protoContent.MimeType,
		SizeBytes: int(protoContent.SizeBytes),
	}

	if protoContent.Caption != nil {
		content.Caption = protoContent.Caption
	}
	if protoContent.Width != nil {
		width := int(*protoContent.Width)
		content.Width = &width
	}
	if protoContent.Height != nil {
		height := int(*protoContent.Height)
		content.Height = &height
	}

	return content
}

func ModelTextContentToProto(modelContent *noteModel.UpdateTextContent) *blockPB.TextContent {
	if modelContent == nil {
		return nil
	}

	formats := make([]*blockPB.BlockTextFormat, len(modelContent.Formats))
	for i, f := range modelContent.Formats {
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
		Text:    modelContent.Text,
		Formats: formats,
	}
}

func ModelCodeContentToProto(modelContent *noteModel.UpdateCodeContent) *blockPB.CodeContent {
	if modelContent == nil {
		return nil
	}

	return &blockPB.CodeContent{
		Code:     modelContent.Code,
		Language: modelContent.Language,
	}
}
