package delivery

import (
	"backend/apiutils"
	"backend/dto"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type TicketUsecase interface {
	GetAllTicketsByUserId(ctx context.Context, userID uint64) ([]models.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint64, userID uint64, tittle, desc *string) (*models.Ticket, error)
	GetTicketById(ctx context.Context, userID uint64, ticketID uint64) (*models.Ticket, error)
	GetStatistics(ctx context.Context) (dto.Statistics, error)
	CreateTicket(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error)
	GetAllTickets(ctx context.Context) ([]models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID uint64, status string) (*models.Ticket, error)
}

type TicketDelivery struct {
	Usecase TicketUsecase
}

func NewTicketDelivery(u TicketUsecase) *TicketDelivery {
	return &TicketDelivery{
		Usecase: u,
	}
}

func (d *TicketDelivery) GetAllTicketsByUserId(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	tickets, err := d.Usecase.GetAllTicketsByUserId(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get all tickets")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get all tickets")
	}

	apiutils.WriteJSON(w, http.StatusOK, tickets)
}

func (s *TicketDelivery) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	tickets, err := s.Usecase.GetAllTickets(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get all tickets")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get all tickets")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, tickets)
}

type UpdateTicketRequest struct {
	Title      *string `json:"tittle"`
	Description *string `json:"description"`
}

func (d *TicketDelivery) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["ticket_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("invalid ticket id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedTicket, err := d.Usecase.UpdateTicket(r.Context(), ticketID, userID, req.Title, req.Description)
	if err != nil {
		log.Error().Err(err).Msg("failed to update ticket")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to update ticket")
	}
	apiutils.WriteJSON(w, http.StatusOK, updatedTicket)
}

func (d *TicketDelivery) GetTicketById(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	ticketID, _ := strconv.ParseUint(vars["ticket_id"], 10, 64)

	ticket, err := d.Usecase.GetTicketById(r.Context(), userID, ticketID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get ticket")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get ticket")
	}
	apiutils.WriteJSON(w, http.StatusOK, ticket)
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

	var req models.Ticket
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid json for ticket creation")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ticket, err := d.Usecase.CreateTicket(r.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("failed to create ticket")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, ticket)
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (s *TicketDelivery) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["ticket_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("invalid ticket_id format")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid ticket_id format")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body for status update")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status != models.TicketStatusOpen && req.Status != models.TicketStatusInProgress && req.Status != models.TicketStatusClosed {
		log.Warn().Str("status", req.Status).Msg("invalid ticket status provided")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid ticket status")
		return
	}

	updatedTicket, err := s.Usecase.UpdateTicketStatus(r.Context(), ticketID, req.Status)
	if err != nil {
		log.Error().Err(err).Msg("failed to update ticket status")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to update ticket status")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, updatedTicket)
}
