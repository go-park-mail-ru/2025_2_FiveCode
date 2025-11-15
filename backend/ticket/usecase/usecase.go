package usecase

import (
	"backend/dto"
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
	"strings"
)

type TicketRepository interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error)
	GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID uint64) (*models.Ticket, error)
	GetStatistics(ctx context.Context) (dto.Statistics, error)
	CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error)
	GetAllTickets(ctx context.Context) ([]models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID uint64, status string) (*models.Ticket, error)
	CreateTicketMessage(ctx context.Context, message *models.TicketMessage) (*models.TicketMessage, error)
	GetTicketMessages(ctx context.Context, ticketID uint64) ([]models.TicketMessage, error)
}

type TicketUsecase struct {
	Repository TicketRepository
}

func NewTicketUsecase(r TicketRepository) *TicketUsecase {
	return &TicketUsecase{
		Repository: r,
	}
}

func (uc *TicketUsecase) GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error) {
	log := logger.FromContext(ctx)
	tickets, err := uc.Repository.GetAllTicketsByUserId(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("error getting all tickets")
		return nil, fmt.Errorf("error getting all tickets: %w", err)
	}

	return tickets, nil
}

func (uc *TicketUsecase) GetAllTickets(ctx context.Context) ([]models.Ticket, error) {
	log := logger.FromContext(ctx)
	tickets, err := uc.Repository.GetAllTickets(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error getting all tickets")
		return nil, fmt.Errorf("error getting all tickets: %w", err)
	}

	return tickets, nil
}

func (uc *TicketUsecase) UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error) {
	log := logger.FromContext(ctx)
	tickets, err := uc.Repository.UpdateTicket(ctx, ticketID, userID, tittle, desc)
	if err != nil {
		log.Error().Err(err).Msg("error updating ticket")
		return nil, fmt.Errorf("error updating ticket: %w", err)
	}

	return tickets, nil
}

func (uc *TicketUsecase) GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error) {
	log := logger.FromContext(ctx)
	ticket, err := uc.Repository.GetTicketById(ctx, userID, ticketID)
	if err != nil {
		log.Error().Err(err).Msg("error getting ticket")
		return nil, fmt.Errorf("error getting ticket: %w", err)
	}

	return ticket, nil
}

func (uc *TicketUsecase) GetTicketByID(ctx context.Context, ticketID uint64) (*models.Ticket, error) {
	log := logger.FromContext(ctx)
	ticket, err := uc.Repository.GetTicketByID(ctx, ticketID)
	if err != nil {
		log.Error().Err(err).Msg("error getting ticket by id without user")
		return nil, fmt.Errorf("error getting ticket: %w", err)
	}

	return ticket, nil
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

func (uc *TicketUsecase) CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	ticket.Status = models.TicketStatusOpen

	createdTicket, err := uc.Repository.CreateTicket(ctx, ticket)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return createdTicket, nil
}

func (uc *TicketUsecase) UpdateTicketStatus(ctx context.Context, ticketID uint64, status string) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	ticket, err := uc.Repository.UpdateTicketStatus(ctx, ticketID, status)
	if err != nil {
		log.Error().Err(err).Msg("error updating ticket status in repository")
		return nil, fmt.Errorf("error updating ticket status: %w", err)
	}

	return ticket, nil
}

func (uc *TicketUsecase) CreateTicketMessage(ctx context.Context, senderID, ticketID uint64, body string, isAdmin bool) (*models.TicketMessage, error) {
	log := logger.FromContext(ctx)

	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("message body is empty")
	}

	senderType := models.TicketMessageSenderTypeUser
	if isAdmin {
		senderType = models.TicketMessageSenderTypeAdmin
	}

	message := &models.TicketMessage{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Body:       body,
	}

	createdMessage, err := uc.Repository.CreateTicketMessage(ctx, message)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket message")
		return nil, fmt.Errorf("failed to create ticket message: %w", err)
	}

	return createdMessage, nil
}

func (uc *TicketUsecase) GetTicketMessages(ctx context.Context, ticketID uint64) ([]models.TicketMessage, error) {
	log := logger.FromContext(ctx)

	messages, err := uc.Repository.GetTicketMessages(ctx, ticketID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get ticket messages")
		return nil, fmt.Errorf("failed to get ticket messages: %w", err)
	}

	return messages, nil
}
