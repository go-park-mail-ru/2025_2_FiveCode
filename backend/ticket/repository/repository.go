package repository

import (
	"backend/logger"
	"backend/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	log := logger.FromContext(ctx)

	var userEmail string
	emailQuery := `SELECT email FROM "user" WHERE id = $1`
	err := r.db.QueryRowContext(ctx, emailQuery, userID).Scan(&userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("user not found")
			return nil, fmt.Errorf("user not found")
		}
		log.Error().Err(err).Msg("failed to get user email")
		return nil, err
	}

	query := `
		SELECT id, email, full_name, category, status, title, description, file_id, created_at, updated_at
		FROM ticket
		WHERE id = $1 AND email = $2
	`

	var ticket models.Ticket
	var fileID sql.NullInt64

	err = r.db.QueryRowContext(ctx, query, ticketID, userEmail).Scan(
		&ticket.ID,
		&ticket.Email,
		&ticket.FullName,
		&ticket.Category,
		&ticket.Status,
		&ticket.Title,
		&ticket.Description,
		&fileID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().Err(err).Msg("ticket not found or access denied")
			return nil, fmt.Errorf("ticket not found or access denied")
		}
		log.Error().Err(err).Msg("failed to get ticket")
		return nil, err
	}

	if fileID.Valid {
		fid := uint64(fileID.Int64)
		ticket.FileID = &fid
	}

	return &ticket, nil
}
