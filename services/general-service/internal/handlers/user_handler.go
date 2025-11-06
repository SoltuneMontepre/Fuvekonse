package handlers

import (
	"general-service/internal/common/utils"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	services *services.Services
}

func NewUserHandler(services *services.Services) *UserHandler {
	return &UserHandler{services: services}
}

// GetMe godoc
// @Summary Get current user information
// @Description Get the profile information of the currently authenticated user
// @Description
// @Description **Usage:**
// @Description 1. Include JWT access token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Description 2. Receive your user profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved user information"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	// Get user ID from JWT context (set by JWTAuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	// Call service to get user information
	user, err := h.services.User.GetUserByID(userID.(string))
	if err != nil {
		if err.Error() == "record not found" {
			utils.RespondNotFound(c, "User not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve user information")
		return
	}

	utils.RespondSuccess(c, user, "Successfully retrieved user information")
}
