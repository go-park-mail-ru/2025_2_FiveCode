package models

import "time"

const (
	BlockTypeText       = "text"
	BlockTypeCode       = "code"
	BlockTypeAttachment = "attachment"
)

type Block struct {
	ID        uint64      `json:"id"`
	NoteID    uint64      `json:"note_id"`
	Type      string      `json:"type"`
	Position  float64     `json:"position"`
	Content   interface{} `json:"content"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type TextContent struct {
	Text    string            `json:"text"`
	Formats []BlockTextFormat `json:"formats"`
}

type CodeContent struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type AttachmentContent struct {
	URL       string  `json:"url"`
	Caption   *string `json:"caption,omitempty"`
	MimeType  string  `json:"mime_type"`
	SizeBytes int     `json:"size_bytes"`
	Width     *int    `json:"width,omitempty"`
	Height    *int    `json:"height,omitempty"`
}

type TextFont string

const (
	FontInter      TextFont = "Inter"
	FontRoboto     TextFont = "Roboto"
	FontMontserrat TextFont = "Montserrat"
	FontManrope    TextFont = "Manrope"
)

type BlockTextFormat struct {
	ID            uint64   `json:"id"`
	StartOffset   int      `json:"start_offset"`
	EndOffset     int      `json:"end_offset"`
	Bold          bool     `json:"bold"`
	Italic        bool     `json:"italic"`
	Underline     bool     `json:"underline"`
	Strikethrough bool     `json:"strikethrough"`
	Link          *string  `json:"link,omitempty"`
	Font          TextFont `json:"font"`
	Size          int      `json:"size"`
}
