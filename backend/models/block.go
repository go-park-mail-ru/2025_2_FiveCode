package models

import "time"

type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeCode       BlockType = "code"
	BlockTypeAttachment BlockType = "attachment"
)

type Block struct {
	ID           int64     `json:"id"`
	NoteID       int64     `json:"note_id"`
	Type         BlockType `json:"type"`
	Position     float64   `json:"position"`
	LastEditedBy *int64    `json:"last_edited_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BlockText - текстовое содержимое блока
type BlockText struct {
	ID        int64     `json:"id"`
	BlockID   int64     `json:"block_id"`
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

// BlockTextFormat - форматирование диапазона текста (range)
type BlockTextFormat struct {
	ID            uint64    `json:"id"`
	BlockTextID   int64     `json:"block_text_id"`
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
