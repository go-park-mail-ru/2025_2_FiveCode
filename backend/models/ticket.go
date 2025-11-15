package models

import "time"

const (
	TicketCategoryBug        = "bug"
	TicketCategorySuggestion = "suggestion"
	TicketCategoryComplaint  = "complaint"
	TicketCategoryOther      = "other"

	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusClosed     = "closed"
)

type Ticket struct {
	ID          uint64    `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileID      *uint64   `json:"file_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
