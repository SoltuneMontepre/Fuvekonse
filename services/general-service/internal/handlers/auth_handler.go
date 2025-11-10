package handlers

import (
	"errors"
	"general-service/internal/common/constants"
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
// @Description Authenticate user with email and password, returns access token and refresh token
// @Description
// @Description **Usage:**
// @Description 1. Send POST request with email and password
// @Description 2. Receive access_token and refresh_token
// @Description 3. Use access_token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.LoginRequest true "Login credentials"
// @Success 200 "Successfully logged in"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - invalid credentials"
// @Failure 429 "Too many failed login attempts - account temporarily locked"
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
		errMsg := err.Error()

		// Check if it's a rate limit error using sentinel
		if errors.Is(err, constants.ErrAccountLocked) {
			utils.RespondTooManyRequests(c, "Too many failed login attempts")
			return
		}

		switch errMsg {
		case "invalid email or password":
			utils.RespondUnauthorized(c, errMsg)
			return
		case "user is not verified":
			utils.RespondForbidden(c, errMsg)
			return
		}
		utils.RespondInternalServerError(c, errMsg)
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
	utils.RespondSuccess(c, response, "Login successful")
}

// ResetPassword godoc
// @Summary Reset user password
// @Description Allow authenticated users to reset their password after logged in!
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.ResetPasswordRequest true "Reset password request"
// @Success 200 "Password reset successful"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req requests.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

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

	if err := h.services.Auth.ResetPassword(userID, &req); err != nil {
		errMsg := err.Error()

		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case
			"failed to hash password",
			"failed to update password":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondBadRequest(c, errMsg)
		}
		return
	}

	utils.RespondSuccess[any](c, nil, "Password reset successful")
}
