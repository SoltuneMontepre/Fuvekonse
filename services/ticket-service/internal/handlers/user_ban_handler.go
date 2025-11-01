package handlers

import (
	"net/http"
	"strconv"
	"ticket-service/internal/dto"
	"ticket-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type UserBanHandler struct {
	userBanService services.UserBanServiceInterface
}

func NewUserBanHandler(userBanService services.UserBanServiceInterface) *UserBanHandler {
	return &UserBanHandler{
		userBanService: userBanService,
	}
}

func (h *UserBanHandler) BanUser(c *gin.Context) {
	userID := c.Param("user_id")
	
	var req dto.BanUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userBan, err := h.userBanService.BanUser(userID, req.PermissionID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.UserBanResponse{
		ID:         userBan.ID,
		UserID:     userBan.UserID,
		PermID:     userBan.PermID,
		Reason:     userBan.Reason,
		CreatedAt:  userBan.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  userBan.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, gin.H{"user_ban": response})
}

func (h *UserBanHandler) UnbanUser(c *gin.Context) {
	userID := c.Param("user_id")
	permissionIDStr := c.Param("permission_id")
	
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid permission ID"})
		return
	}

	if err := h.userBanService.UnbanUser(userID, uint(permissionID)); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "User unbanned successfully"})
}

func (h *UserBanHandler) GetUserBan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ban ID"})
		return
	}

	userBan, err := h.userBanService.GetUserBan(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_ban": userBan})
}

func (h *UserBanHandler) GetUserBans(c *gin.Context) {
	userID := c.Param("user_id")

	userBans, err := h.userBanService.GetUserBans(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_bans": userBans})
}

func (h *UserBanHandler) GetAllUserBans(c *gin.Context) {
	userBans, err := h.userBanService.GetAllUserBans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_bans": userBans})
}

func (h *UserBanHandler) CheckUserBan(c *gin.Context) {
	userID := c.Param("user_id")
	permissionIDStr := c.Query("permission_id")
	
	if permissionIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "permission_id query parameter is required"})
		return
	}

	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid permission ID"})
		return
	}

	isBanned, err := h.userBanService.IsUserBanned(userID, uint(permissionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.CheckUserBanResponse{
		UserID:       userID,
		PermissionID: uint(permissionID),
		IsBanned:     isBanned,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserBanHandler) UpdateBanReason(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ban ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userBanService.UpdateBanReason(uint(id), req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ban reason updated successfully"})
}