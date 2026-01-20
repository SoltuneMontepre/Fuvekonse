package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/ticket/requests"
	"general-service/internal/repositories"
	"general-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	services *services.Services
}

func NewTicketHandler(services *services.Services) *TicketHandler {
	return &TicketHandler{services: services}
}

// ========== Public User Endpoints ==========

// GetTiers godoc
// @Summary Get all available ticket tiers
// @Description Get a list of all active ticket tiers with pricing and benefits
// @Tags tickets
// @Accept json
// @Produce json
// @Success 200 "Successfully retrieved ticket tiers"
// @Failure 500 "Internal server error"
// @Router /tickets/tiers [get]
func (h *TicketHandler) GetTiers(c *gin.Context) {
	ctx := c.Request.Context()
	tiers, err := h.services.Ticket.GetAllTiers(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve ticket tiers")
		return
	}

	utils.RespondSuccess(c, &tiers, "Successfully retrieved ticket tiers")
}

// GetTierByID godoc
// @Summary Get a specific ticket tier
// @Description Get detailed information about a specific ticket tier
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Tier ID" format(uuid)
// @Success 200 "Successfully retrieved ticket tier"
// @Failure 400 "Invalid tier ID"
// @Failure 404 "Tier not found"
// @Failure 500 "Internal server error"
// @Router /tickets/tiers/{id} [get]
func (h *TicketHandler) GetTierByID(c *gin.Context) {
	ctx := c.Request.Context()
	tierID := c.Param("id")
	if tierID == "" {
		utils.RespondValidationError(c, "Tier ID is required")
		return
	}

	tier, err := h.services.Ticket.GetTierByID(ctx, tierID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTierID) {
			utils.RespondBadRequest(c, "Invalid tier ID format")
			return
		}
		if errors.Is(err, repositories.ErrTicketTierNotFound) {
			utils.RespondNotFound(c, "Ticket tier not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve ticket tier")
		return
	}

	utils.RespondSuccess(c, tier, "Successfully retrieved ticket tier")
}

// GetMyTicket godoc
// @Summary Get current user's ticket
// @Description Get the ticket information for the currently authenticated user
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved user ticket"
// @Failure 401 "Unauthorized"
// @Failure 500 "Internal server error"
// @Router /tickets/me [get]
func (h *TicketHandler) GetMyTicket(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	ticket, err := h.services.Ticket.GetMyTicket(ctx, userID.(string))
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve ticket")
		return
	}

	// ticket can be nil if user has no ticket
	utils.RespondSuccess(c, ticket, "Successfully retrieved user ticket")
}

// PurchaseTicket godoc
// @Summary Purchase a ticket
// @Description Purchase a ticket for a specific tier. Creates a pending ticket and decrements stock.
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.PurchaseTicketRequest true "Purchase request"
// @Success 201 "Ticket purchased successfully"
// @Failure 400 "Invalid request or tier ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "User is blacklisted"
// @Failure 409 "User already has a ticket or tier is out of stock"
// @Failure 500 "Internal server error"
// @Router /tickets/purchase [post]
func (h *TicketHandler) PurchaseTicket(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	var req requests.PurchaseTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	ticket, err := h.services.Ticket.PurchaseTicket(ctx, userID.(string), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTierID):
			utils.RespondBadRequest(c, "Invalid tier ID format")
		case errors.Is(err, repositories.ErrTicketTierNotFound):
			utils.RespondNotFound(c, "Ticket tier not found")
		case errors.Is(err, repositories.ErrOutOfStock):
			utils.RespondError(c, 409, "OUT_OF_STOCK", "This ticket tier is sold out")
		case errors.Is(err, repositories.ErrUserAlreadyHasTicket):
			utils.RespondError(c, 409, "ALREADY_HAS_TICKET", "You already have a ticket")
		case errors.Is(err, repositories.ErrUserBlacklisted):
			utils.RespondForbidden(c, "You are not allowed to purchase tickets. Contact support.")
		default:
			utils.RespondInternalServerError(c, "Failed to purchase ticket")
		}
		return
	}

	utils.RespondCreated(c, ticket, "Ticket purchased successfully. Please complete payment.")
}

// ConfirmPayment godoc
// @Summary Confirm payment for pending ticket
// @Description Mark the user's pending ticket as self-confirmed (user claims they have paid)
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Payment confirmation submitted"
// @Failure 401 "Unauthorized"
// @Failure 404 "No pending ticket found"
// @Failure 409 "Ticket is not in pending status"
// @Failure 500 "Internal server error"
// @Router /tickets/me/confirm [patch]
func (h *TicketHandler) ConfirmPayment(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	ticket, err := h.services.Ticket.ConfirmPayment(ctx, userID.(string))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoTicketFound):
			utils.RespondNotFound(c, "No pending ticket found")
		case errors.Is(err, repositories.ErrInvalidTicketStatus):
			utils.RespondError(c, 409, "INVALID_STATUS", "Ticket is not in pending status")
		default:
			utils.RespondInternalServerError(c, "Failed to confirm payment")
		}
		return
	}

	utils.RespondSuccess(c, ticket, "Payment confirmation submitted. Awaiting staff verification.")
}

