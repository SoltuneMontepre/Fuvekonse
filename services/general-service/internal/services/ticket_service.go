package services

import (
	"context"
	"general-service/internal/common/constants"
	"general-service/internal/dto/common"
	"general-service/internal/dto/ticket/requests"
	"general-service/internal/dto/ticket/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"math"

	"github.com/google/uuid"
)

// Re-export sentinel errors from constants for backward compatibility
var (
	ErrInvalidTierID       = constants.ErrInvalidTierID
	ErrInvalidTicketID     = constants.ErrInvalidTicketID
	ErrInvalidUserID       = constants.ErrInvalidUserID
	ErrNoTicketFound       = constants.ErrNoTicketFound
	ErrInvalidTicketStatus = constants.ErrInvalidTicketStatus
)

type TicketService struct {
	repos *repositories.Repositories
}

func NewTicketService(repos *repositories.Repositories) *TicketService {
	return &TicketService{repos: repos}
}

// ========== Public User Endpoints ==========

// GetAllTiers returns all active ticket tiers
func (s *TicketService) GetAllTiers(ctx context.Context) ([]responses.TicketTierResponse, error) {
	tiers, err := s.repos.Ticket.GetAllActiveTiers(ctx)
	if err != nil {
		return nil, err
	}
	return mappers.MapTicketTiersToResponse(tiers), nil
}

// GetTierByID returns a specific ticket tier
func (s *TicketService) GetTierByID(ctx context.Context, tierID string) (*responses.TicketTierResponse, error) {
	id, err := uuid.Parse(tierID)
	if err != nil {
		return nil, ErrInvalidTierID
	}

	tier, err := s.repos.Ticket.GetTierByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mappers.MapTicketTierToResponse(tier), nil
}

// GetMyTicket returns the current user's ticket (if any)
func (s *TicketService) GetMyTicket(ctx context.Context, userID string) (*responses.UserTicketResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	ticket, err := s.repos.Ticket.GetUserTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, nil // No ticket found - valid response
	}
	return mappers.MapUserTicketToResponse(ticket, false), nil
}

// PurchaseTicket creates a new pending ticket for the user
func (s *TicketService) PurchaseTicket(ctx context.Context, userID string, req *requests.PurchaseTicketRequest) (*responses.UserTicketResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	tierID, err := uuid.Parse(req.TierID)
	if err != nil {
		return nil, ErrInvalidTierID
	}

	ticket, err := s.repos.Ticket.PurchaseTicket(ctx, uid, tierID)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, false), nil
}

// ConfirmPayment updates the user's ticket to self_confirmed status
func (s *TicketService) ConfirmPayment(ctx context.Context, userID string) (*responses.UserTicketResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	// Get user's current ticket
	existingTicket, err := s.repos.Ticket.GetUserTicket(ctx, uid)
	if err != nil {
		return nil, err
	}
	if existingTicket == nil {
		return nil, ErrNoTicketFound
	}

	ticket, err := s.repos.Ticket.ConfirmPayment(ctx, existingTicket.Id, uid)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, false), nil
}

// CancelTicket cancels a pending or self_confirmed ticket
func (s *TicketService) CancelTicket(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}

	// Get user's current ticket
	existingTicket, err := s.repos.Ticket.GetUserTicket(ctx, uid)
	if err != nil {
		return err
	}
	if existingTicket == nil {
		return ErrNoTicketFound
	}

	return s.repos.Ticket.CancelTicket(ctx, existingTicket.Id, uid)
}

// UpdateBadgeDetails updates badge details after ticket is approved
func (s *TicketService) UpdateBadgeDetails(ctx context.Context, userID string, req *requests.UpdateBadgeDetailsRequest) (*responses.UserTicketResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	// Get user's current ticket
	existingTicket, err := s.repos.Ticket.GetUserTicket(ctx, uid)
	if err != nil {
		return nil, err
	}
	if existingTicket == nil {
		return nil, ErrNoTicketFound
	}

	ticket, err := s.repos.Ticket.UpdateBadgeDetails(ctx, existingTicket.Id, uid, req.ConBadgeName, req.BadgeImage, req.IsFursuiter, req.IsFursuitStaff)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, false), nil
}

