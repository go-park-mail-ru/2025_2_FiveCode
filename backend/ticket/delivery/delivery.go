package delivery

type TicketUsecase interface {
}

type TicketDelivery struct {
	Usecase TicketUsecase
}

func NewTicketDelivery(u TicketUsecase) *TicketDelivery {
	return &TicketDelivery{
		Usecase: u,
	}
}

func (d *TicketDelivery) GetStatistics() {
	// Implementation will go here
}
