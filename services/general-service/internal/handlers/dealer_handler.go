package handlers

import (
	"general-service/internal/common/utils"
	"general-service/internal/dto/dealer/requests"
	"general-service/internal/services"
	"strconv"

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
// @Description Requires authentication and an approved ticket. User will become the owner of the dealer booth.
// @Description
// @Description **Requirements:**
// @Description - User must have a valid, approved ticket
// @Description - User cannot already be a dealer staff
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
// @Failure 403 "Forbidden - user does not have an approved ticket"
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
		case "user must have a ticket to register as a dealer":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "user ticket must be approved to register as a dealer":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "failed to register dealer":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to register dealer")
		}
		return
	}

	utils.RespondCreated(c, booth, "Successfully registered as dealer")
}

// GetDealersForAdmin godoc
// @Summary Get all dealer booths (admin)
// @Description Get a paginated list of all dealer booths with optional filters
// @Description Admins can view all dealer booths and filter by verification status
// @Tags admin-dealers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20) maximum(100)
// @Param is_verified query bool false "Filter by verification status"
// @Success 200 "Successfully retrieved dealer booths"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - admin only"
// @Failure 500 "Internal server error"
// @Router /admin/dealers [get]
func (h *DealerHandler) GetDealersForAdmin(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	// Parse filter parameters
	var isVerified *bool

	if isVerifiedStr := c.Query("is_verified"); isVerifiedStr != "" {
		val := isVerifiedStr == "true"
		isVerified = &val
	}

	// Call service
	booths, meta, err := h.services.Dealer.GetAllDealersForAdmin(page, pageSize, isVerified)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve dealer booths")
		return
	}

	utils.RespondSuccessWithMeta(c, &booths, meta, "Successfully retrieved dealer booths")
}

// GetDealerByIDForAdmin godoc
// @Summary Get dealer booth by ID (admin)
// @Description Get detailed information about a specific dealer booth including staff
// @Tags admin-dealers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dealer Booth ID" format(uuid)
// @Success 200 "Successfully retrieved dealer booth"
// @Failure 400 "Invalid dealer booth ID"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - admin only"
// @Failure 404 "Dealer booth not found"
// @Failure 500 "Internal server error"
// @Router /admin/dealers/{id} [get]
func (h *DealerHandler) GetDealerByIDForAdmin(c *gin.Context) {
	boothID := c.Param("id")
	if boothID == "" {
		utils.RespondBadRequest(c, "Dealer booth ID is required")
		return
	}

	// Call service
	booth, err := h.services.Dealer.GetDealerByIDForAdmin(boothID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "dealer booth not found":
			utils.RespondNotFound(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to retrieve dealer booth")
		}
		return
	}

	utils.RespondSuccess(c, booth, "Successfully retrieved dealer booth")
}

// VerifyDealer godoc
// @Summary Verify dealer booth (admin)
// @Description Verify a dealer booth and generate a unique 6-character alphanumeric booth code
// @Description The booth code is automatically generated and consists of uppercase letters and numbers
// @Tags admin-dealers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dealer Booth ID" format(uuid)
// @Success 200 "Successfully verified dealer booth"
// @Failure 400 "Invalid dealer booth ID"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - admin only"
// @Failure 404 "Dealer booth not found"
// @Failure 409 "Conflict - booth already verified"
// @Failure 500 "Internal server error"
// @Router /admin/dealers/{id}/verify [patch]
func (h *DealerHandler) VerifyDealer(c *gin.Context) {
	boothID := c.Param("id")
	if boothID == "" {
		utils.RespondBadRequest(c, "Dealer booth ID is required")
		return
	}

	// Call service to verify dealer
	booth, err := h.services.Dealer.VerifyDealer(boothID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "dealer booth not found":
			utils.RespondNotFound(c, errMsg)
		case "dealer booth is already verified":
			utils.RespondError(c, 409, "CONFLICT", errMsg)
		case "failed to generate booth code",
			"failed to check booth code availability",
			"failed to generate unique booth code after multiple attempts",
			"failed to verify dealer booth":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to verify dealer booth")
		}
		return
	}

	utils.RespondSuccess(c, booth, "Successfully verified dealer booth")
}

