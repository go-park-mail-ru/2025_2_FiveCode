package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/models"
	"context"
	"net/http"
)

type TicketUsecase interface {
	GetStatistics(ctx context.Context) (models.Statistics, error)
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
