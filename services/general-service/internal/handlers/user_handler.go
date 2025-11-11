package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/user/requests"
	"general-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
// @Description Returns detailed user information including email and identification details
// @Description
// @Description **Usage:**
// @Description 1. Include JWT access token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Description 2. Receive your user profile information with sensitive fields
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
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	// Safe type assertion to prevent panic
	userID, ok := userIDRaw.(string)
	if !ok {
		utils.RespondUnauthorized(c, "Invalid user ID in token")
		return
	}

	// Call service to get detailed user information (including sensitive PII)
	// This is appropriate here since users are accessing their own data
	user, err := h.services.User.GetUserDetailedByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondNotFound(c, "User not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve user information")
		return
	}

	utils.RespondSuccess(c, user, "Successfully retrieved user information")
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update the profile information of the currently authenticated user
// @Description Allows partial updates of profile fields (fursona_name, first_name, last_name, country, identification_id, passport_id)
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateProfileRequest true "Profile update request"
// @Success 200 "Successfully updated user profile"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req requests.UpdateProfileRequest

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

	// Call service to update profile
	user, err := h.services.User.UpdateProfile(userID, &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "failed to update profile":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to update profile")
		}
		return
	}

	utils.RespondSuccess(c, user, "Profile updated successfully")
}

// UpdateAvatar godoc
// @Summary Update current user avatar
// @Description Update the avatar URL of the currently authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateAvatarRequest true "Avatar update request"
// @Success 200 "Successfully updated user avatar"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /users/me/avatar [patch]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	var req requests.UpdateAvatarRequest

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

	// Call service to update avatar
	user, err := h.services.User.UpdateAvatar(userID, &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "failed to update avatar":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to update avatar")
		}
		return
	}

	utils.RespondSuccess(c, user, "Avatar updated successfully")
}

// GetAllUsers godoc
// @Summary Get all users (admin only)
// @Description Get a paginated list of all users. Only accessible by admins.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Page size" default(10) minimum(1) maximum(100)
// @Success 200 "Successfully retrieved users list"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - insufficient permissions"
// @Failure 500 "Internal server error"
// @Router /admin/users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	// Call service
	users, meta, err := h.services.User.GetAllUsers(page, pageSize)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve users")
		return
	}

	// Convert slice to pointer for response
	usersPtr := &users
	utils.RespondSuccessWithMeta(c, usersPtr, meta, "Successfully retrieved users list")
}

// GetUserByIDForAdmin godoc
// @Summary Get user by ID (admin only)
// @Description Get detailed user information by ID. Admins can view deleted users.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Success 200 "Successfully retrieved user information"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - insufficient permissions"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserByIDForAdmin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	// Call service
	user, err := h.services.User.GetUserByIDForAdmin(userID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to retrieve user information")
		}
		return
	}

	utils.RespondSuccess(c, user, "Successfully retrieved user information")
}

// UpdateUserByAdmin godoc
// @Summary Update user by admin
// @Description Update user information. Admins can update any field including role and verification status.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Param request body requests.AdminUpdateUserRequest true "User update request"
// @Success 200 "Successfully updated user"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - insufficient permissions"
// @Failure 404 "User not found"
// @Failure 409 "Conflict - email already exists"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id} [put]
func (h *UserHandler) UpdateUserByAdmin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	var req requests.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Call service
	user, err := h.services.User.UpdateUserByAdmin(userID, &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "email already exists":
			utils.RespondError(c, 409, "CONFLICT", errMsg)
		case "failed to update user":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to update user")
		}
		return
	}

	utils.RespondSuccess(c, user, "User updated successfully")
}

// DeleteUser godoc
// @Summary Delete user (admin only)
// @Description Soft delete a user account. Only accessible by admins.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Success 200 "Successfully deleted user"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - insufficient permissions"
// @Failure 404 "User not found"
// @Failure 409 "Conflict - user already deleted"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	// Call service
	err := h.services.User.DeleteUser(userID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "user already deleted":
			utils.RespondError(c, 409, "CONFLICT", errMsg)
		case "failed to delete user":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to delete user")
		}
		return
	}

	utils.RespondSuccess[any](c, nil, "User deleted successfully")
}

// VerifyUser godoc
// @Summary Verify user account (admin only)
// @Description Verify a user account. Only accessible by admins.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Success 200 "Successfully verified user"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - insufficient permissions"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /admin/users/{id}/verify [patch]
func (h *UserHandler) VerifyUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.RespondBadRequest(c, "User ID is required")
		return
	}

	// Call service
	user, err := h.services.User.VerifyUser(userID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "failed to verify user":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to verify user")
		}
		return
	}

	utils.RespondSuccess(c, user, "User verified successfully")
}
