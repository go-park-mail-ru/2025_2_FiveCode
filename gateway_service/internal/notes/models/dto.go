package models

type UpdateNoteInput struct {
	ID         uint64
	UserID     uint64
	Title      *string
	IsArchived *bool
}

type UpdateBlockInput struct {
	BlockID uint64
	Type    string
	Content interface{}
}

type UpdateTextContent struct {
	Text    string            `json:"text"`
	Formats []BlockTextFormat `json:"formats"`
}

type UpdateCodeContent struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type CreateTextBlockInput struct {
	UserID        uint64
	NoteID        uint64
	BeforeBlockID *uint64
}

type CreateCodeBlockInput struct {
	UserID        uint64
	NoteID        uint64
	BeforeBlockID *uint64
}

type CreateAttachmentBlockInput struct {
	UserID        uint64
	NoteID        uint64
	BeforeBlockID *uint64
	FileID        uint64
}
