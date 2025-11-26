package handlers

import (
	"errors"
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/services"
	"os"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	services     *services.Services
	cookieConfig utils.CookieConfig
}

// getEnvOr returns the environment variable value or a default value if not set
func getEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewAuthHandler(services *services.Services) *AuthHandler {
	// Load cookie configuration from environment
	secure := getEnvOr("COOKIE_SECURE", "true") != "false" // Default to true
	sameSite := getEnvOr("COOKIE_SAMESITE", "Strict")      // Default to Strict for security

	cookieConfig := utils.CookieConfig{
		Domain:   getEnvOr("COOKIE_DOMAIN", ""),
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   int(utils.GetAccessTokenExpiry().Seconds()),
	}

	return &AuthHandler{
		services:     services,
		cookieConfig: cookieConfig,
	}
}

// Register godoc
// @Summary Register a new user account
// @Description Create a new user account with email verification. An OTP will be sent to the provided email.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.RegisterRequest true "Registration details"
// @Success 200 "Registration successful, OTP sent to email"
// @Failure 400 "Bad request - validation error or passwords don't match"
// @Failure 409 "Conflict - user already exists"
// @Failure 500 "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	fromEmail := getEnvOr("SES_EMAIL_IDENTITY", "")
	
	response, err := h.services.Auth.Register(c.Request.Context(), &req, h.services.Mail, fromEmail)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "user with this email already exists" {
			utils.RespondError(c, 409, "USER_EXISTS", errMsg)
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondBadRequest(c, "Passwords do not match")
			return
		}
		utils.RespondInternalServerError(c, "Failed to register user")
		return
	}

	utils.RespondSuccess(c, response, response.Message)
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

	// Call service with context
	response, err := h.services.Auth.Login(c.Request.Context(), &req)
	if err != nil {
		// Check if it's a rate limit error using sentinel
		if errors.Is(err, constants.ErrAccountLocked) {
			utils.RespondTooManyRequests(c, "Too many failed login attempts")
			return
		}

		// Check for invalid credentials
		if errors.Is(err, constants.ErrInvalidCredentials) {
			utils.RespondUnauthorized(c, err.Error())
			return
		}

		// Check for unverified user
		if errors.Is(err, constants.ErrUserNotVerified) {
			utils.RespondForbidden(c, err.Error())
			return
		}

		// Default to internal server error
		utils.RespondInternalServerError(c, "An error occurred during login")
		return
	}

	// Set JWT access token in cookie using helper
	utils.SetAuthCookie(c, response.AccessToken, h.cookieConfig)
	utils.RespondSuccess[any](c, nil, "Login successful")
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

	// Extract user ID from context (stored as string from JWT)
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
		// Use sentinel errors for consistent error handling
		if errors.Is(err, constants.ErrUserNotFound) {
			utils.RespondNotFound(c, err.Error())
			return
		}
		if errors.Is(err, constants.ErrCurrentPasswordIncorrect) {
			utils.RespondUnauthorized(c, err.Error())
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) || errors.Is(err, constants.ErrSamePassword) {
			utils.RespondBadRequest(c, err.Error())
			return
		}

		// Default to internal server error for unknown errors
		utils.RespondInternalServerError(c, "Failed to reset password")
		return
	}

	utils.RespondSuccess[any](c, nil, "Password reset successful")
}

// Logout godoc
// @Summary Logout from the system
// @Description Remove access token cookie to logout user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully logged out"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the access token cookie using helper
	utils.ClearAuthCookie(c, h.cookieConfig)
	utils.RespondSuccess[any](c, nil, "Logout successful")
}

// VerifyOtp godoc
// @Summary Verify OTP code
// @Description Verify the OTP code sent to user's email and mark account as verified
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.VerifyOtpRequest true "Verify OTP request"
// @Success 200 "Email verified successfully"
// @Failure 400 "Bad request - validation error or invalid/expired OTP"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOtp(c *gin.Context) {
	var req requests.VerifyOtpRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	success, err := h.services.Auth.VerifyOtp(ctx, req.Email, req.Otp)

	if err != nil {
		errMsg := err.Error()
		if constants.ErrCodeNotFound == errMsg {
			utils.RespondNotFound(c, errMsg)
			return
		}
		utils.RespondInternalServerError(c, errMsg)
		return
	}

	if !success {
		utils.RespondBadRequest(c, "Invalid or expired OTP")
		return
	}

	utils.RespondSuccess[any](c, nil, "Email verified successfully")
}

// ForgotPassword godoc
// @Tags auth
// @Accept json
// @Produce json
// @Summary Request password reset token (sent by email)
// @Param request body requests.ForgotPasswordRequest true "Forgot password request"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req requests.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	fromEmail := getEnvOr("SES_EMAIL_IDENTITY", "") //test cuz idk
	frontendURL := getEnvOr("FRONTEND_URL", "")

	if err := h.services.Auth.ForgotPassword(c.Request.Context(), req.Email, h.services.Mail, frontendURL, fromEmail); err != nil {
		utils.RespondInternalServerError(c, "Failed to process password reset request")
		return
	}

	utils.RespondSuccess[any](c, nil, "If an account with that email exists, a password reset token has been sent")
}

// ResetPasswordConfirm godoc
// @Tags auth
// @Accept json
// @Produce json
// @Summary Reset password using token
// @Param request body requests.ResetPasswordTokenRequest true "Reset password with token"
// @Router /auth/reset-password/confirm [post]
func (h *AuthHandler) ResetPasswordConfirm(c *gin.Context) {
	var req requests.ResetPasswordTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if err := h.services.Auth.ResetPasswordWithToken(req.Token, &req); err != nil {
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, constants.ErrUserNotFound) {
			utils.RespondNotFound(c, err.Error())
			return
		}
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondSuccess[any](c, nil, "Password has been reset successfully")
}
