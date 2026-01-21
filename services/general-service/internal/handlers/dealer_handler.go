package handlers

import (
	"general-service/internal/common/utils"
	"general-service/internal/dto/dealer/requests"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type DealerHandler struct {
	services *services.Services
}

func NewDealerHandler(services *services.Services) *DealerHandler {
	return &DealerHandler{services: services}
}

// RegisterDealer godoc
// @Summary Register as a dealer
// @Description Create a new dealer booth and assign the current user as the owner
// @Description Requires authentication. User will become the owner of the dealer booth.
// @Description
// @Description **Usage:**
// @Description 1. Include JWT access token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Description 2. Provide booth information (booth_name, description, price_sheet)
// @Description 3. Receive the created dealer booth information
// @Tags dealer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.DealerRegisterRequest true "Dealer registration request"
// @Success 201 "Successfully registered as dealer"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 409 "Conflict - user is already a dealer staff"
// @Failure 500 "Internal server error"
// @Router /dealer/register [post]
func (h *DealerHandler) RegisterDealer(c *gin.Context) {
	var req requests.DealerRegisterRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Get user ID from JWT context
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		utils.RespondUnauthorized(c, "Invalid user ID in token")
		return
	}

	// Call service to register dealer
	booth, err := h.services.Dealer.RegisterDealer(userID, &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "user is already registered as a dealer staff":
			utils.RespondError(c, 409, "CONFLICT", errMsg)
		case "failed to register dealer":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to register dealer")
		}
		return
	}

	utils.RespondCreated(c, booth, "Successfully registered as dealer")
}
