package handlers

import (
	"general-service/internal/dto/user_ticket/requests"
	"general-service/internal/dto/user_ticket/responses"
	"general-service/internal/models"
	"general-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserTicketHandler struct {
	userTicketService services.UserTicketServiceInterface
}

func NewUserTicketHandler(userTicketService services.UserTicketServiceInterface) *UserTicketHandler {
	return &UserTicketHandler{
		userTicketService: userTicketService,
	}
}

// CreateUserTicket godoc
// @Summary Create a user ticket
// @Description Create a user ticket after successful payment
// @Tags user-tickets
// @Accept json
// @Produce json
// @Param userTicket body dto.CreateUserTicketRequest true "User ticket data"
// @Success 201 {object} dto.UserTicketResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user-tickets [post]
func (h *UserTicketHandler) CreateUserTicket(c *gin.Context) {
	var req requests.CreateUserTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userTicket := &models.UserTicket{
		Id:             uuid.New(),
		UserId:         req.UserId,
		TicketId:       req.TicketId,
		ConBadgeName:   req.ConBadgeName,
		BadgeImage:     req.BadgeImage,
		IsFursuiter:    req.IsFursuiter,
		IsFursuitStaff: req.IsFursuitStaff,
		IsCheckedIn:    req.IsCheckedIn,
	}

	if err := h.userTicketService.CreateUserTicket(userTicket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := responses.UserTicketResponse{
		Id:             userTicket.Id,
		UserId:         userTicket.UserId,
		TicketId:       userTicket.TicketId,
		ConBadgeName:   userTicket.ConBadgeName,
		BadgeImage:     userTicket.BadgeImage,
		IsFursuiter:    userTicket.IsFursuiter,
		IsFursuitStaff: userTicket.IsFursuitStaff,
		IsCheckedIn:    userTicket.IsCheckedIn,
		CreatedAt:      userTicket.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}