// ========== Admin Endpoints ==========

// GetTicketsForAdmin returns tickets with filters for admin view
func (s *TicketService) GetTicketsForAdmin(ctx context.Context, req *requests.AdminTicketFilterRequest) ([]responses.UserTicketResponse, *common.PaginationMeta, error) {
	// Validate and set defaults for pagination params
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := repositories.AdminTicketFilter{
		Search:        req.Search,
		PendingOver24: req.PendingOver24,
		Page:          page,
		PageSize:      pageSize,
	}

	// Parse status if provided
	if req.Status != "" {
		status := models.TicketStatus(req.Status)
		if !isValidTicketStatus(status) {
			return nil, nil, ErrInvalidTicketStatus
		}
		filter.Status = &status
	}

	// Parse tier ID if provided
	if req.TierID != "" {
		tierID, err := uuid.Parse(req.TierID)
		if err != nil {
			return nil, nil, ErrInvalidTierID
		}
		filter.TierID = &tierID
	}

	tickets, total, err := s.repos.Ticket.GetTicketsForAdmin(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination metadata using validated values
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	meta := &common.PaginationMeta{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalItems:  total,
	}

	return mappers.MapUserTicketsToResponse(tickets, true), meta, nil
}

// GetTicketByID returns a specific ticket by ID (admin)
func (s *TicketService) GetTicketByID(ctx context.Context, ticketID string) (*responses.UserTicketResponse, error) {
	id, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrInvalidTicketID
	}

	ticket, err := s.repos.Ticket.GetUserTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// ApproveTicket approves a ticket (admin action)
func (s *TicketService) ApproveTicket(ctx context.Context, ticketID string, staffID string) (*responses.UserTicketResponse, error) {
	tid, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrInvalidTicketID
	}

	sid, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	ticket, err := s.repos.Ticket.ApproveTicket(ctx, tid, sid)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// DenyTicket denies a ticket (admin action)
func (s *TicketService) DenyTicket(ctx context.Context, ticketID string, staffID string, req *requests.DenyTicketRequest) (*responses.UserTicketResponse, error) {
	tid, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrInvalidTicketID
	}

	sid, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	ticket, err := s.repos.Ticket.DenyTicket(ctx, tid, sid, req.Reason)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// GetTicketStatistics returns ticket statistics for admin dashboard
func (s *TicketService) GetTicketStatistics(ctx context.Context) (*responses.TicketStatisticsResponse, error) {
	stats, err := s.repos.Ticket.GetTicketStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return mappers.MapTicketStatisticsToResponse(stats), nil
}

// ========== Blacklist Management ==========

// GetBlacklistedUsers returns all blacklisted users
func (s *TicketService) GetBlacklistedUsers(ctx context.Context, page, pageSize int) ([]responses.BlacklistedUserResponse, *common.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	users, total, err := s.repos.Ticket.GetBlacklistedUsers(ctx, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	meta := &common.PaginationMeta{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalItems:  total,
	}

	return mappers.MapUsersToBlacklistedResponse(users), meta, nil
}

// BlacklistUser manually blacklists a user
func (s *TicketService) BlacklistUser(ctx context.Context, userID string, req *requests.BlacklistUserRequest) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}

	return s.repos.Ticket.BlacklistUser(ctx, id, req.Reason)
}

// UnblacklistUser removes a user from blacklist
func (s *TicketService) UnblacklistUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}

	return s.repos.Ticket.UnblacklistUser(ctx, id)
}

// ========== Helper Functions ==========

func isValidTicketStatus(status models.TicketStatus) bool {
	switch status {
	case models.TicketStatusPending,
		models.TicketStatusSelfConfirmed,
		models.TicketStatusApproved,
		models.TicketStatusDenied:
		return true
	default:
		return false
	}
}
