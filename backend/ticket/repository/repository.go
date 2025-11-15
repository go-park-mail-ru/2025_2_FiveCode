package repository

import (
	"backend/logger"
	"backend/models"
	"context"
	"database/sql"
	"fmt"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{
		db: db,
	}
}

func (r *TicketRepository) CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	query := `
		INSERT INTO ticket (email, full_name, category, status, title, description, file_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, full_name, category, status, title, description, file_id, created_at, updated_at
	`
	created := &models.Ticket{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		ticket.Email,
		ticket.FullName,
		ticket.Category,
		ticket.Status,
		ticket.Title,
		ticket.Description,
		ticket.FileID,
	).Scan(
		&created.ID,
		&created.Email,
		&created.FullName,
		&created.Category,
		&created.Status,
		&created.Title,
		&created.Description,
		&created.FileID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return created, nil
}
