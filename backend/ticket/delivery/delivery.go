package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"context"
	"net/http"
)

type TicketUsecase interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, userID uint64, ticket *models.Ticket) ([]models.Ticket, error)
	GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error)
}

type TicketDelivery struct {
	Usecase TicketUsecase
}

func NewTicketDelivery(u TicketUsecase) *TicketDelivery {
	return &TicketDelivery{
		Usecase: u,
	}
}

func (s *TicketDelivery) GetAllTicketsByUserId(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	tickets, err := s.Usecase.GetAllTicketsByUserId(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get all tickets")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get all tickets")
	}

	apiutils.WriteJSON(w, http.StatusOK, tickets)
}

func (s *TicketDelivery) UpdateTicket(w http.ResponseWriter, r *http.Request) {

}

func (s *TicketDelivery) GetTicketById(w http.ResponseWriter, r *http.Request) {

}
