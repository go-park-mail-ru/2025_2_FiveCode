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

func (r *TicketRepository) UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	var userEmail string
	emailQuery := `SELECT email FROM "user" WHERE id = $1`
	err := r.db.QueryRowContext(ctx, emailQuery, userID).Scan(&userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("failed to update ticket, user not found")
			return nil, fmt.Errorf("user not found")
		}
		log.Error().Err(err).Msg("failed to get user email")
		return nil, err
	}

	updates := []string{}
	args := []interface{}{}
	argPosition := 1

	if tittle != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argPosition))
		args = append(args, *tittle)
		argPosition++
	}

	if desc != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argPosition))
		args = append(args, *desc)
		argPosition++
	}

	if len(updates) == 0 {
		log.Warn().Msg("no fields to update")
		return nil, fmt.Errorf("no fields to update")
	}

	args = append(args, ticketID, userEmail)

	query := fmt.Sprintf(`
		UPDATE ticket
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $%d AND email = $%d
		RETURNING id, email, full_name, category, status, title, description, file_id, created_at, updated_at
	`, strings.Join(updates, ", "), argPosition, argPosition+1)

	var ticket models.Ticket
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().Err(err).Msg("fticket not found or access denied")
			return nil, fmt.Errorf("ticket not found or access denied")
		}
		log.Error().Err(err).Msg("failed to update ticket")
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepository) GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error) {
	return nil, nil
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
