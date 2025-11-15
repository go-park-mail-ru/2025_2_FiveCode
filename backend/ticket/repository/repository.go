package repository

import (
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

func (r *TicketRepository) GetStatistics(ctx context.Context) (models.Statistics, error) {
	stats := models.Statistics{}

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
		return models.Statistics{}, fmt.Errorf("failed to get statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat models.StatisticForCategory
		
		err := rows.Scan(
			&stat.Category,
			&stat.TotalTickets,
			&stat.OpenTickets,
			&stat.InProgressTickets,
			&stat.ClosedTickets,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan statistics row")
			return models.Statistics{}, fmt.Errorf("failed to scan statistics row: %w", err)
		}

		stats.Statistics = append(stats.Statistics, stat)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("error during statistics rows iteration")
		return models.Statistics{}, fmt.Errorf("error during statistics rows iteration: %w", err)
	}

	return stats, nil
}
