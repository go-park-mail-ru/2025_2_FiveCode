package usecase

import (
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
)

type TicketRepository interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, userID uint64, ticket *models.Ticket) ([]models.Ticket, error)
	GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error)
}

type TicketUsecase struct {
	Repository TicketRepository
}

func NewTicketUsecase(r TicketRepository) *TicketUsecase {
	return &TicketUsecase{
		Repository: r,
	}
}

func (u *TicketUsecase) GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error) {
	log := logger.FromContext(ctx)
	tickets, err := u.Repository.GetAllTicketsByUserId(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("error getting all tickets")
		return nil, fmt.Errorf("error getting all tickets: %w", err)
	}

	return tickets, nil
}
