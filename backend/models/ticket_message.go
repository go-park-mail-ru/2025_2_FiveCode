package models

import "time"

const (
	TicketMessageSenderTypeUser  = "user"
	TicketMessageSenderTypeAdmin = "admin"
)

type TicketMessage struct {
	ID         uint64    `json:"id"`
	TicketID   uint64    `json:"ticket_id"`
	SenderID   uint64    `json:"sender_id"`
	SenderType string    `json:"sender_type"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}
