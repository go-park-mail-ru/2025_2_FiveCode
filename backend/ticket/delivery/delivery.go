package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"context"
	"net/http"
)

type TicketUsecase interface {
	GetStatistics(ctx context.Context) (StatisticsResponse, error)
}

type TicketDelivery struct {
	Usecase TicketUsecase
}

func NewTicketDelivery(u TicketUsecase) *TicketDelivery {
	return &TicketDelivery{
		Usecase: u,
	}
}

type StatisticsForCategory struct {
	Category 	 string `json:"category"`
	TotalTickets     int `json:"total_tickets"`
	OpenTickets      int `json:"open_tickets"`
	ClosedTickets    int `json:"closed_tickets"`
	InProgressTickets int `json:"in_progress_tickets"`
}

type StatisticsResponse struct {
	Statistics []StatisticsForCategory `json:"statistics"`
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
