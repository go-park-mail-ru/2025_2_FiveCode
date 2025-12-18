package models

import "time"

type Note struct {
	ID           uint64     `json:"id"`
	OwnerID      uint64     `json:"owner_id"`
	ParentNoteID *uint64    `json:"parent_note_id,omitempty"`
	Title        string     `json:"title"`
	Icon         *Icon      `json:"icon,omitempty"`
	Header       *Header    `json:"header,omitempty"`
	IsFavorite   bool       `json:"is_favorite"`
	IsArchived   bool       `json:"is_archived"`
	IsShared     bool       `json:"is_shared"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type SearchResult struct {
	NoteID           uint64    `json:"note_id"`
	Title            string    `json:"title"`
	HighlightedTitle string    `json:"highlighted_title"`
	ContentSnippet   string    `json:"content_snippet"`
	Rank             float32   `json:"rank"`
	UpdatedAt        time.Time `json:"updated_at"`
	Icon             *Icon     `json:"icon,omitempty"`
}

type SearchNotesResponse struct {
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
}

type Icon struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Header struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
