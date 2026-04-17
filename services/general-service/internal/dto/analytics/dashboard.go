package analytics

import (
	ticketresponses "general-service/internal/dto/ticket/responses"
	userresponses "general-service/internal/dto/user/responses"
)

// DashboardResponse is the consolidated admin dashboard analytics (single-query response).
type DashboardResponse struct {
	TicketStats    *ticketresponses.TicketStatisticsResponse `json:"ticket_stats"`
	SalesTimeline  []ticketresponses.SalesByDayResponse      `json:"sales_timeline"`
	Revenue        *ticketresponses.RevenueResponse          `json:"revenue"`
	UserCount      int64                                     `json:"user_count"`
	DealerCount    int64                                     `json:"dealer_count"`
	UsersByCountry []userresponses.CountByCountryItem        `json:"users_by_country"`
}