// JoinDealerBooth godoc
// @Summary Join a dealer booth
// @Description Join a dealer booth as a staff member using a booth code
// @Description Requires authentication and an approved ticket. User cannot be a staff member of another booth.
// @Description
// @Description **Requirements:**
// @Description - User must have a valid, approved ticket
// @Description - User cannot already be a dealer staff member
// @Description - Booth must be verified
// @Description
// @Description **Usage:**
// @Description 1. Include JWT access token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Description 2. Provide the 6-character booth code
// @Description 3. Receive the booth information with all staff members
// @Tags dealer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.DealerJoinRequest true "Dealer join request"
// @Success 200 "Successfully joined dealer booth"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - user does not have an approved ticket or booth not verified"
// @Failure 404 "Dealer booth not found with provided code"
// @Failure 409 "Conflict - user is already a dealer staff"
// @Failure 500 "Internal server error"
// @Router /dealer/join [post]
func (h *DealerHandler) JoinDealerBooth(c *gin.Context) {
	var req requests.DealerJoinRequest

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

	// Call service to join dealer booth
	booth, err := h.services.Dealer.JoinDealerBooth(userID, req.BoothCode)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "dealer booth not found with provided code":
			utils.RespondNotFound(c, errMsg)
		case "user is already a staff member of a dealer booth":
			utils.RespondError(c, 409, "CONFLICT", errMsg)
		case "user must have a ticket to join a dealer booth":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "user ticket must be approved to join a dealer booth":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "dealer booth must be verified before accepting staff":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "failed to join dealer booth":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to join dealer booth")
		}
		return
	}

	utils.RespondSuccess(c, booth, "Successfully joined dealer booth")
}

// RemoveStaffFromBooth godoc
// @Summary Remove staff from booth (owner only)
// @Description Remove a staff member from the dealer booth. Only booth owners can remove staff members.
// @Description Owners cannot remove themselves or other owners.
// @Description
// @Description **Requirements:**
// @Description - User must be the booth owner
// @Description - Cannot remove yourself
// @Description - Cannot remove other owners
// @Description
// @Description **Usage:**
// @Description 1. Include JWT access token in Authorization header: Bearer YOUR_ACCESS_TOKEN
// @Description 2. Provide the user ID of the staff member to remove
// @Description 3. Receive the updated booth information
// @Tags dealer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.DealerRemoveStaffRequest true "Remove staff request"
// @Success 200 "Successfully removed staff member"
// @Failure 400 "Bad request - validation error"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 403 "Forbidden - only owners can remove staff or trying to remove owner"
// @Failure 404 "Staff member not found in booth"
// @Failure 500 "Internal server error"
// @Router /dealer/staff/remove [delete]
func (h *DealerHandler) RemoveStaffFromBooth(c *gin.Context) {
	var req requests.DealerRemoveStaffRequest

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

	// Call service to remove staff member
	booth, err := h.services.Dealer.RemoveStaffFromBooth(userID, req.StaffUserId)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			utils.RespondNotFound(c, errMsg)
		case "you are not a staff member of any dealer booth":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "you are not a staff member of this booth":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "only booth owners can remove staff members":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "booth owner cannot remove themselves":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "cannot remove booth owner":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "staff member not found in this booth":
			utils.RespondNotFound(c, errMsg)
		case "failed to remove staff member":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to remove staff member")
		}
		return
	}

	utils.RespondSuccess(c, booth, "Successfully removed staff member")
}

// GetMyDealer godoc
// @Summary Get current user's dealer booth
// @Description Get the dealer booth information for the currently authenticated user
// @Description Returns detailed booth information including staff members
// @Tags dealer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved user's dealer booth"
// @Failure 401 "Unauthorized - missing or invalid token"
// @Failure 404 "User is not a staff member of any dealer booth"
// @Failure 500 "Internal server error"
// @Router /dealer/me [get]
func (h *DealerHandler) GetMyDealer(c *gin.Context) {
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

	// Call service to get user's dealer booth
	booth, err := h.services.Dealer.GetMyDealer(userID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user is not a staff member of any dealer booth":
			utils.RespondNotFound(c, errMsg)
		case "failed to retrieve dealer booth information":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to retrieve dealer booth")
		}
		return
	}

	utils.RespondSuccess(c, booth, "Successfully retrieved dealer booth")
}