// CancelTicket godoc
// @Summary Cancel a ticket
// @Description Cancel a pending or self_confirmed ticket and re-increment stock
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Ticket cancelled successfully"
// @Failure 401 "Unauthorized"
// @Failure 404 "No ticket found"
// @Failure 409 "Ticket cannot be cancelled (already approved or denied)"
// @Failure 500 "Internal server error"
// @Router /tickets/me/cancel [delete]
func (h *TicketHandler) CancelTicket(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	err := h.services.Ticket.CancelTicket(ctx, userID.(string))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoTicketFound):
			utils.RespondNotFound(c, "No ticket found")
		case errors.Is(err, repositories.ErrInvalidTicketStatus):
			utils.RespondError(c, 409, "INVALID_STATUS", "Ticket cannot be cancelled (already approved or denied)")
		default:
			utils.RespondInternalServerError(c, "Failed to cancel ticket")
		}
		return
	}

	emptyData := struct{}{}
	utils.RespondSuccess(c, &emptyData, "Ticket cancelled successfully. You can now purchase a new ticket.")
}

// UpdateBadgeDetails godoc
// @Summary Update badge details
// @Description Update badge details for an approved ticket (con badge name, image, fursuiter status)
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateBadgeDetailsRequest true "Badge details"
// @Success 200 "Badge details updated"
// @Failure 400 "Invalid request"
// @Failure 401 "Unauthorized"
// @Failure 404 "No approved ticket found"
// @Failure 409 "Ticket is not approved"
// @Failure 500 "Internal server error"
// @Router /tickets/me/badge [patch]
func (h *TicketHandler) UpdateBadgeDetails(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	var req requests.UpdateBadgeDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	ticket, err := h.services.Ticket.UpdateBadgeDetails(ctx, userID.(string), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoTicketFound):
			utils.RespondNotFound(c, "No ticket found")
		case errors.Is(err, repositories.ErrInvalidTicketStatus):
			utils.RespondError(c, 409, "INVALID_STATUS", "Ticket must be approved to update badge details")
		default:
			utils.RespondInternalServerError(c, "Failed to update badge details")
		}
		return
	}

	utils.RespondSuccess(c, ticket, "Badge details updated successfully")
}

// ========== Admin Endpoints ==========

// GetTicketsForAdmin godoc
// @Summary Get all tickets (admin)
// @Description Get a paginated list of all tickets with filters
// @Tags admin-tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, self_confirmed, approved, denied)"
// @Param tier_id query string false "Filter by tier ID"
// @Param search query string false "Search by reference code, user name, or email"
// @Param pending_over_24 query bool false "Only show tickets pending > 24 hours"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 "Successfully retrieved tickets"
// @Failure 400 "Invalid filter parameters"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/tickets [get]
func (h *TicketHandler) GetTicketsForAdmin(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	req := &requests.AdminTicketFilterRequest{
		Status:        c.Query("status"),
		TierID:        c.Query("tier_id"),
		Search:        c.Query("search"),
		PendingOver24: c.Query("pending_over_24") == "true",
		Page:          page,
		PageSize:      pageSize,
	}

	tickets, meta, err := h.services.Ticket.GetTicketsForAdmin(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTicketStatus):
			utils.RespondBadRequest(c, "Invalid status filter")
		case errors.Is(err, services.ErrInvalidTierID):
			utils.RespondBadRequest(c, "Invalid tier ID format")
		default:
			utils.RespondInternalServerError(c, "Failed to retrieve tickets")
		}
		return
	}

	utils.RespondSuccessWithMeta(c, &tickets, meta, "Successfully retrieved tickets")
}

// GetTicketByID godoc
// @Summary Get ticket by ID (admin)
// @Description Get detailed information about a specific ticket
// @Tags admin-tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID" format(uuid)
// @Success 200 "Successfully retrieved ticket"
// @Failure 400 "Invalid ticket ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 404 "Ticket not found"
// @Failure 500 "Internal server error"
// @Router /admin/tickets/{id} [get]
func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	ctx := c.Request.Context()
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.RespondBadRequest(c, "Ticket ID is required")
		return
	}

	ticket, err := h.services.Ticket.GetTicketByID(ctx, ticketID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTicketID):
			utils.RespondBadRequest(c, "Invalid ticket ID format")
		case errors.Is(err, repositories.ErrTicketNotFound):
			utils.RespondNotFound(c, "Ticket not found")
		default:
			utils.RespondInternalServerError(c, "Failed to retrieve ticket")
		}
		return
	}

	utils.RespondSuccess(c, ticket, "Successfully retrieved ticket")
}

