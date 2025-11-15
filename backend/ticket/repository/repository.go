package repository

import (
	"backend/logger"
	"backend/models"
	"context"
	"database/sql"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{
		db: db,
	}
}

func (r *TicketRepository) GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT t.id, t.email, t.full_name, t.category, t.status, t.title, t.description, t.file_id, t.created_at, t.updated_at
		FROM ticket t
		INNER JOIN "user" u ON t.email = u.email
		WHERE u.id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query tickets by user_id")
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	var tickets []models.Ticket

	for rows.Next() {
		var ticket models.Ticket

		err := rows.Scan(
			&ticket.ID,
			&ticket.Email,
			&ticket.FullName,
			&ticket.Category,
			&ticket.Status,
			&ticket.Title,
			&ticket.Description,
			&ticket.FileID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan ticket")
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("failed to scan tickets")
		return nil, err
	}

	if len(tickets) == 0 {
		return []models.Ticket{}, nil
	}

	return tickets, nil
}

func (r *TicketRepository) UpdateTicket(ctx context.Context, userID uint64, ticket *models.Ticket) ([]models.Ticket, error) {
	return nil, nil
}

func (r *TicketRepository) GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error) {
	return nil, nil
}
