package services

import (
	"context"
	"encoding/json"
	"general-service/internal/common/constants"
	"general-service/internal/dto/common"
	"general-service/internal/dto/ticket/requests"
	"general-service/internal/dto/ticket/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"log"
	"math"
	"os"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	mail  *MailService
}

func NewTicketService(repos *repositories.Repositories, mail *MailService) *TicketService {
	return &TicketService{repos: repos, mail: mail}
}

// ========== Public User Endpoints ==========

// GetAllTiers returns all ticket tiers (active and deactivated; excludes deleted).
func (s *TicketService) GetAllTiers(ctx context.Context) ([]responses.TicketTierResponse, error) {
	tiers, err := s.repos.Ticket.GetAllTiersForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return mappers.MapTicketTiersToResponse(tiers), nil
}

// GetAllTiersForAdmin returns all non-deleted tiers (active and inactive) for admin.
func (s *TicketService) GetAllTiersForAdmin(ctx context.Context) ([]responses.TicketTierResponse, error) {
	tiers, err := s.repos.Ticket.GetAllTiersForAdmin(ctx)
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

// CreateTierForAdmin creates a new ticket tier (admin only)
func (s *TicketService) CreateTierForAdmin(ctx context.Context, req *requests.CreateTicketTierRequest) (*responses.TicketTierResponse, error) {
	benefitsJSON := ""
	if len(req.Benefits) > 0 {
		b, err := json.Marshal(req.Benefits)
		if err != nil {
			return nil, err
		}
		benefitsJSON = string(b)
	}

	price := 0.0
	if req.Price != nil {
		price = *req.Price
	}
	stock := 0
	if req.Stock != nil {
		stock = *req.Stock
	}

	tier := &models.TicketTier{
		TicketName:  req.TicketName,
		Description: req.Description,
		Benefits:    benefitsJSON,
		Price:       decimal.NewFromFloat(price),
		Stock:       stock,
		IsActive:    req.IsActive,
	}

	created, err := s.repos.Ticket.CreateTier(ctx, tier)
	if err != nil {
		return nil, err
	}
	return mappers.MapTicketTierToResponse(created), nil
}

// UpdateTierForAdmin updates a ticket tier (admin only). Only provided fields are updated.
func (s *TicketService) UpdateTierForAdmin(ctx context.Context, tierID string, req *requests.UpdateTicketTierRequest) (*responses.TicketTierResponse, error) {
	id, err := uuid.Parse(tierID)
	if err != nil {
		return nil, ErrInvalidTierID
	}
	updates := make(map[string]interface{})
	if req.TicketName != nil {
		updates["ticket_name"] = *req.TicketName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Benefits != nil {
		benefitsJSON := ""
		if len(req.Benefits) > 0 {
			b, err := json.Marshal(req.Benefits)
			if err != nil {
				return nil, err
			}
			benefitsJSON = string(b)
		}
		updates["benefits"] = benefitsJSON
	}
	if req.Price != nil {
		updates["price"] = decimal.NewFromFloat(*req.Price)
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) == 0 {
		// No updates, just return current tier
		tier, err := s.repos.Ticket.GetTierByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return mappers.MapTicketTierToResponse(tier), nil
	}
	tier, err := s.repos.Ticket.UpdateTier(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	return mappers.MapTicketTierToResponse(tier), nil
}

// DeleteTierForAdmin permanently deletes a ticket tier and its user tickets (admin only).
func (s *TicketService) DeleteTierForAdmin(ctx context.Context, tierID string) error {
	id, err := uuid.Parse(tierID)
	if err != nil {
		return ErrInvalidTierID
	}
	return s.repos.Ticket.DeleteTier(ctx, id)
}

// SetTierActiveForAdmin sets is_active for a ticket tier (admin only).
func (s *TicketService) SetTierActiveForAdmin(ctx context.Context, tierID string, active bool) (*responses.TicketTierResponse, error) {
	id, err := uuid.Parse(tierID)
	if err != nil {
		return nil, ErrInvalidTierID
	}
	tier, err := s.repos.Ticket.SetTierActive(ctx, id, active)
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

// GetTicketByID returns a specific ticket by ID (UUID) or by reference code (admin/staff).
// If the input parses as a valid UUID, lookup is by ticket id; otherwise by reference code.
func (s *TicketService) GetTicketByID(ctx context.Context, ticketIDOrRef string) (*responses.UserTicketResponse, error) {
	ticket, err := s.getTicketByIDOrRef(ctx, ticketIDOrRef)
	if err != nil {
		return nil, err
	}
	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// getTicketByIDOrRef returns the ticket model by UUID or reference code. Caller must not pass empty string.
func (s *TicketService) getTicketByIDOrRef(ctx context.Context, ticketIDOrRef string) (*models.UserTicket, error) {
	id, err := uuid.Parse(ticketIDOrRef)
	if err == nil {
		return s.repos.Ticket.GetUserTicketByID(ctx, id)
	}
	return s.repos.Ticket.GetUserTicketByReference(ctx, ticketIDOrRef)
}

// ConfirmCheckIn sets is_checked_in = true for a ticket (admin/staff). Accepts ticket ID or reference code.
func (s *TicketService) ConfirmCheckIn(ctx context.Context, ticketIDOrRef, staffID string) (*responses.UserTicketResponse, error) {
	if ticketIDOrRef == "" {
		return nil, ErrInvalidTicketID
	}
	ticket, err := s.getTicketByIDOrRef(ctx, ticketIDOrRef)
	if err != nil {
		return nil, err
	}
	sid, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	updated, err := s.repos.Ticket.UpdateTicketForAdmin(ctx, ticket.Id, map[string]interface{}{"is_checked_in": true}, sid)
	if err != nil {
		return nil, err
	}
	return mappers.MapUserTicketToResponse(updated, true), nil
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

	// Send ticket approved email with QR code to the user
	if s.mail != nil && ticket.User.Email != "" {
		fromEmail := os.Getenv("SES_EMAIL_IDENTITY")
		if fromEmail != "" {
			tierName := ""
			if ticket.Ticket.TicketName != "" {
				tierName = ticket.Ticket.TicketName
			}
			if err := s.mail.SendTicketApprovedWithQREmail(ctx, fromEmail, ticket.User.Email, ticket.ReferenceCode, tierName, LangFromCountry(ticket.User.Country)); err != nil {
				log.Printf("Failed to send ticket approved email with QR to %s: %v", ticket.User.Email, err)
			}
		}
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

// CreateTicketForAdmin creates a ticket for a user (admin). Ticket is created as approved.
func (s *TicketService) CreateTicketForAdmin(ctx context.Context, staffID string, req *requests.CreateTicketForAdminRequest) (*responses.UserTicketResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	tierID, err := uuid.Parse(req.TierID)
	if err != nil {
		return nil, ErrInvalidTierID
	}

	staffUUID, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	ticket, err := s.repos.Ticket.CreateTicketForAdmin(ctx, userID, tierID, staffUUID)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// UpdateTicketForAdmin updates any ticket fields without validation (admin back-door).
func (s *TicketService) UpdateTicketForAdmin(ctx context.Context, ticketID, staffID string, req *requests.UpdateTicketForAdminRequest) (*responses.UserTicketResponse, error) {
	tid, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrInvalidTicketID
	}
	sid, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.TierID != nil {
		tierID, err := uuid.Parse(*req.TierID)
		if err != nil {
			return nil, ErrInvalidTierID
		}
		updates["ticket_id"] = tierID
	}
	if req.ConBadgeName != nil {
		updates["con_badge_name"] = *req.ConBadgeName
	}
	if req.BadgeImage != nil {
		updates["badge_image"] = *req.BadgeImage
	}
	if req.IsFursuiter != nil {
		updates["is_fursuiter"] = *req.IsFursuiter
	}
	if req.IsFursuitStaff != nil {
		updates["is_fursuit_staff"] = *req.IsFursuitStaff
	}
	if req.IsCheckedIn != nil {
		updates["is_checked_in"] = *req.IsCheckedIn
	}
	if req.DenialReason != nil {
		updates["denial_reason"] = *req.DenialReason
	}

	if len(updates) == 0 {
		return s.GetTicketByID(ctx, ticketID)
	}

	ticket, err := s.repos.Ticket.UpdateTicketForAdmin(ctx, tid, updates, sid)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
}

// DeleteTicketForAdmin soft-deletes a ticket and re-increments stock if applicable.
func (s *TicketService) DeleteTicketForAdmin(ctx context.Context, ticketID string) (*responses.UserTicketResponse, error) {
	tid, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ErrInvalidTicketID
	}

	ticket, err := s.repos.Ticket.DeleteTicketForAdmin(ctx, tid)
	if err != nil {
		return nil, err
	}

	return mappers.MapUserTicketToResponse(ticket, true), nil
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
