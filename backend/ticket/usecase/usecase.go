package usecase

import (
	"backend/dto"
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
)

type TicketRepository interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error)
	GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error)
	GetStatistics(ctx context.Context) (dto.Statistics, error)
	CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error)
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
	return nil, nil
}

func (uc *TicketUsecase) GetStatistics(ctx context.Context) (dto.Statistics, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting statistics")

	stats, err := uc.Repository.GetStatistics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get statistics from repository")
		return dto.Statistics{}, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

func (u *TicketUsecase) CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	ticket.Status = models.TicketStatusOpen

	createdTicket, err := u.Repository.CreateTicket(ctx, ticket)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return createdTicket, nil
}