// ApproveTicket godoc
// @Summary Approve a ticket (admin)
// @Description Approve a pending or self-confirmed ticket
// @Tags admin-tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID" format(uuid)
// @Success 200 "Ticket approved successfully"
// @Failure 400 "Invalid ticket ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 404 "Ticket not found"
// @Failure 409 "Ticket cannot be approved (wrong status)"
// @Failure 500 "Internal server error"
// @Router /admin/tickets/{id}/approve [patch]
func (h *TicketHandler) ApproveTicket(c *gin.Context) {
	ctx := c.Request.Context()
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.RespondBadRequest(c, "Ticket ID is required")
		return
	}

	staffID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "Staff ID not found in token")
		return
	}

	ticket, err := h.services.Ticket.ApproveTicket(ctx, ticketID, staffID.(string))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTicketID):
			utils.RespondBadRequest(c, "Invalid ticket ID format")
		case errors.Is(err, repositories.ErrTicketNotFound):
			utils.RespondNotFound(c, "Ticket not found")
		case errors.Is(err, repositories.ErrInvalidTicketStatus):
			utils.RespondError(c, 409, "INVALID_STATUS", "Ticket cannot be approved (wrong status)")
		default:
			utils.RespondInternalServerError(c, "Failed to approve ticket")
		}
		return
	}

	utils.RespondSuccess(c, ticket, "Ticket approved successfully")
}

// DenyTicket godoc
// @Summary Deny a ticket (admin)
// @Description Deny a pending or self-confirmed ticket and return stock
// @Tags admin-tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID" format(uuid)
// @Param request body requests.DenyTicketRequest true "Denial reason (optional)"
// @Success 200 "Ticket denied successfully"
// @Failure 400 "Invalid ticket ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 404 "Ticket not found"
// @Failure 409 "Ticket cannot be denied (wrong status)"
// @Failure 500 "Internal server error"
// @Router /admin/tickets/{id}/deny [patch]
func (h *TicketHandler) DenyTicket(c *gin.Context) {
	ctx := c.Request.Context()
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.RespondBadRequest(c, "Ticket ID is required")
		return
	}

	staffID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "Staff ID not found in token")
		return
	}

	var req requests.DenyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Denial reason is optional, so we don't fail on binding error
		req = requests.DenyTicketRequest{}
	}

	ticket, err := h.services.Ticket.DenyTicket(ctx, ticketID, staffID.(string), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTicketID):
			utils.RespondBadRequest(c, "Invalid ticket ID format")
		case errors.Is(err, repositories.ErrTicketNotFound):
			utils.RespondNotFound(c, "Ticket not found")
		case errors.Is(err, repositories.ErrInvalidTicketStatus):
			utils.RespondError(c, 409, "INVALID_STATUS", "Ticket cannot be denied (wrong status)")
		default:
			utils.RespondInternalServerError(c, "Failed to deny ticket")
		}
		return
	}

	utils.RespondSuccess(c, ticket, "Ticket denied successfully")
}

// GetTicketStatistics godoc
// @Summary Get ticket statistics (admin)
// @Description Get ticket statistics for admin dashboard
// @Tags admin-tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved ticket statistics"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/tickets/statistics [get]
func (h *TicketHandler) GetTicketStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.services.Ticket.GetTicketStatistics(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve ticket statistics")
		return
	}

	utils.RespondSuccess(c, stats, "Successfully retrieved ticket statistics")
}

// ========== Blacklist Management ==========

// GetBlacklistedUsers godoc
// @Summary Get blacklisted users (admin)
// @Description Get a paginated list of blacklisted users
// @Tags admin-users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 "Successfully retrieved blacklisted users"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/users/blacklisted [get]
func (h *TicketHandler) GetBlacklistedUsers(c *gin.Context) {
	ctx := c.Request.Context()

	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	users, meta, err := h.services.Ticket.GetBlacklistedUsers(ctx, page, pageSize)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve blacklisted users")
		return
	}

	utils.RespondSuccessWithMeta(c, &users, meta, "Successfully retrieved blacklisted users")
}

// BlacklistUser godoc
// @Summary Blacklist a user (admin)
// @Description Manually blacklist a user from purchasing tickets
// @Tags admin-users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Param request body requests.BlacklistUserRequest true "Blacklist reason"
// @Success 200 "User blacklisted successfully"
// @Failure 400 "Invalid user ID or missing reason"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id}/blacklist [patch]
func (h *TicketHandler) BlacklistUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	var req requests.BlacklistUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	err := h.services.Ticket.BlacklistUser(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) {
			utils.RespondBadRequest(c, "Invalid user ID format")
			return
		}
		utils.RespondInternalServerError(c, "Failed to blacklist user")
		return
	}

	utils.RespondSuccess[any](c, nil, "User blacklisted successfully")
}

// UnblacklistUser godoc
// @Summary Remove user from blacklist (admin)
// @Description Remove a user from the blacklist, allowing them to purchase tickets again
// @Tags admin-users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Success 200 "User removed from blacklist"
// @Failure 400 "Invalid user ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id}/unblacklist [patch]
func (h *TicketHandler) UnblacklistUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	err := h.services.Ticket.UnblacklistUser(ctx, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) {
			utils.RespondBadRequest(c, "Invalid user ID format")
			return
		}
		utils.RespondInternalServerError(c, "Failed to remove user from blacklist")
		return
	}

	utils.RespondSuccess[any](c, nil, "User removed from blacklist successfully")
}
