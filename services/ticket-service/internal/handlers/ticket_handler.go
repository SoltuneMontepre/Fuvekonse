package handlers

import (
	"net/http"
	"ticket-service/internal/dto"
	"ticket-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TicketHandler struct {
	ticketService services.TicketServiceInterface
}

func NewTicketHandler(ticketService services.TicketServiceInterface) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// GetTicketTiers godoc
// @Summary Get all ticket tiers
// @Description Get a list of all ticket tiers
// @Tags tickets
// @Produce json
// @Success 200 {array} dto.TicketTierResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tickets/tiers [get]
func (h *TicketHandler) GetTicketTiers(c *gin.Context) {
	tiers, err := h.ticketService.GetAllTicketTiers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	responses := make([]dto.TicketTierResponse, len(tiers))
	for i, tier := range tiers {
		responses[i] = dto.TicketTierResponse{
			ID:          tier.ID,
			TicketName:  tier.TicketName,
			Description: tier.Description,
			Price:       tier.Price,
			Stock:       tier.Stock,
			IsActive:    tier.IsActive,
			BannerImage: tier.BannerImage,
		}
	}

	c.JSON(http.StatusOK, responses)
}

// GetActiveTicketTiers godoc
// @Summary Get active ticket tiers
// @Description Get a list of active ticket tiers
// @Tags tickets
// @Produce json
// @Success 200 {array} dto.TicketTierResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tickets/tiers/active [get]
func (h *TicketHandler) GetActiveTicketTiers(c *gin.Context) {
	tiers, err := h.ticketService.GetActiveTicketTiers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	responses := make([]dto.TicketTierResponse, len(tiers))
	for i, tier := range tiers {
		responses[i] = dto.TicketTierResponse{
			ID:          tier.ID,
			TicketName:  tier.TicketName,
			Description: tier.Description,
			Price:       tier.Price,
			Stock:       tier.Stock,
			IsActive:    tier.IsActive,
			BannerImage: tier.BannerImage,
		}
	}

	c.JSON(http.StatusOK, responses)
}

// GetTicketTier godoc
// @Summary Get ticket tier by ID
// @Description Get a specific ticket tier by ID
// @Tags tickets
// @Produce json
// @Param id path string true "Ticket Tier ID"
// @Success 200 {object} dto.TicketTierResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /tickets/tiers/{id} [get]
func (h *TicketHandler) GetTicketTier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid ticket tier ID"})
		return
	}

	tier, err := h.ticketService.GetTicketTierByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.TicketTierResponse{
		ID:          tier.ID,
		TicketName:  tier.TicketName,
		Description: tier.Description,
		Price:       tier.Price,
		Stock:       tier.Stock,
		IsActive:    tier.IsActive,
		BannerImage: tier.BannerImage,
	}

	c.JSON(http.StatusOK, response)
}

// GetTicket godoc
// @Summary Get ticket by ID
// @Description Get a specific ticket by ID
// @Tags tickets
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 200 {object} dto.TicketResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid ticket ID"})
		return
	}

	ticket, err := h.ticketService.GetTicketByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.TicketResponse{
		ID:          ticket.ID,
		UserID:      ticket.UserID,
		TicketTierID: ticket.TicketTierID,
		OrderCode:   ticket.OrderCode,
		Status:      ticket.Status,
		Price:       ticket.Price,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserTickets godoc
// @Summary Get tickets by user ID
// @Description Get all tickets for a specific user
// @Tags tickets
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} dto.TicketResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tickets/user/{user_id} [get]
func (h *TicketHandler) GetUserTickets(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user ID"})
		return
	}

	tickets, err := h.ticketService.GetTicketsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	responses := make([]dto.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = dto.TicketResponse{
			ID:          ticket.ID,
			UserID:      ticket.UserID,
			TicketTierID: ticket.TicketTierID,
			OrderCode:   ticket.OrderCode,
			Status:      ticket.Status,
			Price:       ticket.Price,
		}
	}

	c.JSON(http.StatusOK, responses)
}

