package models

import "time"

// File представляет загруженный файл
type File struct {
	ID        uint64     `json:"id"`
	URL       string     `json:"url"`
	MimeType  string     `json:"mime_type"`
	SizeBytes int64      `json:"size_bytes"`
	Width     *int       `json:"width,omitempty"`
	Height    *int       `json:"height,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
