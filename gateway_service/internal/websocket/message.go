package websocket

import (
	"time"

	"backend/gateway_service/internal/notes/models"
)

type MessageType string

const (
	MessageTypeNoteUpdate MessageType = "note_update"
	MessageTypeError      MessageType = "error"
)

type ClientMessage struct {
	Type   MessageType `json:"type"`
	NoteID int         `json:"note_id,omitempty"`
}

type ServerMessage struct {
	Type      MessageType `json:"type"`
	NoteID    int         `json:"note_id,omitempty"`
	UpdatedBy int         `json:"updated_by,omitempty"`
	UpdatedAt time.Time   `json:"updated_at,omitempty"`
	Blocks    interface{} `json:"blocks,omitempty"`
	Message   string      `json:"message,omitempty"`
	Title     string      `json:"title,omitempty"`

	Icon   *models.Icon   `json:"icon,omitempty"`
	Header *models.Header `json:"header,omitempty"`
}
