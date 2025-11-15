package repository

import (
	"backend/dto"
	"backend/logger"
	"backend/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{
		db: db,
	}
}

func (r *TicketRepository) GetStatistics(ctx context.Context) (dto.Statistics, error) {
	stats := dto.Statistics{}

	query := `
		SELECT
			category,
			COUNT(*) AS total_tickets,
			COUNT(*) FILTER (WHERE status = 'open') AS open_tickets,
			COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress_tickets,
			COUNT(*) FILTER (WHERE status = 'closed') AS closed_tickets
		FROM
			ticket
		GROUP BY
			category
		ORDER BY
			category`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute statistics query")
		return dto.Statistics{}, fmt.Errorf("failed to get statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat dto.StatisticForCategory

		err := rows.Scan(
			&stat.Category,
			&stat.TotalTickets,
			&stat.OpenTickets,
			&stat.InProgressTickets,
			&stat.ClosedTickets,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan statistics row")
			return dto.Statistics{}, fmt.Errorf("failed to scan statistics row: %w", err)
		}

		stats.Statistics = append(stats.Statistics, stat)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("error during statistics rows iteration")
		return dto.Statistics{}, fmt.Errorf("error during statistics rows iteration: %w", err)
	}

	return stats, nil
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
