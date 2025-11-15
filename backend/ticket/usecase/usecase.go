package usecase

import (
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
)

type TicketRepository interface {
	GetStatistics(ctx context.Context) (models.Statistics, error)
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

func (uc *TicketUsecase) GetStatistics(ctx context.Context) (models.Statistics, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting statistics")

	stats, err := uc.Repository.GetStatistics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get statistics from repository")
		return models.Statistics{}, fmt.Errorf("failed to get statistics: %w", err)
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
