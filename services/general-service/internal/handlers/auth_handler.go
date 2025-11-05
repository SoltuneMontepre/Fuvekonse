package handlers

import (
	"general-service/internal/common/utils"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	services *services.Services
}

func NewAuthHandler(services *services.Services) *AuthHandler {
	return &AuthHandler{services: services}
}

// Login godoc
// @Summary Login to the system
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.LoginRequest true "Login credentials"
// @Success 200 {object} responses.LoginResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Call service
	response, err := h.services.Auth.Login(&req)
	if err != nil {
		if err.Error() == "invalid email or password" {
			utils.RespondUnauthorized(c, err.Error())
			return
		} else if err.Error() == "user is not verified" {
			utils.RespondForbidden(c, err.Error())
			return
		}
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	// Set JWT access token in cookie
	c.SetCookie(
		"access_token",
		response.AccessToken,
		int(utils.GetAccessTokenExpiry().Seconds()),
		"/",
		"",
		true, // secure
		true, // httpOnly
	)

	c.SetCookie(
		"refresh_token",
		response.RefreshToken,
		int(7*24*60*60), // 7 days
		"/",
		"",
		true,
		true,
	)

	utils.RespondSuccess(c, response, "Login successful")
}
