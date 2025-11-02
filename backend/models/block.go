package models

import "time"

type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeCode       BlockType = "code"
	BlockTypeAttachment BlockType = "attachment"
)

type Block struct {
	ID           uint64    `json:"id"`
	NoteID       uint64    `json:"note_id"`
	Type         BlockType `json:"type"`
	Position     float64   `json:"position"`
	LastEditedBy *uint64   `json:"last_edited_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BlockText struct {
	ID        uint64    `json:"id"`
	BlockID   uint64    `json:"block_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TextFont string

const (
	FontInter      TextFont = "Inter"
	FontRoboto     TextFont = "Roboto"
	FontMontserrat TextFont = "Montserrat"
	FontManrope    TextFont = "Manrope"
)

type BlockTextFormat struct {
	ID            uint64    `json:"id"`
	BlockTextID   uint64    `json:"block_text_id"`
	StartOffset   int       `json:"start_offset"`
	EndOffset     int       `json:"end_offset"`
	Bold          bool      `json:"bold"`
	Italic        bool      `json:"italic"`
	Underline     bool      `json:"underline"`
	Strikethrough bool      `json:"strikethrough"`
	Link          *string   `json:"link,omitempty"`
	Font          TextFont  `json:"font"`
	Size          int       `json:"size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BlockWithContent struct {
	Block
	Text    string            `json:"text,omitempty"`
	Formats []BlockTextFormat `json:"formats,omitempty"`
}
