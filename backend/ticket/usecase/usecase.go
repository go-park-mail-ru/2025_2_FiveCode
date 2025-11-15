package usecase

import (
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
)

type TicketRepository interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error)
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

func (u *TicketUsecase) UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error) {
	log := logger.FromContext(ctx)
	tickets, err := u.Repository.UpdateTicket(ctx, ticketID, userID, tittle, desc)
	if err != nil {
		log.Error().Err(err).Msg("error updating ticket")
		return nil, fmt.Errorf("error updating ticket: %w", err)
	}

	return tickets, nil
}

func (u *TicketUsecase) GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error) {
	log := logger.FromContext(ctx)
	ticket, err := u.Repository.GetTicketById(ctx, userID, ticketID)
	if err != nil {
		log.Error().Err(err).Msg("error getting ticket")
		return nil, fmt.Errorf("error getting ticket: %w", err)
	}

	return ticket, nil
}
