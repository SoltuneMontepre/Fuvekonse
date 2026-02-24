package handlers

import (
	"errors"
	"fmt"
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/services"
	"os"
	"strings"

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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	fromEmail := getEnvOr("SES_EMAIL_IDENTITY", "")

	response, err := h.services.Auth.Register(c.Request.Context(), &req, h.services.Mail, fromEmail)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "user with this email already exists" {
			utils.RespondErrorWithErrorMessage(c, 409, "USER_EXISTS", errMsg, "userExists")
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, "Passwords do not match", "passwordsDoNotMatch")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "Failed to register user", "registerFailed")
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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	// Guard: ensure services are initialized
	if h.services == nil || h.services.Auth == nil {
		fmt.Printf("[ERROR] Auth handler called but services.Auth is nil: services=%v\n", h.services)
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "An error occurred during login", "loginFailed")
		return
	}

	// Call service with context
	response, err := h.services.Auth.Login(c.Request.Context(), &req)
	if err != nil {
		// Check if it's a rate limit error using sentinel
		if errors.Is(err, constants.ErrAccountLocked) {
			utils.RespondErrorWithErrorMessage(c, 429, constants.ErrCodeTooManyRequests, "Too many failed login attempts", "accountLocked")
			return
		}

		// Check for invalid credentials — return i18n key for frontend
		if errors.Is(err, constants.ErrInvalidCredentials) {
			utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "Invalid credentials", "invalidEmailOrPassword")
			return
		}

		// Check for unverified user
		if errors.Is(err, constants.ErrUserNotVerified) {
			utils.RespondErrorWithErrorMessage(c, 403, constants.ErrCodeForbidden, "User not verified", "userNotVerified")
			return
		}

		// Default to internal server error
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "An error occurred during login", "loginFailed")
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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	// Extract user ID from context (stored as string from JWT)
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "User ID not found in token", "unauthorized")
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "Invalid user ID in token", "unauthorized")
		return
	}

	if err := h.services.Auth.ResetPassword(userID, &req); err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			utils.RespondErrorWithErrorMessage(c, 404, constants.ErrCodeNotFound, err.Error(), "userNotFound")
			return
		}
		if errors.Is(err, constants.ErrCurrentPasswordIncorrect) {
			utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, err.Error(), "currentPasswordIncorrect")
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "passwordsDoNotMatch")
			return
		}
		if errors.Is(err, constants.ErrSamePassword) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "samePassword")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "Failed to reset password", "resetPasswordFailed")
		return
	}

	utils.RespondSuccess[any](c, nil, "Password reset successful")
}

// ChangePassword godoc
// @Summary Change password (authenticated)
// @Description Allow authenticated users to change their password by providing current password and new password.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.ResetPasswordRequest true "Change password request (current_password, new_password, confirm_password)"
// @Success 200 "Password changed successfully"
// @Failure 400 "Bad request - validation error or new password same as current"
// @Failure 401 "Unauthorized - missing/invalid token or wrong current password"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req requests.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "User ID not found in token", "unauthorized")
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "Invalid user ID in token", "unauthorized")
		return
	}

	if err := h.services.Auth.ResetPassword(userID, &req); err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			utils.RespondErrorWithErrorMessage(c, 404, constants.ErrCodeNotFound, err.Error(), "userNotFound")
			return
		}
		if errors.Is(err, constants.ErrCurrentPasswordIncorrect) {
			utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, err.Error(), "currentPasswordIncorrect")
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "passwordsDoNotMatch")
			return
		}
		if errors.Is(err, constants.ErrSamePassword) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "samePassword")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "Failed to change password", "changePasswordFailed")
		return
	}

	utils.RespondSuccess[any](c, nil, "Password changed successfully")
}

// GoogleLogin godoc
// @Summary Login or register with Google
// @Description Verify Google ID token (credential from Sign-In). If user exists (by Google ID or email), log in; otherwise create account and log in. Sets access token cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.GoogleLoginRequest true "Google ID token (credential)"
// @Success 200 "Logged in or registered successfully"
// @Failure 400 "Bad request - missing or invalid credential"
// @Failure 401 "Invalid Google token"
// @Failure 500 "Internal server error"
// @Router /auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req requests.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}
	googleClientID := getEnvOr("GOOGLE_CLIENT_ID", "")
	if googleClientID == "" {
		utils.RespondErrorWithErrorMessage(c, 503, "SERVICE_UNAVAILABLE", "Google sign-in is not configured", "googleNotConfigured")
		return
	}
	response, err := h.services.Auth.GoogleLoginOrRegister(c.Request.Context(), &req, googleClientID)
	if err != nil {
		if errors.Is(err, constants.ErrGoogleRegistrationDetailsRequired) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "googleRegistrationDetailsRequired")
			return
		}
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, "Passwords do not match", "passwordsDoNotMatch")
			return
		}
		if strings.Contains(err.Error(), "invalid google token") {
			utils.RespondErrorWithErrorMessage(c, 401, constants.ErrCodeUnauthorized, "Invalid Google token", "invalidGoogleToken")
			return
		}
		if strings.Contains(err.Error(), "password must be at least 6") {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "Google sign-in failed", "googleLoginFailed")
		return
	}
	utils.SetAuthCookie(c, response.AccessToken, h.cookieConfig)
	utils.RespondSuccess[any](c, nil, "Login successful")
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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	ctx := c.Request.Context()
	success, err := h.services.Auth.VerifyOtp(ctx, req.Email, req.Otp)

	if err != nil {
		errMsg := err.Error()
		if constants.ErrCodeNotFound == errMsg {
			utils.RespondErrorWithErrorMessage(c, 404, constants.ErrCodeNotFound, errMsg, "userNotFound")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, errMsg, "verifyOtpFailed")
		return
	}

	if !success {
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, "Invalid or expired OTP", "invalidOrExpiredOtp")
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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	fromEmail := getEnvOr("SES_EMAIL_IDENTITY", "")
	frontendURL := getEnvOr("FRONTEND_URL", "")

	if err := h.services.Auth.ForgotPassword(c.Request.Context(), req.Email, h.services.Mail, frontendURL, fromEmail); err != nil {
		utils.RespondErrorWithErrorMessage(c, 500, constants.ErrCodeInternalServerError, "Failed to process password reset request", "forgotPasswordFailed")
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
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	if err := h.services.Auth.ResetPasswordWithToken(req.Token, &req); err != nil {
		if errors.Is(err, constants.ErrPasswordMismatch) {
			utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "passwordsDoNotMatch")
			return
		}
		if errors.Is(err, constants.ErrUserNotFound) {
			utils.RespondErrorWithErrorMessage(c, 404, constants.ErrCodeNotFound, err.Error(), "userNotFound")
			return
		}
		utils.RespondErrorWithErrorMessage(c, 400, constants.ErrCodeBadRequest, err.Error(), "resetPasswordConfirmFailed")
		return
	}

	utils.RespondSuccess[any](c, nil, "Password has been reset successfully")
}
