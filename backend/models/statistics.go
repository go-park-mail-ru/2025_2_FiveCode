package models

type StatisticForCategory struct {
	Category          string `json:"category"`
	TotalTickets      int    `json:"total_tickets"`
	OpenTickets       int    `json:"open_tickets"`
	ClosedTickets     int    `json:"closed_tickets"`
	InProgressTickets int    `json:"in_progress_tickets"`
}

type Statistics struct {
	Statistics []StatisticForCategory `json:"statistics"`
}

var StatisticsCategory = []string{
	TicketCategoryBug,
	TicketCategorySuggestion,
	TicketCategoryComplaint,
	TicketCategoryOther,
}
