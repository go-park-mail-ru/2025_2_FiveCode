package delivery

import (
	"backend/apiutils"
	"backend/dto"
	"backend/logger"
	"backend/models"
	"context"
	"io"
	"net/http"
)

type TicketUsecase interface {
	GetStatistics(ctx context.Context) (dto.Statistics, error)
	CreateTicket(ctx context.Context, input CreateTicketInput) (*models.Ticket, error)
}

type TicketDelivery struct {
	Usecase TicketUsecase
}

func NewTicketDelivery(u TicketUsecase) *TicketDelivery {
	return &TicketDelivery{
		Usecase: u,
	}
}

func (d *TicketDelivery) GetStatistics(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	log.Info().Msg("GetStatistics called")
	stats, err := d.Usecase.GetStatistics(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get statistics")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, stats)
}

type CreateTicketInput struct {
	Email       string `json:"email" valid:"required,email"`
	FullName    string `json:"full_name" valid:"required"`
	Category    string `json:"category" valid:"required,in(bug|suggestion|complaint|other)"`
	Title       string `json:"title" valid:"required"`
	Description string `json:"description" valid:"required"`
	FileID      uint64 `json:"file_id"`
}

func (d *TicketDelivery) CreateTicket(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("failed to read request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var req CreateTicketInput
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid json for ticket creation")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ticket, err := d.Usecase.CreateTicket(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, ticket)
}
