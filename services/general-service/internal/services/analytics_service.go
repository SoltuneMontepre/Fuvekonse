package services

import (
	"context"
	"general-service/internal/dto/analytics"
	"general-service/internal/dto/user/responses"
	"general-service/internal/repositories"
	"sync"
)

// AnalyticsService provides consolidated dashboard/analytics data.
type AnalyticsService struct {
	repos  *repositories.Repositories
	ticket *TicketService
}

// NewAnalyticsService creates an analytics service (depends on ticket service and repos).
func NewAnalyticsService(repos *repositories.Repositories, ticket *TicketService) *AnalyticsService {
	return &AnalyticsService{repos: repos, ticket: ticket}
}

// GetDashboard returns all dashboard analytics in one call. Independent operations run in parallel so response time is roughly the slowest of them, not the sum.
func (s *AnalyticsService) GetDashboard(ctx context.Context, timelineDays, revenueDays int) (*analytics.DashboardResponse, error) {
	if timelineDays <= 0 {
		timelineDays = 90
	}
	if timelineDays > 365 {
		timelineDays = 365
	}
	if revenueDays <= 0 {
		revenueDays = 90
	}
	if revenueDays > 365 {
		revenueDays = 365
	}

	out := &analytics.DashboardResponse{}
	var (
		statsErr, timelineErr, revenueErr, userErr, dealerErr, countryErr error
		mu                                                                 sync.Mutex
	)
	var wg sync.WaitGroup

	// Ticket statistics
	wg.Add(1)
	go func() {
		defer wg.Done()
		stats, err := s.ticket.GetTicketStatistics(ctx)
		mu.Lock()
		if err != nil {
			statsErr = err
		} else {
			out.TicketStats = stats
		}
		mu.Unlock()
	}()

	// Sales timeline
	wg.Add(1)
	go func() {
		defer wg.Done()
		items, err := s.ticket.GetTicketSalesTimeline(ctx, timelineDays)
		mu.Lock()
		if err != nil {
			timelineErr = err
		} else {
			out.SalesTimeline = items
		}
		mu.Unlock()
	}()

	// Revenue
	wg.Add(1)
	go func() {
		defer wg.Done()
		rev, err := s.ticket.GetTicketRevenue(ctx, revenueDays)
		mu.Lock()
		if err != nil {
			revenueErr = err
		} else {
			out.Revenue = rev
		}
		mu.Unlock()
	}()

	// User count
	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := s.repos.User.Count()
		mu.Lock()
		if err != nil {
			userErr = err
		} else {
			out.UserCount = n
		}
		mu.Unlock()
	}()

	// Dealer count
	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := s.repos.Dealer.CountBooths()
		mu.Lock()
		if err != nil {
			dealerErr = err
		} else {
			out.DealerCount = n
		}
		mu.Unlock()
	}()

	// Users by country
	wg.Add(1)
	go func() {
		defer wg.Done()
		byCountry, err := s.repos.User.CountByCountry()
		mu.Lock()
		if err != nil {
			countryErr = err
		} else {
			out.UsersByCountry = make([]responses.CountByCountryItem, len(byCountry))
			for i := range byCountry {
				out.UsersByCountry[i] = responses.CountByCountryItem{
					Country: byCountry[i].Country,
					Count:   int(byCountry[i].Count),
				}
			}
		}
		mu.Unlock()
	}()

	wg.Wait()

	// Return first error if any
	for _, e := range []error{statsErr, timelineErr, revenueErr, userErr, dealerErr, countryErr} {
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
