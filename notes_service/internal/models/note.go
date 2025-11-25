package models

import "time"

type Note struct {
	ID           uint64     `json:"id"`
	OwnerID      uint64     `json:"owner_id"`
	ParentNoteID *uint64    `json:"parent_note_id,omitempty"`
	Title        string     `json:"title"`
	IconFileID   *uint64    `json:"icon_file_id,omitempty"`
	IsFavorite   bool       `json:"is_favorite"`
	IsArchived   bool       `json:"is_archived"`
	IsShared     bool       `json:"is_shared"`
	ShareUUID    *string    `json:"share_uuid,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
