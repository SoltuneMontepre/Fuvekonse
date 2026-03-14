package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/conbook/requests"
	"general-service/internal/repositories"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type ConbookHandler struct {
	services *services.Services
}

func NewConbookHandler(services *services.Services) *ConbookHandler {
	return &ConbookHandler{services: services}
}

// UploadConbook godoc
// @Summary Upload a new conbook
// @Description Create a new conbook entry. Users can have maximum 10 conbooks. Each conbook can be a single image or text file.
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateConbookRequest true "Conbook upload request"
// @Success 201 "Conbook uploaded successfully"
// @Failure 400 "Invalid request"
// @Failure 401 "Unauthorized"
// @Failure 409 "Maximum conbook uploads (10) reached"
// @Failure 500 "Internal server error"
// @Router /conbooks/upload [post]
func (h *ConbookHandler) UploadConbook(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	var req requests.CreateConbookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	conbook, err := h.services.Conbook.UploadConbook(ctx, userID.(string), &req)
	if err != nil {
		if errors.Is(err, services.ErrConbookLimit) {
			utils.RespondError(c, 409, "CONBOOK_LIMIT_EXCEEDED", "Maximum conbook uploads (10) reached")
			return
		}
		utils.RespondInternalServerError(c, "Failed to upload conbook")
		return
	}

	utils.RespondCreated(c, conbook, "Conbook uploaded successfully")
}

// GetMyConbooks godoc
// @Summary Get current user's conbooks
// @Description Retrieve all conbooks uploaded by the authenticated user
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved conbooks"
// @Failure 401 "Unauthorized"
// @Failure 500 "Internal server error"
// @Router /conbooks [get]
func (h *ConbookHandler) GetMyConbooks(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	conbooks, err := h.services.Conbook.GetUserConbooks(ctx, userID.(string))
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve conbooks")
		return
	}

	utils.RespondSuccess(c, &conbooks, "Successfully retrieved conbooks")
}

// GetConbookByID godoc
// @Summary Get a specific conbook
// @Description Retrieve details of a specific conbook by ID
// @Tags conbooks
// @Accept json
// @Produce json
// @Param id path string true "Conbook ID" format(uuid)
// @Success 200 "Successfully retrieved conbook"
// @Failure 400 "Invalid conbook ID"
// @Failure 404 "Conbook not found"
// @Failure 500 "Internal server error"
// @Router /conbooks/{id} [get]
func (h *ConbookHandler) GetConbookByID(c *gin.Context) {
	ctx := c.Request.Context()
	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	conbook, err := h.services.Conbook.GetConbookByID(ctx, conbookID)
	if err != nil {
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Successfully retrieved conbook")
}

// EditConbook godoc
// @Summary Update a conbook
// @Description Update conbook details. Can only be edited while status is pending. User can only edit their own conbooks.
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Param request body requests.UpdateConbookRequest true "Update request"
// @Success 200 "Conbook updated successfully"
// @Failure 400 "Invalid request or conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Cannot edit non-pending conbook or not owner"
// @Failure 404 "Conbook not found"
// @Failure 500 "Internal server error"
// @Router /conbooks/{id} [put]
func (h *ConbookHandler) EditConbook(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	var req requests.UpdateConbookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	conbook, err := h.services.Conbook.EditConbook(ctx, userID.(string), conbookID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedConbook) {
			utils.RespondForbidden(c, "Cannot edit non-pending conbook or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrConbookNotEditable) {
			utils.RespondForbidden(c, "Cannot edit non-pending conbook")
			return
		}
		utils.RespondInternalServerError(c, "Failed to update conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook updated successfully")
}

// DeleteConbook godoc
// @Summary Delete a conbook
// @Description Delete a conbook. Can only be deleted while status is pending. User can only delete their own conbooks.
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 204 "Conbook deleted successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Cannot delete non-pending conbook or not owner"
// @Failure 404 "Conbook not found"
// @Failure 500 "Internal server error"
// @Router /conbooks/{id} [delete]
func (h *ConbookHandler) DeleteConbook(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	err := h.services.Conbook.DeleteConbook(ctx, userID.(string), conbookID)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedConbook) {
			utils.RespondForbidden(c, "Cannot delete non-pending conbook or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrConbookNotEditable) {
			utils.RespondForbidden(c, "Cannot delete non-pending conbook")
			return
		}
		utils.RespondInternalServerError(c, "Failed to delete conbook")
		return
	}

	c.JSON(204, nil)
}

// GetPendingConbooks godoc
// @Summary Get pending conbooks for review
// @Description Retrieve all pending conbooks for staff review (staff only)
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved pending conbooks"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/pending [get]
func (h *ConbookHandler) GetPendingConbooks(c *gin.Context) {
	ctx := c.Request.Context()
	conbooks, err := h.services.Conbook.GetPendingConbooks(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve pending conbooks")
		return
	}

	utils.RespondSuccess(c, &conbooks, "Successfully retrieved pending conbooks")
}

// GetApprovedConbooks godoc
// @Summary Get approved conbooks
// @Description Retrieve all approved conbooks (admin/staff only)
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved approved conbooks"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/approved [get]
func (h *ConbookHandler) GetApprovedConbooks(c *gin.Context) {
	ctx := c.Request.Context()
	conbooks, err := h.services.Conbook.GetApprovedConbooks(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve approved conbooks")
		return
	}

	utils.RespondSuccess(c, &conbooks, "Successfully retrieved approved conbooks")
}

// GetDeniedConbooks godoc
// @Summary Get denied conbooks
// @Description Retrieve all denied conbooks (admin/staff only)
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Successfully retrieved denied conbooks"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/denied [get]
func (h *ConbookHandler) GetDeniedConbooks(c *gin.Context) {
	ctx := c.Request.Context()
	conbooks, err := h.services.Conbook.GetDeniedConbooks(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve denied conbooks")
		return
	}

	utils.RespondSuccess(c, &conbooks, "Successfully retrieved denied conbooks")
}

// ApproveConbook godoc
// @Summary Approve a conbook
// @Description Mark a conbook as approved by staff.
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 200 "Conbook approved successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 404 "Conbook not found"
// @Failure 409 "Conbook already has approved status"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/{id}/approve [patch]
func (h *ConbookHandler) ApproveConbook(c *gin.Context) {
	ctx := c.Request.Context()
	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	conbook, err := h.services.Conbook.ApproveConbook(ctx, conbookID)
	if err != nil {
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Conbook already has approved status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to approve conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook approved successfully")
}

// DenyConbook godoc
// @Summary Deny a conbook
// @Description Mark a conbook as denied by staff/admin.
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 200 "Conbook denied successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 404 "Conbook not found"
// @Failure 409 "Conbook already has denied status"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/{id}/deny [patch]
func (h *ConbookHandler) DenyConbook(c *gin.Context) {
	ctx := c.Request.Context()
	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	conbook, err := h.services.Conbook.DenyConbook(ctx, conbookID)
	if err != nil {
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Conbook already has denied status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to deny conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook denied successfully")
}

// MarkConbookPending godoc
// @Summary Mark conbook as pending
// @Description Move a conbook back to pending status by staff/admin.
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 200 "Conbook status set to pending successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 404 "Conbook not found"
// @Failure 409 "Conbook already has pending status"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/{id}/pending [patch]
func (h *ConbookHandler) MarkConbookPending(c *gin.Context) {
	ctx := c.Request.Context()
	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	conbook, err := h.services.Conbook.MarkConbookPending(ctx, conbookID)
	if err != nil {
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Conbook already has pending status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to set conbook status to pending")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook status set to pending successfully")
}
