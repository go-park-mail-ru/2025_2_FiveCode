package usecase

import (
	"backend/logger"
	"backend/models"
	"context"
	"fmt"
)

type CreateTicketInput struct {
	Email       string `json:"email" valid:"required,email"`
	FullName    string `json:"full_name" valid:"required"`
	Category    string `json:"category" valid:"required,in(bug|suggestion|complaint|other)"`
	Title       string `json:"title" valid:"required"`
	Description string `json:"description" valid:"required"`
	FileID      uint64 `json:"file_id"`
}

type TicketRepository interface {
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

func (u *TicketUsecase) CreateTicket(ctx context.Context, input CreateTicketInput) (*models.Ticket, error) {
	log := logger.FromContext(ctx)

	ticket := &models.Ticket{
		Email:       input.Email,
		FullName:    input.FullName,
		Category:    input.Category,
		Status:      models.TicketStatusOpen,
		Title:       input.Title,
		Description: input.Description,
		FileID:      input.FileID,
	}

	createdTicket, err := u.Repository.CreateTicket(ctx, ticket)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return createdTicket, nil
}
