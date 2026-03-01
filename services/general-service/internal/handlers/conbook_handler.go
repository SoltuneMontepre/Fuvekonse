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
// @Router /conbooks [post]
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
// @Description Update conbook details. Can only be edited before staff verification. User can only edit their own conbooks.
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Param request body requests.UpdateConbookRequest true "Update request"
// @Success 200 "Conbook updated successfully"
// @Failure 400 "Invalid request or conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Cannot edit verified conbook or not owner"
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
			utils.RespondForbidden(c, "Cannot edit verified conbook or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrConbookVerified) {
			utils.RespondForbidden(c, "Cannot edit verified conbook")
			return
		}
		utils.RespondInternalServerError(c, "Failed to update conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook updated successfully")
}

// DeleteConbook godoc
// @Summary Delete a conbook
// @Description Delete a conbook. Can only be deleted before staff verification. User can only delete their own conbooks.
// @Tags conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 204 "Conbook deleted successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Cannot delete verified conbook or not owner"
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
			utils.RespondForbidden(c, "Cannot delete verified conbook or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if errors.Is(err, services.ErrConbookVerified) {
			utils.RespondForbidden(c, "Cannot delete verified conbook")
			return
		}
		utils.RespondInternalServerError(c, "Failed to delete conbook")
		return
	}

	c.JSON(204, nil)
}

// GetPendingConbooks godoc
// @Summary Get pending conbooks for review
// @Description Retrieve all unverified conbooks for staff review (staff only)
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
	conbooks, err := h.services.Conbook.GetUnverifiedConbooks(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve pending conbooks")
		return
	}

	utils.RespondSuccess(c, &conbooks, "Successfully retrieved pending conbooks")
}

// VerifyConbook godoc
// @Summary Verify a conbook
// @Description Mark a conbook as verified by staff. After verification, users cannot edit the conbook.
// @Tags admin-conbooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conbook ID" format(uuid)
// @Success 200 "Conbook verified successfully"
// @Failure 400 "Invalid conbook ID"
// @Failure 401 "Unauthorized"
// @Failure 403 "Insufficient permissions"
// @Failure 404 "Conbook not found"
// @Failure 409 "Conbook already verified"
// @Failure 500 "Internal server error"
// @Router /admin/conbooks/{id}/verify [patch]
func (h *ConbookHandler) VerifyConbook(c *gin.Context) {
	ctx := c.Request.Context()
	conbookID := c.Param("id")
	if conbookID == "" {
		utils.RespondValidationError(c, "Conbook ID is required")
		return
	}

	conbook, err := h.services.Conbook.VerifyConbook(ctx, conbookID)
	if err != nil {
		if errors.Is(err, repositories.ErrConbookNotFound) {
			utils.RespondNotFound(c, "Conbook not found")
			return
		}
		if err.Error() == "conbook is already verified" {
			utils.RespondError(c, 409, "ALREADY_VERIFIED", "Conbook is already verified")
			return
		}
		utils.RespondInternalServerError(c, "Failed to verify conbook")
		return
	}

	utils.RespondSuccess(c, conbook, "Conbook verified successfully")
}
