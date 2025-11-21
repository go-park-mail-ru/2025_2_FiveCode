package utils

import (
	blockPB "backend/gen/go/block"
	notePB "backend/gen/go/note"
	"backend/pkg/models"

	userPB "backend/gen/go/user"
)

func ProtoUserToModel(p *userPB.User) *models.User {
	if p == nil {
		return nil
	}
	m := &models.User{
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

func ProtoNoteToModel(protoNote *notePB.Note) *models.Note {
	if protoNote == nil {
		return nil
	}

	note := &models.Note{
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

func ProtoNotesToModels(protoNotes []*notePB.Note) []models.Note {
	if protoNotes == nil {
		return []models.Note{}
	}

	notes := make([]models.Note, len(protoNotes))
	for i, protoNote := range protoNotes {
		if protoNote != nil {
			notes[i] = *ProtoNoteToModel(protoNote)
		}
	}

	return notes
}

func ProtoBlockToModel(protoBlock *blockPB.Block) *models.Block {
	if protoBlock == nil {
		return nil
	}

	block := &models.Block{
		BaseBlock: models.BaseBlock{
			ID:        protoBlock.Id,
			NoteID:    protoBlock.NoteId,
			Type:      protoBlock.Type,
			Position:  protoBlock.Position,
			CreatedAt: protoBlock.CreatedAt.AsTime(),
			UpdatedAt: protoBlock.UpdatedAt.AsTime(),
		},
	}

	// Конвертируем content в зависимости от типа
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

// ProtoBlocksToModels converts slice of proto Blocks to slice of model Blocks
func ProtoBlocksToModels(protoBlocks []*blockPB.Block) []models.Block {
	if protoBlocks == nil {
		return []models.Block{}
	}

	blocks := make([]models.Block, len(protoBlocks))
	for i, protoBlock := range protoBlocks {
		if protoBlock != nil {
			blocks[i] = *ProtoBlockToModel(protoBlock)
		}
	}

	return blocks
}

// Helper functions для конвертации content

func protoTextContentToModel(protoContent *blockPB.TextContent) models.TextContent {
	if protoContent == nil {
		return models.TextContent{}
	}

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

	return models.TextContent{
		Text:    protoContent.Text,
		Formats: formats,
	}
}

func protoCodeContentToModel(protoContent *blockPB.CodeContent) models.CodeContent {
	if protoContent == nil {
		return models.CodeContent{}
	}

	return models.CodeContent{
		Code:     protoContent.Code,
		Language: protoContent.Language,
	}
}

func protoAttachmentContentToModel(protoContent *blockPB.AttachmentContent) models.AttachmentContent {
	if protoContent == nil {
		return models.AttachmentContent{}
	}

	content := models.AttachmentContent{
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

// Конвертация model -> proto (для UpdateBlock)

func ModelTextContentToProto(modelContent *models.UpdateTextContent) *blockPB.TextContent {
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

func ModelCodeContentToProto(modelContent *models.UpdateCodeContent) *blockPB.CodeContent {
	if modelContent == nil {
		return nil
	}

	return &blockPB.CodeContent{
		Code:     modelContent.Code,
		Language: modelContent.Language,
	}
}
